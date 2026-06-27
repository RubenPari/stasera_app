package service

import (
	"math"
	"strings"
	"testing"

	"github.com/stasera/stasera-api/internal/model"
)

func TestParseQuantity(t *testing.T) {
	cases := []struct {
		input string
		v     float64
		unit  string
		ok    bool
	}{
		{"80g", 80, "g", true},
		{"1 scatoletta", 1, "scatoletta", true},
		{"6", 6, "", true},
		{"1,5 kg", 1.5, "kg", true},
		{"q.b.", 0, "", false},
		{"mezzo", 0, "", false},
		{"", 0, "", false},
	}
	for _, tc := range cases {
		v, u, ok := parseQuantity(tc.input)
		if ok != tc.ok {
			t.Errorf("parseQuantity(%q): ok=%v want %v", tc.input, ok, tc.ok)
			continue
		}
		if !ok {
			continue
		}
		if math.Abs(v-tc.v) > 1e-9 {
			t.Errorf("parseQuantity(%q): value=%.4f want %.4f", tc.input, v, tc.v)
		}
		if strings.TrimSpace(u) != tc.unit {
			t.Errorf("parseQuantity(%q): unit=%q want %q", tc.input, u, tc.unit)
		}
	}
}

func TestCategorize(t *testing.T) {
	// Expected aisles match the actual aisleRules defined in shopping_list_service.go.
	cases := []struct {
		name  string
		aisle string
	}{
		{"pasta", "dispensa"},
		{"pomodori pelati in scatola", "dispensa"},
		{"manzo macinato", "carne"},
		{"salmone", "carne"},   // salmone is in the carne aisle
		{"latte", "frigo"},
		{"mela", "verdura"},    // mela is in verdura aisle
		{"dado da brodo", "dispensa"},
	}
	for _, tc := range cases {
		got := categorize(tc.name)
		if got != tc.aisle {
			t.Errorf("categorize(%q)=%q want %q", tc.name, got, tc.aisle)
		}
	}
}

func TestAggregateIngredients_Sums(t *testing.T) {
	ings := []model.RecipeIngredient{
		{Name: "pasta", Qty: "80g"},
		{Name: "pasta", Qty: "100g"},
	}
	items := aggregateIngredients(ings)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Quantity != "180g" {
		t.Errorf("expected 180g, got %q", items[0].Quantity)
	}
}

func TestAggregateIngredients_IncompatibleUnits(t *testing.T) {
	ings := []model.RecipeIngredient{
		{Name: "sale", Qty: "1g"},
		{Name: "sale", Qty: "1 cucchiaio"},
	}
	items := aggregateIngredients(ings)
	if len(items) != 1 {
		t.Fatalf("expected 1 aggregated item, got %d", len(items))
	}
	if !strings.Contains(items[0].Quantity, "g") {
		t.Errorf("expected 'g' in quantity, got %q", items[0].Quantity)
	}
	if !strings.Contains(items[0].Quantity, "cucchiaio") {
		t.Errorf("expected 'cucchiaio' in quantity, got %q", items[0].Quantity)
	}
}

func TestAggregateIngredients_NonNumeric(t *testing.T) {
	items := aggregateIngredients([]model.RecipeIngredient{
		{Name: "aglio", Qty: "q.b."},
	})
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Quantity != "q.b." {
		t.Errorf("expected q.b., got %q", items[0].Quantity)
	}
}