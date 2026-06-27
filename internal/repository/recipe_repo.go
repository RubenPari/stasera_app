package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/model"
)

// RecipeRepository manages persistence for generated recipes.
type RecipeRepository struct {
	db DBTX
}

// NewRecipeRepository returns a new RecipeRepository backed by the provided pool.
func NewRecipeRepository(db *sql.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

// WithTx returns a copy of the repository bound to the provided transaction.
func (r *RecipeRepository) WithTx(tx DBTX) *RecipeRepository {
	return &RecipeRepository{db: tx}
}

// recipeColumns is the canonical SELECT column list for recipe rows.
const recipeColumns = "id, user_id, name, prep_minutes, servings, ingredients, steps, is_rescue, times_cooked, last_cooked_at, created_at"

// Create inserts a new recipe and returns the created record.
func (r *RecipeRepository) Create(ctx context.Context, userID uuid.UUID, name string, prepMinutes, servings int, ingredients []model.RecipeIngredient, steps []model.RecipeStep, isRescue bool) (model.Recipe, error) {
	ingBytes, err := json.Marshal(ingredients)
	if err != nil {
		return model.Recipe{}, err
	}
	stepsBytes, err := json.Marshal(steps)
	if err != nil {
		return model.Recipe{}, err
	}

	id := uuid.NewString()
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO recipes (id, user_id, name, prep_minutes, servings, ingredients, steps, is_rescue)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, id, userID.String(), name, prepMinutes, servings, string(ingBytes), string(stepsBytes), isRescue)
	if err != nil {
		return model.Recipe{}, err
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return model.Recipe{}, err
	}
	return model.Recipe{
		ID:          uid,
		UserID:      userID,
		Name:        name,
		PrepMinutes: prepMinutes,
		Servings:    servings,
		Ingredients: ingredients,
		Steps:       steps,
		IsRescue:    isRescue,
		CreatedAt:   time.Now(),
	}, nil
}

