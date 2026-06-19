package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

const maxTokens = 2000

// Gateway wraps the OpenAI-compatible Vercel AI Gateway client.
type Gateway struct {
	client *openai.Client
	model  string
}

// NewGateway creates a Gateway using the provided client and model name.
func NewGateway(client *openai.Client, model string) *Gateway {
	return &Gateway{client: client, model: model}
}

// RawRecipe is the intermediate shape returned by the AI before being stored.
type RawRecipe struct {
	DayOfWeek   int                    `json:"day_of_week,omitempty"`
	Name        string                 `json:"name"`
	PrepMinutes int                    `json:"prep_minutes"`
	Ingredients []map[string]string    `json:"ingredients"`
	Steps       []map[string]interface{} `json:"steps"`
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
// Used by the SwapDay regenerate path.
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

func (g *Gateway) requestRecipes(ctx context.Context, systemPrompt, userPrompt string) ([]RawRecipe, error) {
	if g.client == nil {
		return nil, fmt.Errorf("AI gateway client is not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: g.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		MaxTokens:   maxTokens,
		Temperature: 0.7,
	})
	if err != nil {
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

func stripMarkdownCodeFence(s string) string {
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
