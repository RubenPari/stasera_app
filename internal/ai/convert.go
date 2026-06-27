package ai

import (
	"encoding/json"

	"github.com/stasera/stasera-api/internal/model"
)

// RawIngredient is the typed shape of an ingredient in the AI JSON response.
type RawIngredient struct {
	Name string `json:"name"`
	Qty  string `json:"qty"`
}

// RawStep is the typed shape of a preparation step in the AI JSON response.
// timer_seconds may arrive as either a JSON number (float) or an integer; the
// custom UnmarshalJSON normalizes both to int.
type RawStep struct {
	Text         string `json:"text"`
	TimerSeconds int    `json:"timer_seconds"`
}

// UnmarshalJSON accepts timer_seconds as a float or int (JSON numbers decode to
// float64 by default) and also tolerates a missing field.
func (s *RawStep) UnmarshalJSON(data []byte) error {
	type alias RawStep
	var raw alias
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	// Re-parse to capture timer_seconds regardless of numeric kind.
	var probe struct {
		TimerSeconds json.Number `json:"timer_seconds"`
	}
	_ = json.Unmarshal(data, &probe)
	if probe.TimerSeconds != "" {
		if f, err := probe.TimerSeconds.Float64(); err == nil {
			raw.TimerSeconds = int(f)
		}
	}
	*s = RawStep(raw)
	return nil
}

// ToIngredients converts the raw AI ingredients into domain recipe ingredients.
func (r RawRecipe) ToIngredients() []model.RecipeIngredient {
	out := make([]model.RecipeIngredient, 0, len(r.Ingredients))
	for _, ing := range r.Ingredients {
		out = append(out, model.RecipeIngredient{Name: ing.Name, Qty: ing.Qty})
	}
	return out
}

// ToSteps converts the raw AI steps into domain recipe steps.
func (r RawRecipe) ToSteps() []model.RecipeStep {
	out := make([]model.RecipeStep, 0, len(r.Steps))
	for _, st := range r.Steps {
		out = append(out, model.RecipeStep{Text: st.Text, TimerSeconds: st.TimerSeconds})
	}
	return out
}