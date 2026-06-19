package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/model"
)

// MealPlanRepository manages persistence for meal plans and their days.
type MealPlanRepository struct {
	db *sql.DB
}

// NewMealPlanRepository returns a new MealPlanRepository backed by the provided pool.
func NewMealPlanRepository(db *sql.DB) *MealPlanRepository {
	return &MealPlanRepository{db: db}
}

// Create inserts a new meal plan for the user starting on weekStart.
func (r *MealPlanRepository) Create(ctx context.Context, userID uuid.UUID, weekStart time.Time) (model.MealPlan, error) {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO meal_plans (id, user_id, week_start, status)
		VALUES (?, ?, ?, 'active')
	`, id, userID.String(), weekStart)
	if err != nil {
		return model.MealPlan{}, err
	}
	plan, err := r.GetByID(ctx, uuid.MustParse(id))
	if err != nil {
		return model.MealPlan{}, err
	}
	return *plan, nil
}

// GetCurrent returns the active meal plan for the current week, if any.
func (r *MealPlanRepository) GetCurrent(ctx context.Context, userID uuid.UUID) (*model.MealPlan, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, week_start, status, created_at
		FROM meal_plans
		WHERE user_id = ?
		  AND status = 'active'
		ORDER BY week_start DESC
		LIMIT 1
	`, userID.String())
	plan, err := scanMealPlan(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// GetByID returns a meal plan by primary key.
func (r *MealPlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MealPlan, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, week_start, status, created_at
		FROM meal_plans
		WHERE id = ?
	`, id.String())
	plan, err := scanMealPlan(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// AddDay inserts a day into a meal plan.
func (r *MealPlanRepository) AddDay(ctx context.Context, planID uuid.UUID, dayOfWeek int, recipeID uuid.UUID) (model.MealPlanDay, error) {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO meal_plan_days (id, plan_id, day_of_week, recipe_id)
		VALUES (?, ?, ?, ?)
	`, id, planID.String(), dayOfWeek, recipeID.String())
	if err != nil {
		return model.MealPlanDay{}, err
	}

	var d model.MealPlanDay
	var idStr, planIDStr, recipeIDStr string
	err = r.db.QueryRowContext(ctx, `
		SELECT id, plan_id, day_of_week, recipe_id
		FROM meal_plan_days
		WHERE id = ?
	`, id).Scan(&idStr, &planIDStr, &d.DayOfWeek, &recipeIDStr)
	if err != nil {
		return model.MealPlanDay{}, err
	}
	d.ID = uuid.MustParse(idStr)
	d.PlanID = uuid.MustParse(planIDStr)
	d.RecipeID = uuid.MustParse(recipeIDStr)
	return d, nil
}

// GetDays returns all days for a plan, ordered by day_of_week.
func (r *MealPlanRepository) GetDays(ctx context.Context, planID uuid.UUID) ([]model.MealPlanDay, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, plan_id, day_of_week, recipe_id
		FROM meal_plan_days
		WHERE plan_id = ?
		ORDER BY day_of_week
	`, planID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []model.MealPlanDay
	for rows.Next() {
		var d model.MealPlanDay
		var idStr, planIDStr, recipeIDStr string
		if err := rows.Scan(&idStr, &planIDStr, &d.DayOfWeek, &recipeIDStr); err != nil {
			return nil, err
		}
		d.ID = uuid.MustParse(idStr)
		d.PlanID = uuid.MustParse(planIDStr)
		d.RecipeID = uuid.MustParse(recipeIDStr)
		days = append(days, d)
	}
	return days, rows.Err()
}

// UpdateDayRecipe changes the recipe assigned to a specific day.
func (r *MealPlanRepository) UpdateDayRecipe(ctx context.Context, planID uuid.UUID, dayOfWeek int, recipeID uuid.UUID) (model.MealPlanDay, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE meal_plan_days
		SET recipe_id = ?
		WHERE plan_id = ? AND day_of_week = ?
	`, recipeID.String(), planID.String(), dayOfWeek)
	if err != nil {
		return model.MealPlanDay{}, err
	}

	var d model.MealPlanDay
	var idStr, planIDStr, recipeIDStr string
	err = r.db.QueryRowContext(ctx, `
		SELECT id, plan_id, day_of_week, recipe_id
		FROM meal_plan_days
		WHERE plan_id = ? AND day_of_week = ?
	`, planID.String(), dayOfWeek).Scan(&idStr, &planIDStr, &d.DayOfWeek, &recipeIDStr)
	if err != nil {
		return model.MealPlanDay{}, err
	}
	d.ID = uuid.MustParse(idStr)
	d.PlanID = uuid.MustParse(planIDStr)
	d.RecipeID = uuid.MustParse(recipeIDStr)
	return d, nil
}

// GetTodayRecipeID returns the recipe ID assigned to the current day of week in the active plan.
// Returns nil if today is not a weekday or no plan exists.
func (r *MealPlanRepository) GetTodayRecipeID(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	dayOfWeek := int(time.Now().Weekday())
	if dayOfWeek == 0 || dayOfWeek == 6 {
		return nil, nil
	}
	// Go: Sunday=0, Monday=1 ... Friday=5, Saturday=6. API uses 1=lun ... 5=ven.
	apiDayOfWeek := dayOfWeek

	var recipeIDStr string
	err := r.db.QueryRowContext(ctx, `
		SELECT d.recipe_id
		FROM meal_plan_days d
		JOIN meal_plans p ON p.id = d.plan_id
		WHERE p.user_id = ?
		  AND p.status = 'active'
		  AND d.day_of_week = ?
		ORDER BY p.week_start DESC
		LIMIT 1
	`, userID.String(), apiDayOfWeek).Scan(&recipeIDStr)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	recipeID := uuid.MustParse(recipeIDStr)
	return &recipeID, nil
}

// ArchiveOldPlans marks every active plan older than the given week_start as archived.
func (r *MealPlanRepository) ArchiveOldPlans(ctx context.Context, userID uuid.UUID, weekStart time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE meal_plans
		SET status = 'archived'
		WHERE user_id = ? AND week_start < ? AND status = 'active'
	`, userID.String(), weekStart)
	return err
}

// PlanExistsForWeek reports whether the user already has a plan for the given week start.
func (r *MealPlanRepository) PlanExistsForWeek(ctx context.Context, userID uuid.UUID, weekStart time.Time) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM meal_plans
			WHERE user_id = ? AND week_start = ?
		)
	`, userID.String(), weekStart).Scan(&exists)
	return exists, err
}

// mealPlanScanner is the common interface accepted by scanMealPlan.
type mealPlanScanner interface {
	Scan(dest ...interface{}) error
}

func scanMealPlan(s mealPlanScanner) (*model.MealPlan, error) {
	var plan model.MealPlan
	var idStr, userIDStr string
	if err := s.Scan(&idStr, &userIDStr, &plan.WeekStart, &plan.Status, &plan.CreatedAt); err != nil {
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
	plan.ID = parsedID
	plan.UserID = parsedUserID
	return &plan, nil
}