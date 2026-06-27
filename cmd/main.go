package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/stasera/stasera-api/internal/ai"
	"github.com/stasera/stasera-api/internal/config"
	"github.com/stasera/stasera-api/internal/db"
	"github.com/stasera/stasera-api/internal/handler"
	"github.com/stasera/stasera-api/internal/middleware"
	"github.com/stasera/stasera-api/internal/repository"
	"github.com/stasera/stasera-api/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("database pool: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	userRepo := repository.NewUserRepository(pool)
	stapleRepo := repository.NewStapleRepository(pool)
	prefsRepo := repository.NewPreferencesRepository(pool)
	jwtManager := middleware.NewJWTManager(cfg.JWTSecret, cfg.JWTAccessExpiryMinutes, cfg.JWTRefreshExpiryDays)
	authHandler := handler.NewAuthHandler(userRepo, jwtManager)

	aiClient := ai.NewGatewayClient(cfg.AIGatewayAPIKey, cfg.AIGatewayBaseURL)
	aiGateway := ai.NewGateway(aiClient, cfg.AIModel, ai.GatewayConfig{
		MaxTokens:      cfg.AIMaxTokens,
		Temperature:    cfg.AITemperature,
		TimeoutSeconds: cfg.AITimeoutSeconds,
	})

	recipeRepo := repository.NewRecipeRepository(pool)
	mealPlanRepo := repository.NewMealPlanRepository(pool)
	shoppingRepo := repository.NewShoppingListRepository(pool)
	mealPlanService := service.NewMealPlanService(pool, aiGateway, recipeRepo, mealPlanRepo, prefsRepo, stapleRepo)
	mealPlanHandler := handler.NewMealPlanHandler(mealPlanService)
	rescueService := service.NewRescueService(aiGateway, recipeRepo, stapleRepo)
	rescueHandler := handler.NewRescueHandler(rescueService)

	shoppingService := service.NewShoppingListService(pool, mealPlanRepo, recipeRepo, shoppingRepo)
	shoppingHandler := handler.NewShoppingListHandler(shoppingService)
	recipeHandler := handler.NewRecipeHandler(recipeRepo)
	stapleHandler := handler.NewStapleHandler(stapleRepo)
	prefsHandler := handler.NewPreferencesHandler(prefsRepo)

	e := echo.New()
	e.Validator = handler.NewValidator()
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	// Public auth routes.
	e.POST("/api/v1/auth/register", authHandler.Register)
	e.POST("/api/v1/auth/login", authHandler.Login)
	e.POST("/api/v1/auth/refresh", authHandler.Refresh)

	// Protected routes.
	api := e.Group("/api/v1", middleware.AuthMiddleware(jwtManager))
	api.GET("/auth/me", authHandler.Me)
	api.PATCH("/auth/me", authHandler.UpdateMe)
	api.POST("/auth/change-password", authHandler.ChangePassword)
	// Meal plan routes.
	api.GET("/meal-plan/current", mealPlanHandler.Current)
	api.POST("/meal-plan/generate", mealPlanHandler.Generate)
	api.PATCH("/meal-plan/:planId/days/:dayOfWeek", mealPlanHandler.SwapDay)
	api.GET("/meal-plan/today", mealPlanHandler.Today)
// AI routes.
api.POST("/ai/rescue", rescueHandler.Rescue)

	// Recipe routes.
	api.GET("/recipes", recipeHandler.List)
	api.GET("/recipes/:id", recipeHandler.Get)
	api.POST("/recipes/:id/cooked", recipeHandler.MarkCooked)
	api.DELETE("/recipes/:id", recipeHandler.Delete)

	// Shopping list routes.
	api.GET("/shopping-list/current", shoppingHandler.Current)
	api.POST("/shopping-list/generate", shoppingHandler.Generate)
	api.PATCH("/shopping-list/items/:itemId", shoppingHandler.UpdateItem)
	api.POST("/shopping-list/current/complete", shoppingHandler.Complete)

	// Staples routes.
	api.GET("/staples", stapleHandler.List)
	api.POST("/staples", stapleHandler.Create)
	api.PATCH("/staples/:id", stapleHandler.Update)
	api.DELETE("/staples/:id", stapleHandler.Delete)

	// Preferences routes.
	api.GET("/preferences", prefsHandler.Get)
	api.PATCH("/preferences", prefsHandler.Update)

	// Graceful shutdown.
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
}
