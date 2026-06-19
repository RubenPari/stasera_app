package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/model"
)

// ShoppingListRepository manages persistence for shopping lists and items.
type ShoppingListRepository struct {
	db *sql.DB
}

// NewShoppingListRepository returns a new ShoppingListRepository backed by the provided pool.
func NewShoppingListRepository(db *sql.DB) *ShoppingListRepository {
	return &ShoppingListRepository{db: db}
}

// GetCurrent returns the user's most recent open shopping list, if any.
// Open means it has not been marked completed yet.
func (r *ShoppingListRepository) GetCurrent(ctx context.Context, userID uuid.UUID) (*model.ShoppingList, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, plan_id, created_at, completed_at
		FROM shopping_lists
		WHERE user_id = ? AND completed_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, userID.String())
	list, err := scanShoppingList(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteOpenByUserID removes every open (not completed) shopping list for the user.
// Used before generating a new list so the previous one is replaced.
func (r *ShoppingListRepository) DeleteOpenByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM shopping_lists
		WHERE user_id = ? AND completed_at IS NULL
	`, userID.String())
	return err
}

// CreateWithItems inserts a new shopping list together with its items in a single transaction.
// sort_order is taken from each item as provided. Returns the created list (without items).
func (r *ShoppingListRepository) CreateWithItems(ctx context.Context, userID uuid.UUID, planID *uuid.UUID, items []model.ShoppingItem) (model.ShoppingList, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.ShoppingList{}, err
	}
	defer tx.Rollback()

	id := uuid.NewString()
	var planIDArg interface{}
	if planID != nil {
		planIDArg = planID.String()
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO shopping_lists (id, user_id, plan_id)
		VALUES (?, ?, ?)
	`, id, userID.String(), planIDArg)
	if err != nil {
		return model.ShoppingList{}, err
	}

	for _, it := range items {
		itemID := uuid.NewString()
		_, err := tx.ExecContext(ctx, `
			INSERT INTO shopping_items (id, list_id, name, quantity, aisle, is_checked, sort_order)
			VALUES (?, ?, ?, ?, ?, FALSE, ?)
		`, itemID, id, it.Name, it.Quantity, it.Aisle, it.SortOrder)
		if err != nil {
			return model.ShoppingList{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.ShoppingList{}, err
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, plan_id, created_at, completed_at
		FROM shopping_lists
		WHERE id = ?
	`, id)
	list, err := scanShoppingList(row)
	if err != nil {
		return model.ShoppingList{}, err
	}
	return *list, nil
}

// GetItemsByListID returns all items of a list ordered by sort_order.
func (r *ShoppingListRepository) GetItemsByListID(ctx context.Context, listID uuid.UUID) ([]model.ShoppingItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, list_id, name, quantity, aisle, is_checked, sort_order
		FROM shopping_items
		WHERE list_id = ?
		ORDER BY sort_order, name
	`, listID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.ShoppingItem
	for rows.Next() {
		var it model.ShoppingItem
		var idStr, listIDStr string
		if err := rows.Scan(&idStr, &listIDStr, &it.Name, &it.Quantity, &it.Aisle, &it.IsChecked, &it.SortOrder); err != nil {
			return nil, err
		}
		it.ID = uuid.MustParse(idStr)
		it.ListID = uuid.MustParse(listIDStr)
		items = append(items, it)
	}
	return items, rows.Err()
}

// UpdateItemChecked toggles the checked state of an item owned (transitively) by the user.
// Returns nil, nil when the item does not exist or does not belong to the user.
func (r *ShoppingListRepository) UpdateItemChecked(ctx context.Context, userID, itemID uuid.UUID, isChecked bool) (*model.ShoppingItem, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE shopping_items i
		INNER JOIN shopping_lists l ON i.list_id = l.id
		SET i.is_checked = ?
		WHERE i.id = ? AND l.user_id = ?
	`, isChecked, itemID.String(), userID.String())
	if err != nil {
		return nil, err
	}

	var it model.ShoppingItem
	var idStr, listIDStr string
	err = r.db.QueryRowContext(ctx, `
		SELECT id, list_id, name, quantity, aisle, is_checked, sort_order
		FROM shopping_items
		WHERE id = ?
	`, itemID.String()).Scan(&idStr, &listIDStr, &it.Name, &it.Quantity, &it.Aisle, &it.IsChecked, &it.SortOrder)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	it.ID = uuid.MustParse(idStr)
	it.ListID = uuid.MustParse(listIDStr)
	return &it, nil
}

// MarkCompleted sets completed_at on the user's open shopping list.
// Returns nil, nil when no open list exists for the user.
func (r *ShoppingListRepository) MarkCompleted(ctx context.Context, userID uuid.UUID) (*model.ShoppingList, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE shopping_lists
		SET completed_at = NOW()
		WHERE user_id = ? AND completed_at IS NULL
	`, userID.String())
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, plan_id, created_at, completed_at
		FROM shopping_lists
		WHERE user_id = ? AND completed_at IS NOT NULL
		ORDER BY completed_at DESC
		LIMIT 1
	`, userID.String())
	list, err := scanShoppingList(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return list, nil
}

// shoppingListScanner is the common interface accepted by scanShoppingList.
type shoppingListScanner interface {
	Scan(dest ...interface{}) error
}

func scanShoppingList(s shoppingListScanner) (*model.ShoppingList, error) {
	var list model.ShoppingList
	var idStr, userIDStr string
	var planIDStr sql.NullString
	var completedAt sql.NullTime
	if err := s.Scan(&idStr, &userIDStr, &planIDStr, &list.CreatedAt, &completedAt); err != nil {
		return nil, err
	}
	list.ID = uuid.MustParse(idStr)
	list.UserID = uuid.MustParse(userIDStr)
	if planIDStr.Valid {
		p := uuid.MustParse(planIDStr.String)
		list.PlanID = &p
	}
	if completedAt.Valid {
		list.CompletedAt = &completedAt.Time
	}
	return &list, nil
}