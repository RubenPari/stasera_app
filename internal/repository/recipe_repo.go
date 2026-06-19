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
	db *sql.DB
}

// NewRecipeRepository returns a new RecipeRepository backed by the provided pool.
func NewRecipeRepository(db *sql.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

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
	recipe, err := r.GetByID(ctx, uuid.MustParse(id))
	if err != nil {
		return model.Recipe{}, err
	}
	return *recipe, nil
}

// GetByID looks up a recipe by ID.
// Returns nil, nil when not found.
func (r *RecipeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Recipe, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, prep_minutes, servings, ingredients, steps, is_rescue, times_cooked, last_cooked_at, created_at
		FROM recipes
		WHERE id = ?
	`, id.String())
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
	query := `
		SELECT id, user_id, name, prep_minutes, servings, ingredients, steps, is_rescue, times_cooked, last_cooked_at, created_at
		FROM recipes
		WHERE user_id = ?
	`
	args := []interface{}{userID.String()}
	if isRescue != nil {
		query += ` AND is_rescue = ?`
		args = append(args, *isRescue)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []model.Recipe
	for rows.Next() {
		r, err := scanRecipe(rows)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, *r)
	}
	return recipes, rows.Err()
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
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
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

// ErrRecipeInUse is returned when trying to delete a recipe referenced by a plan.
var ErrRecipeInUse = errors.New("recipe is used in a meal plan")

// GetByPlanID returns recipes used in a specific weekly plan.
func (r *RecipeRepository) GetByPlanID(ctx context.Context, planID uuid.UUID) ([]model.Recipe, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.user_id, r.name, r.prep_minutes, r.servings, r.ingredients, r.steps, r.is_rescue, r.times_cooked, r.last_cooked_at, r.created_at
		FROM recipes r
		JOIN meal_plan_days d ON d.recipe_id = r.id
		WHERE d.plan_id = ?
		ORDER BY d.day_of_week
	`, planID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []model.Recipe
	for rows.Next() {
		r, err := scanRecipe(rows)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, *r)
	}
	return recipes, rows.Err()
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

// UsedByPlanID reports whether the recipe is referenced by a meal plan.
func (r *RecipeRepository) UsedByPlanID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM meal_plan_days WHERE recipe_id = ?
	`, id.String()).Scan(&count)
	return count > 0, err
}

// recipeScanner is the common interface accepted by scanRecipe.
type recipeScanner interface {
	Scan(dest ...interface{}) error
}

// scanRecipe reads a recipe row from a scanner (QueryRow or Rows).
func scanRecipe(s recipeScanner) (*model.Recipe, error) {
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

// ptrDate keeps compatibility with older call sites.
func ptrDate(t time.Time) *time.Time {
	return &t
}