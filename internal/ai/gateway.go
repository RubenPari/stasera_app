package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// ChatClient is the minimal OpenAI-compatible surface the Gateway depends on,
// so a fake can be substituted in tests. *openai.Client satisfies it.
type ChatClient interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// GatewayConfig holds configurable Gateway parameters.
type GatewayConfig struct {
	MaxTokens       int
	Temperature     float64
	TimeoutSeconds  int
}

// DefaultGatewayConfig is the baseline config used when specific values are zero.
var DefaultGatewayConfig = GatewayConfig{
	MaxTokens:      2000,
	Temperature:    0.7,
	TimeoutSeconds: 30,
}

// Gateway wraps the OpenAI-compatible Vercel AI Gateway client.
type Gateway struct {
	client ChatClient
	model  string
	cfg    GatewayConfig
}

// NewGateway creates a Gateway using the provided client, model name, and optional config.
// Zero fields in cfg fall back to DefaultGatewayConfig values.
func NewGateway(client ChatClient, model string, cfg GatewayConfig) *Gateway {
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = DefaultGatewayConfig.MaxTokens
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = DefaultGatewayConfig.Temperature
	}
	if cfg.TimeoutSeconds == 0 {
		cfg.TimeoutSeconds = DefaultGatewayConfig.TimeoutSeconds
	}
	return &Gateway{client: client, model: model, cfg: cfg}
}

// RawRecipe is the intermediate shape returned by the AI before being stored.
type RawRecipe struct {
	DayOfWeek   int             `json:"day_of_week,omitempty"`
	Name        string          `json:"name"`
	PrepMinutes int             `json:"prep_minutes"`
	Ingredients []RawIngredient `json:"ingredients"`
	Steps       []RawStep       `json:"steps"`
}

// GenerateMealPlan asks the AI for 5 weeknight dinners and returns parsed raw recipes.
func (g *Gateway) GenerateMealPlan(ctx context.Context, input MealPlanInput) ([]RawRecipe, error) {
	system, user := BuildMealPlanPrompt(input)
	return g.requestRecipes(ctx, system, user)
}

// GenerateRescueMeals asks the AI for 3 emergency meals using only the provided staples.
func (g *Gateway) GenerateRescueMeals(ctx context.Context, input RescueInput) ([]RawRecipe, error) {
	system, user := BuildRescuePrompt(input)
	return g.requestRecipes(ctx, system, user)
}

// GenerateSingleRecipe asks the AI for a single dinner for the given day of week.
func (g *Gateway) GenerateSingleRecipe(ctx context.Context, input MealPlanInput, dayOfWeek int) (RawRecipe, error) {
	system, user := BuildSingleRecipePrompt(input, dayOfWeek)
	recipes, err := g.requestRecipes(ctx, system, user)
	if err != nil {
		return RawRecipe{}, err
	}
	if len(recipes) == 0 {
		return RawRecipe{}, fmt.Errorf("AI returned no recipe")
	}
	recipes[0].DayOfWeek = dayOfWeek
	return recipes[0], nil
}

// retryBackoffs are the wait durations between attempts (3 attempts total).
var retryBackoffs = []time.Duration{500 * time.Millisecond, time.Second, 2 * time.Second}

func (g *Gateway) requestRecipes(ctx context.Context, systemPrompt, userPrompt string) ([]RawRecipe, error) {
	if g.client == nil {
		return nil, fmt.Errorf("AI gateway client is not configured")
	}

	timeout := time.Duration(g.cfg.TimeoutSeconds) * time.Second
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-reqCtx.Done():
				return nil, fmt.Errorf("AI request cancelled: %w", reqCtx.Err())
			case <-time.After(retryBackoffs[attempt-1]):
			}
		}

		resp, err := g.client.CreateChatCompletion(reqCtx, openai.ChatCompletionRequest{
			Model: g.model,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
				{Role: openai.ChatMessageRoleUser, Content: userPrompt},
			},
			MaxTokens:   g.cfg.MaxTokens,
			Temperature: float32(g.cfg.Temperature),
		})
		if err != nil {
			lastErr = err
			// Retry on server errors and rate-limit; fail fast on client errors.
			if isRetryableAIError(err) {
				continue
			}
			return nil, fmt.Errorf("AI gateway request failed: %w", err)
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("AI gateway returned no choices")
		}

		content := strings.TrimSpace(resp.Choices[0].Message.Content)
		content = stripMarkdownCodeFence(content)

		var recipes []RawRecipe
		if err := json.Unmarshal([]byte(content), &recipes); err != nil {
			return nil, fmt.Errorf("parse AI response: %w", err)
		}
		return recipes, nil
	}
	return nil, fmt.Errorf("AI gateway request failed after retries: %w", lastErr)
}

// isRetryableAIError returns true for transient HTTP 5xx or 429 errors.
func isRetryableAIError(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		return apiErr.HTTPStatusCode == 429 || apiErr.HTTPStatusCode >= 500
	}
	return false
}

// codeFenceRe matches optional whitespace + code fence + optional json language + content.
var codeFenceRe = regexp.MustCompile(`(?is)^\s*` + "```" + `(?:json)?\s*(.*?)\s*` + "```" + `\s*$`)

func stripMarkdownCodeFence(s string) string {
	if m := codeFenceRe.FindStringSubmatch(s); m != nil {
		return strings.TrimSpace(m[1])
	}
	return strings.TrimSpace(s)
}