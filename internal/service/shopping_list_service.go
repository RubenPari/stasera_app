package service

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/repository"
)

// ShoppingListService generates shopping lists from a meal plan and manages their lifecycle.
type ShoppingListService struct {
	plans    *repository.MealPlanRepository
	recipes  *repository.RecipeRepository
	lists    *repository.ShoppingListRepository
}

// NewShoppingListService returns a new ShoppingListService with the required dependencies.
func NewShoppingListService(
	plans *repository.MealPlanRepository,
	recipes *repository.RecipeRepository,
	lists *repository.ShoppingListRepository,
) *ShoppingListService {
	return &ShoppingListService{plans: plans, recipes: recipes, lists: lists}
}

// Generate aggregates ingredients from the current active meal plan into a new shopping list,
// replacing any previous open list for the user.
func (s *ShoppingListService) Generate(ctx context.Context, userID uuid.UUID) (*model.ShoppingList, error) {
	plan, err := s.plans.GetCurrent(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load current plan: %w", err)
	}
	if plan == nil {
		return nil, ErrNoActivePlan
	}

	ingredients, err := s.recipes.GetIngredientsByPlanID(ctx, plan.ID)
	if err != nil {
		return nil, fmt.Errorf("load plan ingredients: %w", err)
	}

	aggregated := aggregateIngredients(ingredients)

	if err := s.lists.DeleteOpenByUserID(ctx, userID); err != nil {
		return nil, fmt.Errorf("clear previous list: %w", err)
	}

	list, err := s.lists.CreateWithItems(ctx, userID, &plan.ID, aggregated)
	if err != nil {
		return nil, fmt.Errorf("create shopping list: %w", err)
	}

	items, err := s.lists.GetItemsByListID(ctx, list.ID)
	if err != nil {
		return nil, fmt.Errorf("load created items: %w", err)
	}
	list.Items = items
	return &list, nil
}

// GetCurrent returns the user's open shopping list with items, if any.
func (s *ShoppingListService) GetCurrent(ctx context.Context, userID uuid.UUID) (*model.ShoppingList, error) {
	list, err := s.lists.GetCurrent(ctx, userID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, nil
	}
	items, err := s.lists.GetItemsByListID(ctx, list.ID)
	if err != nil {
		return nil, err
	}
	list.Items = items
	return list, nil
}

// UpdateItem toggles an item's checked state, validating ownership.
func (s *ShoppingListService) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, isChecked bool) (*model.ShoppingItem, error) {
	item, err := s.lists.UpdateItemChecked(ctx, userID, itemID, isChecked)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrItemNotFound
	}
	return item, nil
}

// Complete marks the user's open shopping list as completed.
func (s *ShoppingListService) Complete(ctx context.Context, userID uuid.UUID) (*model.ShoppingList, error) {
	list, err := s.lists.MarkCompleted(ctx, userID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, ErrNoOpenList
	}
	items, err := s.lists.GetItemsByListID(ctx, list.ID)
	if err != nil {
		return nil, err
	}
	list.Items = items
	return list, nil
}

// Sentinel errors returned to handlers for clean HTTP mapping.
var (
	ErrNoActivePlan  = fmt.Errorf("no active meal plan")
	ErrItemNotFound  = fmt.Errorf("shopping item not found")
	ErrNoOpenList    = fmt.Errorf("no open shopping list")
)

// aisleOrder defines the display order of aisles in the shopping list.
var aisleOrder = map[string]int{
	"carne":    1,
	"frigo":    2,
	"verdura":  3,
	"dispensa": 4,
	"altro":    5,
}

// aisleKeywords maps each aisle to the list of substrings that identify it.
// The first matching aisle wins, so order matters.
type aisleRule struct {
	aisle    string
	keywords []string
}

var aisleRules = []aisleRule{
	{
		aisle: "carne",
		keywords: []string{
			"pollo", "manzo", "vitello", "vitellone", "maiale", "pancetta", "guanciale",
			"salsiccia", "hamburger", "wurstel", "carne", "fegato", "trippa", "ossobuco",
			"petto", "fesa", "coscia", "spalla", "lombata", "bistecca",
			"pesce", "salmone", "trota", "merluzzo", "baccalà", "tonno fresco",
			"gamberi", "gamberetti", "calamari", "seppie", "vongole", "cozze",
			"prosciutto", "salame", "bresaola", "speck", "mortadella", "salumi", "salsiccia",
		},
	},
	{
		aisle: "frigo",
		keywords: []string{
			"uova", "latte", "burro", "formaggio", "mozzarella", "parmigiano", "pecorino",
			"ricotta", "yogurt", "panna", "feta", "scamorza", "gorgonzola", "latticini",
			"stracchino", "robiola", "burrata", "crescenza", "taleggio", "fontina",
			"uovo", "besciamella", "pasta sfoglia", "pasta brisé", "pasta per pizza",
		},
	},
	{
		aisle: "verdura",
		keywords: []string{
			"pomodoro", "pomodorini", "insalata", "lattuga", "spinaci", "zucchine",
			"carote", "carota", "cipolla", "cipollotto", "porro", "sedano",
			"basilico", "prezzemolo", "rosmarino", "salvia", "timo", "menta", "erba cipollina",
			"funghi", "peperoni", "melanzane", "patate", "patata",
			"frutta", "mela", "mela", "banana", "limone", "arancia", "kiwi",
			"fragola", "fragole", "pesca", "albicocca", "ciliegia", "ciliegie",
			"broccoli", "cavolfiore", "cavolo", "cavolini", "verza",
			"fagiolini", "piselli", "freschi", "verdura", "zucchina",
			"agretti", "asparagi", "cicoria", "radicchio", "indivia", "rucola",
			"valeriana", "songino", "scarola", "cicorino",
		},
	},
	{
		aisle: "dispensa",
		keywords: []string{
			"pasta", "riso", "pane", "farina", "semolino", "fecola",
			"olio", "aceto", "sale", "pepe", "peperoncino", "zucchero",
			"pomodori pelati", "pelati", "passata", "concentrato",
			"scatola", "scatoletta", "tonno in scatoletta", "tonno",
			"legumi", "fagioli", "lenticchie", "ceci", "piselli in scatola",
			"dado", "brodo", "cacao", "caffè", "the", "tisana",
			"aglio", "cipolla essiccata", "pangrattato", "pan grattato",
			"lievito", "bicarbonato", "vaniglia", "cannella", "noce moscata",
			"pasta sfoglia", "besciamella", "miele", "marmellata",
			"piselli surgelati", "spinaci surgelati", "surgelati",
			"maizena", "amido", "panna in brick",
			"capperi", "olive", "funghi secchi", "funghi champignon",
			"zenzero", "curry", "curcuma", "paprika", "origano", "alloro",
		},
	},
}

