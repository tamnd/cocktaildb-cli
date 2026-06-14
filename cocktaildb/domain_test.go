package cocktaildb

import (
	"testing"
)

// These tests are offline: they exercise the URI driver's pure string functions
// and the domain registration. The client's HTTP behaviour is in cocktaildb_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "cocktaildb" {
		t.Errorf("Scheme = %q, want cocktaildb", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "cocktaildb" {
		t.Errorf("Identity.Binary = %q, want cocktaildb", info.Identity.Binary)
	}
}

func TestDomainInfoHost(t *testing.T) {
	if Host != "www.thecocktaildb.com" {
		t.Errorf("Host = %q, want www.thecocktaildb.com", Host)
	}
}

func TestClassifyNumericIsID(t *testing.T) {
	typ, id, err := Domain{}.Classify("11007")
	if err != nil {
		t.Fatal(err)
	}
	if typ != "id" {
		t.Errorf("Classify(\"11007\") type = %q, want id", typ)
	}
	if id != "11007" {
		t.Errorf("Classify(\"11007\") id = %q, want 11007", id)
	}
}

func TestClassifyStringIsQuery(t *testing.T) {
	typ, id, err := Domain{}.Classify("margarita")
	if err != nil {
		t.Fatal(err)
	}
	if typ != "query" {
		t.Errorf("Classify(\"margarita\") type = %q, want query", typ)
	}
	if id != "margarita" {
		t.Errorf("Classify(\"margarita\") id = %q, want margarita", id)
	}
}

func TestClassifyEmpty(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestLocateID(t *testing.T) {
	got, err := Domain{}.Locate("id", "11007")
	want := "https://www.thecocktaildb.com/drink/11007"
	if err != nil || got != want {
		t.Errorf("Locate(id, 11007) = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestLocateQuery(t *testing.T) {
	got, err := Domain{}.Locate("query", "margarita")
	want := "https://www.thecocktaildb.com/drink/margarita-Detail.php"
	if err != nil || got != want {
		t.Errorf("Locate(query, margarita) = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("unknown", "foo")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}

func TestDefaultConfigRate(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Rate.Milliseconds() != 300 {
		t.Errorf("DefaultConfig().Rate = %v, want 300ms", cfg.Rate)
	}
}

func TestToIngredients(t *testing.T) {
	d := rawDrink{
		StrIngredient1: "Tequila",
		StrMeasure1:    "1 1/2 oz",
		StrIngredient2: "Triple sec",
		StrMeasure2:    "1/2 oz",
		StrIngredient3: "Lime juice",
		StrMeasure3:    "1 oz",
		StrIngredient4: "Salt",
		StrMeasure4:    "",
		// Ingredient5 empty — should stop here
	}
	ings := d.toIngredients()
	if len(ings) != 4 {
		t.Fatalf("toIngredients() len = %d, want 4", len(ings))
	}
	if ings[0].Name != "Tequila" {
		t.Errorf("ings[0].Name = %q, want Tequila", ings[0].Name)
	}
	if ings[0].Measure != "1 1/2 oz" {
		t.Errorf("ings[0].Measure = %q, want 1 1/2 oz", ings[0].Measure)
	}
	if ings[3].Name != "Salt" {
		t.Errorf("ings[3].Name = %q, want Salt", ings[3].Name)
	}
	if ings[3].Measure != "" {
		t.Errorf("ings[3].Measure = %q, want empty", ings[3].Measure)
	}
}

func TestToDrink(t *testing.T) {
	d := rawDrink{
		IDDrink:         "11007",
		StrDrink:        "Margarita",
		StrCategory:     "Ordinary Drink",
		StrAlcoholic:    "Alcoholic",
		StrGlass:        "Cocktail glass",
		StrInstructions: "Rub the rim.",
		StrDrinkThumb:   "https://example.com/img.jpg",
		StrIngredient1:  "Tequila",
		StrMeasure1:     "1 oz",
	}
	drink := toDrink(d)
	if drink.ID != "11007" {
		t.Errorf("ID = %q, want 11007", drink.ID)
	}
	if drink.Name != "Margarita" {
		t.Errorf("Name = %q, want Margarita", drink.Name)
	}
	if len(drink.Ingredients) != 1 {
		t.Errorf("Ingredients len = %d, want 1", len(drink.Ingredients))
	}
}