// GetByID looks up a recipe by ID.
// Returns nil, nil when not found.
func (r *RecipeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Recipe, error) {
	row := r.db.QueryRowContext(ctx, "SELECT "+recipeColumns+" FROM recipes WHERE id = ?", id.String())
	recipe, err := scanRecipe(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return recipe, nil
}

// GetByIDForUser looks up a recipe by ID scoped to the authenticated user.
// Returns nil, nil when not found or not owned by the user.
func (r *RecipeRepository) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*model.Recipe, error) {
	row := r.db.QueryRowContext(ctx, "SELECT "+recipeColumns+" FROM recipes WHERE id = ? AND user_id = ?", id.String(), userID.String())
	recipe, err := scanRecipe(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return recipe, nil
}

// GetByUserID returns all recipes owned by a user, optionally filtered by is_rescue.
func (r *RecipeRepository) GetByUserID(ctx context.Context, userID uuid.UUID, isRescue *bool) ([]model.Recipe, error) {
	query := "SELECT " + recipeColumns + " FROM recipes WHERE user_id = ?"
	args := []any{userID.String()}
	if isRescue != nil {
		query += " AND is_rescue = ?"
		args = append(args, *isRescue)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return Collect(rows, func(s Scanner) (model.Recipe, error) {
		rec, err := scanRecipe(s)
		if err != nil {
			return model.Recipe{}, err
		}
		return *rec, nil
	})
}

// FindRecentNames returns recipe names used in the last two weeks by the user.
func (r *RecipeRepository) FindRecentNames(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT r.name
		FROM recipes r
		JOIN meal_plan_days d ON d.recipe_id = r.id
		JOIN meal_plans p ON p.id = d.plan_id
		WHERE p.user_id = ?
		  AND p.week_start >= DATE_SUB(CURRENT_DATE, INTERVAL 14 DAY)
	`, userID.String())
	if err != nil {
		return nil, err
	}
	return Collect(rows, func(s Scanner) (string, error) {
		var name string
		return name, s.Scan(&name)
	})
}

// MarkCooked increments times_cooked and sets last_cooked_at to today.
func (r *RecipeRepository) MarkCooked(ctx context.Context, id uuid.UUID) (model.Recipe, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE recipes
		SET times_cooked = times_cooked + 1,
		    last_cooked_at = CURRENT_DATE
		WHERE id = ?
	`, id.String())
	if err != nil {
		return model.Recipe{}, err
	}
	recipe, err := r.GetByID(ctx, id)
	if err != nil {
		return model.Recipe{}, err
	}
	if recipe == nil {
		return model.Recipe{}, sql.ErrNoRows
	}
	return *recipe, nil
}

// MarkCookedForUser increments times_cooked for a recipe owned by userID.
// Returns nil, nil when the recipe does not exist or is not owned by the user.
func (r *RecipeRepository) MarkCookedForUser(ctx context.Context, id, userID uuid.UUID) (*model.Recipe, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE recipes
		SET times_cooked = times_cooked + 1,
		    last_cooked_at = CURRENT_DATE
		WHERE id = ? AND user_id = ?
	`, id.String(), userID.String())
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
	return r.GetByID(ctx, id)
}

// Delete removes a recipe if it is not referenced by any meal plan day.
func (r *RecipeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var refCount int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM meal_plan_days WHERE recipe_id = ?
	`, id.String()).Scan(&refCount); err != nil {
		return err
	}
	if refCount > 0 {
		return ErrRecipeInUse
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM recipes WHERE id = ?`, id.String())
	return err
}

// DeleteForUser removes a recipe owned by the user. Returns ErrRecipeInUse when
// it is referenced by a meal plan. Returns nil, false when not found/not owned.
func (r *RecipeRepository) DeleteForUser(ctx context.Context, id, userID uuid.UUID) (bool, error) {
	var refCount int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM meal_plan_days WHERE recipe_id = ?
	`, id.String()).Scan(&refCount); err != nil {
		return false, err
	}
	if refCount > 0 {
		return false, ErrRecipeInUse
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM recipes WHERE id = ? AND user_id = ?`, id.String(), userID.String())
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// GetByPlanID returns recipes used in a specific weekly plan.
func (r *RecipeRepository) GetByPlanID(ctx context.Context, planID uuid.UUID) ([]model.Recipe, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.`+recipeColumns+`
		FROM recipes r
		JOIN meal_plan_days d ON d.recipe_id = r.id
		WHERE d.plan_id = ?
		ORDER BY d.day_of_week
	`, planID.String())
	if err != nil {
		return nil, err
	}
	return Collect(rows, func(s Scanner) (model.Recipe, error) {
		rec, err := scanRecipe(s)
		if err != nil {
			return model.Recipe{}, err
		}
		return *rec, nil
	})
}

// GetIngredientsByPlanID aggregates ingredients from all recipes in a plan.
func (r *RecipeRepository) GetIngredientsByPlanID(ctx context.Context, planID uuid.UUID) ([]model.RecipeIngredient, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.ingredients
		FROM recipes r
		JOIN meal_plan_days d ON d.recipe_id = r.id
		WHERE d.plan_id = ?
	`, planID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []model.RecipeIngredient
	for rows.Next() {
		var ingBytes []byte
		if err := rows.Scan(&ingBytes); err != nil {
			return nil, err
		}
		var ings []model.RecipeIngredient
		if err := json.Unmarshal(ingBytes, &ings); err != nil {
			return nil, err
		}
		all = append(all, ings...)
	}
	return all, rows.Err()
}

// scanRecipe reads a recipe row from a Scanner (shared interface for *sql.Row/*sql.Rows).
func scanRecipe(s Scanner) (*model.Recipe, error) {
	var recipe model.Recipe
	var idStr, userIDStr string
	var ingBytes, stepsBytes []byte
	var lastCookedAt sql.NullTime

	if err := s.Scan(
		&idStr,
		&userIDStr,
		&recipe.Name,
		&recipe.PrepMinutes,
		&recipe.Servings,
		&ingBytes,
		&stepsBytes,
		&recipe.IsRescue,
		&recipe.TimesCooked,
		&lastCookedAt,
		&recipe.CreatedAt,
	); err != nil {
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
	recipe.ID = parsedID
	recipe.UserID = parsedUserID

	if err := json.Unmarshal(ingBytes, &recipe.Ingredients); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(stepsBytes, &recipe.Steps); err != nil {
		return nil, err
	}
	if lastCookedAt.Valid {
		t := lastCookedAt.Time
		recipe.LastCookedAt = &t
	}
	return &recipe, nil
}