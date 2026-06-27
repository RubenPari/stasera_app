package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/ai"
	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/repository"
)
// MealPlanService orchestrates AI plan generation and persistence.
type MealPlanService struct {
	pool    *sql.DB
	ai      AIGateway
	recipes *repository.RecipeRepository
	plans   *repository.MealPlanRepository
	prefs   PreferencesStore
	staples StapleStore
}

// NewMealPlanService returns a new MealPlanService with the required dependencies.
func NewMealPlanService(
	pool *sql.DB,
	ai AIGateway,
	recipes *repository.RecipeRepository,
	plans *repository.MealPlanRepository,
	prefs PreferencesStore,
	staples StapleStore,
) *MealPlanService {
	return &MealPlanService{
		pool:    pool,
		ai:      ai,
		recipes: recipes,
		plans:   plans,
		prefs:   prefs,
		staples: staples,
	}
}
// Generate creates a new meal plan for the upcoming week (next Monday) using AI.
func (s *MealPlanService) Generate(ctx context.Context, userID uuid.UUID) (model.MealPlan, []model.MealPlanDay, error) {
	prefs, err := s.prefs.GetByUserID(ctx, userID)
	if err != nil {
		return model.MealPlan{}, nil, fmt.Errorf("load preferences: %w", err)
	}

	recent, err := s.recipes.FindRecentNames(ctx, userID)
	if err != nil {
		return model.MealPlan{}, nil, fmt.Errorf("load recent recipes: %w", err)
	}

	input := ai.MealPlanInput{RecentRecipes: recent, MaxPrepMinutes: defaultMaxPrepMinutes}
	if prefs != nil {
		input.MaxPrepMinutes = prefs.MaxPrepMinutes
		input.Disliked = prefs.DislikedIngredients
		input.PreferredCuisines = prefs.PreferredCuisines
	}

	rawRecipes, err := s.ai.GenerateMealPlan(ctx, input)
	if err != nil {
		return model.MealPlan{}, nil, fmt.Errorf("AI generation: %w", err)
	}

	if len(rawRecipes) != 5 {
		return model.MealPlan{}, nil, fmt.Errorf("AI returned %d recipes instead of 5", len(rawRecipes))
	}

	weekStart := nextMonday()

	if exists, err := s.plans.PlanExistsForWeek(ctx, userID, weekStart); err != nil {
		return model.MealPlan{}, nil, err
	} else if exists {
		return model.MealPlan{}, nil, fmt.Errorf("plan already exists for week starting %s", weekStart.Format("2006-01-02"))
	}

	if err := s.plans.ArchiveOldPlans(ctx, userID, weekStart); err != nil {
		return model.MealPlan{}, nil, err
	}

	tx, err := s.pool.BeginTx(ctx, nil)
	if err != nil {
		return model.MealPlan{}, nil, fmt.Errorf("begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	recipeRepo := s.recipes.WithTx(tx)
	planRepo := s.plans.WithTx(tx)

	plan, err := planRepo.Create(ctx, userID, weekStart)
	if err != nil {
		return model.MealPlan{}, nil, fmt.Errorf("create plan: %w", err)
	}

	days := make([]model.MealPlanDay, 0, 5)
	for _, raw := range rawRecipes {
		recipe, err := recipeRepo.Create(ctx, userID, raw.Name, raw.PrepMinutes, 1, raw.ToIngredients(), raw.ToSteps(), false)
		if err != nil {
			return model.MealPlan{}, nil, fmt.Errorf("save recipe: %w", err)
		}
		day, err := planRepo.AddDay(ctx, plan.ID, raw.DayOfWeek, recipe.ID)
		if err != nil {
			return model.MealPlan{}, nil, fmt.Errorf("save plan day: %w", err)
		}
		day.Recipe = &recipe
		days = append(days, day)
	}

	if err := tx.Commit(); err != nil {
		return model.MealPlan{}, nil, fmt.Errorf("commit plan: %w", err)
	}
	committed = true
	return plan, days, nil
}

// SwapDay replaces the recipe for a specific day of the plan.
func (s *MealPlanService) SwapDay(ctx context.Context, userID uuid.UUID, planID uuid.UUID, dayOfWeek int, recipeID uuid.UUID) (model.MealPlanDay, error) {
	plan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		return model.MealPlanDay{}, err
	}
	if plan == nil || plan.UserID != userID {
		return model.MealPlanDay{}, fmt.Errorf("plan not found")
	}

	recipe, err := s.recipes.GetByID(ctx, recipeID)
	if err != nil {
		return model.MealPlanDay{}, err
	}
	if recipe == nil || recipe.UserID != userID {
		return model.MealPlanDay{}, fmt.Errorf("recipe not found")
	}

	day, err := s.plans.UpdateDayRecipe(ctx, planID, dayOfWeek, recipeID)
	if err != nil {
		return model.MealPlanDay{}, err
	}
	day.Recipe = recipe
	return day, nil
}

// RegenerateDay asks the AI for a brand new recipe for the given day and replaces it
// in the plan. The previous recipe remains in the user's cache but is no longer linked.
func (s *MealPlanService) RegenerateDay(ctx context.Context, userID uuid.UUID, planID uuid.UUID, dayOfWeek int) (model.MealPlanDay, error) {
	if dayOfWeek < 1 || dayOfWeek > 5 {
		return model.MealPlanDay{}, fmt.Errorf("invalid day of week")
	}

	plan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		return model.MealPlanDay{}, err
	}
	if plan == nil || plan.UserID != userID {
		return model.MealPlanDay{}, fmt.Errorf("plan not found")
	}

	prefs, err := s.prefs.GetByUserID(ctx, userID)
	if err != nil {
		return model.MealPlanDay{}, fmt.Errorf("load preferences: %w", err)
	}

	recent, err := s.recipes.FindRecentNames(ctx, userID)
	if err != nil {
		return model.MealPlanDay{}, fmt.Errorf("load recent recipes: %w", err)
	}

	input := ai.MealPlanInput{RecentRecipes: recent, MaxPrepMinutes: defaultMaxPrepMinutes}
	if prefs != nil {
		input.MaxPrepMinutes = prefs.MaxPrepMinutes
		input.Disliked = prefs.DislikedIngredients
		input.PreferredCuisines = prefs.PreferredCuisines
	}

	raw, err := s.ai.GenerateSingleRecipe(ctx, input, dayOfWeek)
	if err != nil {
		return model.MealPlanDay{}, fmt.Errorf("AI generation: %w", err)
	}

	recipe, err := s.recipes.Create(ctx, userID, raw.Name, raw.PrepMinutes, 1, raw.ToIngredients(), raw.ToSteps(), false)
	if err != nil {
		return model.MealPlanDay{}, fmt.Errorf("save recipe: %w", err)
	}

	day, err := s.plans.UpdateDayRecipe(ctx, planID, dayOfWeek, recipe.ID)
	if err != nil {
		return model.MealPlanDay{}, fmt.Errorf("update plan day: %w", err)
	}
	day.Recipe = &recipe
	return day, nil
}

