package ai

import (
	"fmt"
	"strings"
)

// MealPlanInput contains the context used to build the weekly meal plan prompt.
type MealPlanInput struct {
	MaxPrepMinutes   int
	Disliked         []string
	PreferredCuisines []string
	RecentRecipes    []string
}

// RescueInput contains the context used to build the rescue meals prompt.
type RescueInput struct {
	Staples []string
}

// BuildMealPlanPrompt returns the system and user prompts for weekly plan generation.
func BuildMealPlanPrompt(input MealPlanInput) (system, user string) {
	disliked := joinOrNone(input.Disliked)
	cuisines := joinOrNone(input.PreferredCuisines)
	recent := joinOrNone(input.RecentRecipes)

	system = `Sei un assistente culinario per una persona single che lavora a tempo pieno.
Devi suggerire 5 cene per la settimana (lunedì-venerdì).

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
Formato:
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

	system = strings.ReplaceAll(system, "{max_prep_minutes}", fmt.Sprintf("%d", input.MaxPrepMinutes))
	system = strings.ReplaceAll(system, "{disliked_ingredients}", disliked)
	system = strings.ReplaceAll(system, "{preferred_cuisines}", cuisines)
	system = strings.ReplaceAll(system, "{recent_recipes}", recent)

	user = "Genera il piano settimanale per una persona single."
	return system, user
}

// BuildRescuePrompt returns the system and user prompts for emergency meals.
func BuildRescuePrompt(input RescueInput) (system, user string) {
	staples := joinOrNone(input.Staples)

	system = `L'utente è esausto dopo una lunga giornata lavorativa.
Ha in casa solo questi prodotti: {staples_list}

Suggerisci esattamente 3 cene di emergenza con queste regole:
- Tempo MASSIMO 10 minuti
- SOLO ingredienti dalla lista fornita, nessun altro
- Ordinate dal più semplice al meno semplice
- La terza può essere anche fredda/senza cottura

Rispondi SOLO con un array JSON valido, senza testo aggiuntivo, senza markdown.
Stesso formato del piano settimanale, senza il campo day_of_week.`

	system = strings.ReplaceAll(system, "{staples_list}", staples)

	user = "Cosa posso cucinare stasera in pochissimo tempo?"
	return system, user
}

func joinOrNone(values []string) string {
	if len(values) == 0 {
		return "nessuno"
	}
	return strings.Join(values, ", ")
}

// BuildSingleRecipePrompt returns the system and user prompts to generate a single
// replacement recipe for a specific day of the week (1=lun ... 5=ven).
func BuildSingleRecipePrompt(input MealPlanInput, dayOfWeek int) (system, user string) {
	disliked := joinOrNone(input.Disliked)
	cuisines := joinOrNone(input.PreferredCuisines)
	recent := joinOrNone(input.RecentRecipes)
	dayName := dayNameItalian(dayOfWeek)

	system = `Sei un assistente culinario per una persona single che lavora a tempo pieno.
Devi suggerire UNA sola cena per {day_name}.

Regole obbligatorie:
- Tempo di preparazione MASSIMO: {max_prep_minutes} minuti
- Ingredienti vietati: {disliked_ingredients}
- Porzioni per 1 persona
- Cucina preferita: {preferred_cuisines}
- NON suggerire ricette già usate di recente: {recent_recipes}
- {day_rule}

Rispondi SOLO con un array JSON valido contenente un singolo oggetto, senza testo aggiuntivo, senza markdown.
Formato:
[
  {
    "name": "Nome ricetta",
    "prep_minutes": 20,
    "ingredients": [{"name": "pasta", "qty": "80g"}, ...],
    "steps": [{"text": "Descrizione passo", "timer_seconds": 480}, ...]
  }
]`

	system = strings.ReplaceAll(system, "{day_name}", dayName)
	system = strings.ReplaceAll(system, "{max_prep_minutes}", fmt.Sprintf("%d", input.MaxPrepMinutes))
	system = strings.ReplaceAll(system, "{disliked_ingredients}", disliked)
	system = strings.ReplaceAll(system, "{preferred_cuisines}", cuisines)
	system = strings.ReplaceAll(system, "{recent_recipes}", recent)
	system = strings.ReplaceAll(system, "{day_rule}", dayRule(dayOfWeek))

	user = fmt.Sprintf("Genera una cena per %s.", dayName)
	return system, user
}

func dayNameItalian(dayOfWeek int) string {
	names := map[int]string{1: "lunedì", 2: "martedì", 3: "mercoledì", 4: "giovedì", 5: "venerdì"}
	if n, ok := names[dayOfWeek]; ok {
		return n
	}
	return "un giorno feriale"
}

func dayRule(dayOfWeek int) string {
	switch dayOfWeek {
	case 1, 2:
		return "Lunedì o martedì: ricetta più semplice (pochi ingredienti, esecuzione meccanica)"
	case 5:
		return "Venerdì: qualcosa di più soddisfacente, è fine settimana"
	default:
		return "Ricetta bilanciata"
	}
}
