package ai

import (
	"fmt"
	"strings"
)

// MealPlanInput contains the context used to build the weekly meal plan prompt.
type MealPlanInput struct {
	MaxPrepMinutes    int
	Disliked          []string
	PreferredCuisines []string
	RecentRecipes     []string
}

// RescueInput contains the context used to build the rescue meals prompt.
type RescueInput struct {
	Staples []string
}

// weekday bundles the Italian day name and its rule for the meal plan prompt.
type weekday struct {
	Name string
	Rule string
}

// weekdays maps 1=lun...5=ven; index 0 is a fallback for unrecognised values.
var weekdays = []weekday{
	{Name: "un giorno feriale", Rule: "Ricetta bilanciata"},
	{Name: "lunedì", Rule: "Lunedì: ricetta più semplice (pochi ingredienti, esecuzione meccanica)"},
	{Name: "martedì", Rule: "Martedì: ricetta più semplice (pochi ingredienti, esecuzione meccanica)"},
	{Name: "mercoledì", Rule: "Ricetta bilanciata"},
	{Name: "giovedì", Rule: "Ricetta bilanciata"},
	{Name: "venerdì", Rule: "Venerdì: qualcosa di più soddisfacente, è fine settimana"},
}

func dayNameItalian(dayOfWeek int) string {
	if dayOfWeek >= 1 && dayOfWeek <= 5 {
		return weekdays[dayOfWeek].Name
	}
	return weekdays[0].Name
}

func dayRule(dayOfWeek int) string {
	if dayOfWeek >= 1 && dayOfWeek <= 5 {
		return weekdays[dayOfWeek].Rule
	}
	return weekdays[0].Rule
}

// joinOrNone joins values with ", " or returns the provided none placeholder.
func joinOrNone(values []string, none string) string {
	if len(values) == 0 {
		return none
	}
	return strings.Join(values, ", ")
}

// mealPlanSystemTemplate is the shared template for both weekly-plan and single-day prompts.
// Placeholders: {max_prep_minutes}, {disliked_ingredients}, {preferred_cuisines},
// {recent_recipes}, {day_directive}, {format_directive}.
const mealPlanSystemTemplate = `Sei un assistente culinario per una persona single che lavora a tempo pieno.
{day_directive}

Regole obbligatorie:
- Tempo di preparazione MASSIMO: {max_prep_minutes} minuti
- Ingredienti vietati: {disliked_ingredients}
- Non ripetere proteine consecutive (es. pollo lunedì e pollo martedì)
- Lunedì e martedì: ricette più semplici (pochi ingredienti, esecuzione meccanica)
- Venerdì: qualcosa di più soddisfacente, è fine settimana
- Porzioni per 1 persona
- Cucina preferita: {preferred_cuisines}
- NON suggerire ricette già usate di recente: {recent_recipes}

Rispondi SOLO con un array JSON valido, senza testo aggiuntivo, senza markdown.
{format_directive}`

func renderMealPlanSystem(input MealPlanInput, dayOfWeek *int) string {
	disliked := joinOrNone(input.Disliked, "nessuno")
	cuisines := joinOrNone(input.PreferredCuisines, "nessuna preferenza")
	recent := joinOrNone(input.RecentRecipes, "nessuna")

	var dayDirective, formatDirective string
	if dayOfWeek != nil {
		dayDirective = fmt.Sprintf("Devi suggerire UNA sola cena per %s.", dayNameItalian(*dayOfWeek))
		formatDirective = `Formato:
[
  {
    "name": "Nome ricetta",
    "prep_minutes": 20,
    "ingredients": [{"name": "pasta", "qty": "80g"}, ...],
    "steps": [{"text": "Descrizione passo", "timer_seconds": 480}, ...]
  }
]`
	} else {
		dayDirective = "Devi suggerire 5 cene per la settimana (lunedì-venerdì)."
		formatDirective = `Formato:
[
  {
    "day_of_week": 1,
    "name": "Nome ricetta",
    "prep_minutes": 20,
    "ingredients": [{"name": "pasta", "qty": "80g"}, ...],
    "steps": [{"text": "Descrizione passo", "timer_seconds": 480}, ...]
  },
  ...
]`
	}

	r := strings.NewReplacer(
		"{day_directive}", dayDirective,
		"{max_prep_minutes}", fmt.Sprintf("%d", input.MaxPrepMinutes),
		"{disliked_ingredients}", disliked,
		"{preferred_cuisines}", cuisines,
		"{recent_recipes}", recent,
		"{format_directive}", formatDirective,
	)
	return r.Replace(mealPlanSystemTemplate)
}

// BuildMealPlanPrompt returns the system and user prompts for weekly plan generation.
func BuildMealPlanPrompt(input MealPlanInput) (system, user string) {
	system = renderMealPlanSystem(input, nil)
	user = "Genera il piano settimanale per una persona single."
	return
}

// BuildSingleRecipePrompt returns the system and user prompts to generate a single
// replacement recipe for a specific day of the week (1=lun...5=ven).
func BuildSingleRecipePrompt(input MealPlanInput, dayOfWeek int) (system, user string) {
	system = renderMealPlanSystem(input, &dayOfWeek)
	user = fmt.Sprintf("Genera una cena per %s.", dayNameItalian(dayOfWeek))
	return
}

// BuildRescuePrompt returns the system and user prompts for emergency meals.
func BuildRescuePrompt(input RescueInput) (system, user string) {
	staples := joinOrNone(input.Staples, "nessuno")

	system = strings.NewReplacer("{staples_list}", staples).Replace(
		`L'utente è esausto dopo una lunga giornata lavorativa.
Ha in casa solo questi prodotti: {staples_list}

Suggerisci esattamente 3 cene di emergenza con queste regole:
- Tempo MASSIMO 10 minuti
- SOLO ingredienti dalla lista fornita, nessun altro
- Ordinate dal più semplice al meno semplice
- La terza può essere anche fredda/senza cottura

Rispondi SOLO con un array JSON valido, senza testo aggiuntivo, senza markdown.
Stesso formato del piano settimanale, senza il campo day_of_week.`)

	user = "Cosa posso cucinare stasera in pochissimo tempo?"
	return
}