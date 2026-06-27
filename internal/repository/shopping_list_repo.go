package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/model"
)

// ShoppingListRepository manages persistence for shopping lists and items.
type ShoppingListRepository struct {
	db DBTX
}

// NewShoppingListRepository returns a new ShoppingListRepository backed by the provided pool.
func NewShoppingListRepository(db *sql.DB) *ShoppingListRepository {
	return &ShoppingListRepository{db: db}
}

// WithTx returns a copy of the repository bound to the provided transaction.
func (r *ShoppingListRepository) WithTx(tx DBTX) *ShoppingListRepository {
	return &ShoppingListRepository{db: tx}
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

// CreateWithItems inserts a new shopping list together with its items.
// sort_order is taken from each item as provided. The list+items are written
// as a single atomic unit when the caller wraps this in a transaction (see
// ShoppingListService.Generate, which begins the tx and calls WithTx).
func (r *ShoppingListRepository) CreateWithItems(ctx context.Context, userID uuid.UUID, planID *uuid.UUID, items []model.ShoppingItem) (model.ShoppingList, error) {
	id := uuid.NewString()
	var planIDArg interface{}
	if planID != nil {
		planIDArg = planID.String()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO shopping_lists (id, user_id, plan_id)
		VALUES (?, ?, ?)
	`, id, userID.String(), planIDArg)
	if err != nil {
		return model.ShoppingList{}, err
	}

	for _, it := range items {
		itemID := uuid.NewString()
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO shopping_items (id, list_id, name, quantity, aisle, is_checked, sort_order)
			VALUES (?, ?, ?, ?, ?, FALSE, ?)
		`, itemID, id, it.Name, it.Quantity, it.Aisle, it.SortOrder)
		if err != nil {
			return model.ShoppingList{}, err
		}
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return model.ShoppingList{}, err
	}
	list := model.ShoppingList{
		ID:        uid,
		UserID:    userID,
		PlanID:    planID,
		CreatedAt: time.Now(),
	}
	return list, nil
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
	return Collect(rows, func(s Scanner) (model.ShoppingItem, error) {
		var it model.ShoppingItem
		var idStr, listIDStr string
		if err := s.Scan(&idStr, &listIDStr, &it.Name, &it.Quantity, &it.Aisle, &it.IsChecked, &it.SortOrder); err != nil {
			return model.ShoppingItem{}, err
		}
		var parseErr error
		it.ID, parseErr = uuid.Parse(idStr)
		if parseErr != nil {
			return model.ShoppingItem{}, parseErr
		}
		it.ListID, parseErr = uuid.Parse(listIDStr)
		if parseErr != nil {
			return model.ShoppingItem{}, parseErr
		}
		return it, nil
	})
}

// UpdateItemChecked toggles the checked state of an item owned (transitively) by the user.
// Returns nil, nil when the item does not exist or does not belong to the user:
// the UPDATE is filtered by user_id, so zero rows affected means not-owned/not-found,
// and the subsequent SELECT is only run when an owned row was actually updated,
// preventing cross-user data leakage.
func (r *ShoppingListRepository) UpdateItemChecked(ctx context.Context, userID, itemID uuid.UUID, isChecked bool) (*model.ShoppingItem, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE shopping_items i
		INNER JOIN shopping_lists l ON i.list_id = l.id
		SET i.is_checked = ?
		WHERE i.id = ? AND l.user_id = ?
	`, isChecked, itemID.String(), userID.String())
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}

	var it model.ShoppingItem
	var idStr, listIDStr string
	err = r.db.QueryRowContext(ctx, `
		SELECT id, list_id, name, quantity, aisle, is_checked, sort_order
		FROM shopping_items
		WHERE id = ?
	`, itemID.String()).Scan(&idStr, &listIDStr, &it.Name, &it.Quantity, &it.Aisle, &it.IsChecked, &it.SortOrder)
	if err != nil {
		return nil, err
	}
	it.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	it.ListID, err = uuid.Parse(listIDStr)
	if err != nil {
		return nil, err
	}
	return &it, nil
}
func (r *ShoppingListRepository) MarkCompleted(ctx context.Context, userID uuid.UUID) (*model.ShoppingList, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE shopping_lists
		SET completed_at = NOW()
		WHERE user_id = ? AND completed_at IS NULL
	`, userID.String())
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		// No open list existed; avoid returning a stale completed list.
		return nil, nil
	}

	// The just-completed list is now the most recent completed one for the user.
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, plan_id, created_at, completed_at
		FROM shopping_lists
		WHERE user_id = ? AND completed_at IS NOT NULL
		ORDER BY completed_at DESC
		LIMIT 1
	`, userID.String())
	list, err := scanShoppingList(row)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func scanShoppingList(s Scanner) (*model.ShoppingList, error) {
	var list model.ShoppingList
	var idStr, userIDStr string
	var planIDStr sql.NullString
	var completedAt sql.NullTime
	if err := s.Scan(&idStr, &userIDStr, &planIDStr, &list.CreatedAt, &completedAt); err != nil {
		return nil, err
	}
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	parsedUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}
	list.ID = parsedID
	list.UserID = parsedUserID
	if planIDStr.Valid {
		p, err := uuid.Parse(planIDStr.String)
		if err != nil {
			return nil, err
		}
		list.PlanID = &p
	}
	if completedAt.Valid {
		list.CompletedAt = &completedAt.Time
	}
	return &list, nil
}