// quantityPattern matches a numeric part followed by an optional unit.
var quantityPattern = regexp.MustCompile(`^\s*([0-9]+(?:[.,][0-9]+)?)\s*([a-zA-Zµμ%°/0-9.\s\-+]*)$`)

// aggregateIngredients groups ingredients by normalized name, sums numeric quantities
// with a compatible unit, and assigns each group to an aisle. Items are sorted by
// aisle order then by name.
func aggregateIngredients(ingredients []model.RecipeIngredient) []model.ShoppingItem {
	type bucket struct {
		name      string
		displayName string
		total     float64
		unit      string
		hasNumber bool
		extras    []string
	}

	buckets := make(map[string]*bucket)
	order := make([]string, 0, len(ingredients))

	for _, ing := range ingredients {
		key := normalizeName(ing.Name)
		b, ok := buckets[key]
		if !ok {
			b = &bucket{name: key, displayName: ing.Name}
			buckets[key] = b
			order = append(order, key)
		}

		num, unit, ok := parseQuantity(ing.Qty)
		if !ok {
			// Non numeric quantity: keep as a separate textual amount.
			b.extras = append(b.extras, ing.Qty)
			continue
		}

		if !b.hasNumber {
			b.total = num
			b.unit = unit
			b.hasNumber = true
			continue
		}

		// Same unit (or both unit-less) -> sum.
		if b.unit == unit {
			b.total += num
			continue
		}

		// Incompatible units: keep the second as a separate textual amount.
		b.extras = append(b.extras, formatQuantity(num, unit))
	}

	items := make([]model.ShoppingItem, 0, len(buckets))
	for _, key := range order {
		b := buckets[key]
		aisle := categorize(b.name)

		qty := ""
		if b.hasNumber {
			qty = formatQuantity(b.total, b.unit)
		}
		if len(b.extras) > 0 {
			if qty != "" {
				qty += ", "
			}
			qty += strings.Join(b.extras, ", ")
		}
		if qty == "" {
			qty = "q.b."
		}

		items = append(items, model.ShoppingItem{
			Name:     b.displayName,
			Quantity: qty,
			Aisle:    aisle,
		})
	}

	sortItems(items)
	return items
}

func sortItems(items []model.ShoppingItem) {
	sort.SliceStable(items, func(i, j int) bool {
		ai, aj := aisleOrder[items[i].Aisle], aisleOrder[items[j].Aisle]
		if ai == 0 {
			ai = aisleOrder["altro"]
		}
		if aj == 0 {
			aj = aisleOrder["altro"]
		}
		if ai != aj {
			return ai < aj
		}
		return items[i].Name < items[j].Name
	})
	for i := range items {
		items[i].SortOrder = i
	}
}

func categorize(name string) string {
	n := strings.ToLower(name)
	for _, rule := range aisleRules {
		for _, kw := range rule.keywords {
			if strings.Contains(n, kw) {
				return rule.aisle
			}
		}
	}
	return "altro"
}

func normalizeName(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' {
			b.WriteRune(' ')
		}
	}
	out := strings.Join(strings.Fields(b.String()), " ")
	return out
}

// parseQuantity extracts the numeric amount and its unit from a quantity string like
// "80g", "1 scatoletta", "6 pz", "mezzo", "q.b.".
func parseQuantity(qty string) (value float64, unit string, ok bool) {
	qty = strings.TrimSpace(qty)
	if qty == "" {
		return 0, "", false
	}

	m := quantityPattern.FindStringSubmatch(qty)
	if m == nil {
		return 0, "", false
	}

	numStr := strings.ReplaceAll(m[1], ",", ".")
	v, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, "", false
	}
	unit = strings.TrimSpace(m[2])
	return v, unit, true
}

func formatQuantity(v float64, unit string) string {
	// Render integers without decimals.
	s := strconv.FormatFloat(v, 'f', -1, 64)
	if unit == "" {
		return s
	}
	return s + unit
}