// GetCurrent returns the active meal plan with days and recipes populated.
func (s *MealPlanService) GetCurrent(ctx context.Context, userID uuid.UUID) (*model.MealPlan, error) {
	plan, err := s.plans.GetCurrent(ctx, userID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, nil
	}

	if err := s.loadDays(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// GetToday returns the recipe assigned to the current weekday in the active plan.
func (s *MealPlanService) GetToday(ctx context.Context, userID uuid.UUID) (*model.Recipe, error) {
	recipeID, err := s.plans.GetTodayRecipeID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if recipeID == nil {
		return nil, nil
	}
	return s.recipes.GetByID(ctx, *recipeID)
}

func (s *MealPlanService) loadDays(ctx context.Context, plan *model.MealPlan) error {
	days, err := s.plans.GetDays(ctx, plan.ID)
	if err != nil {
		return err
	}

	recipes, err := s.recipes.GetByPlanID(ctx, plan.ID)
	if err != nil {
		return err
	}
	recipeByID := make(map[uuid.UUID]model.Recipe, len(recipes))
	for _, r := range recipes {
		recipeByID[r.ID] = r
	}

	for i := range days {
		if r, ok := recipeByID[days[i].RecipeID]; ok {
			days[i].Recipe = &r
		}
	}
	plan.Days = days
	return nil
}

// defaultMaxPrepMinutes is applied when the user has no preferences row yet.
const defaultMaxPrepMinutes = 30

func nextMonday() time.Time {
	today := time.Now()
	wd := int(today.Weekday())
	delta := 1 - wd
	if delta <= 0 {
		delta += 7
	}
	return today.AddDate(0, 0, delta)
}
