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

func TestIngredientPairs(t *testing.T) {
	d := rawDrink{
		StrIngredient1: "Tequila",
		StrMeasure1:    "1 1/2 oz",
		StrIngredient2: "Triple sec",
		StrMeasure2:    "1/2 oz",
		StrIngredient3: "Lime juice",
		StrMeasure3:    "1 oz",
		StrIngredient4: "Salt",
		StrMeasure4:    "",
		// Ingredient5 empty -- should stop here
	}
	pairs := d.ingredientPairs()
	if len(pairs) != 4 {
		t.Fatalf("ingredientPairs() len = %d, want 4", len(pairs))
	}
	if pairs[0] != "Tequila:1 1/2 oz" {
		t.Errorf("pairs[0] = %q, want Tequila:1 1/2 oz", pairs[0])
	}
	if pairs[3] != "Salt" {
		t.Errorf("pairs[3] = %q, want Salt (no colon when measure empty)", pairs[3])
	}
}

func TestToCocktail(t *testing.T) {
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
	c := toCocktail(d)
	if c.ID != "11007" {
		t.Errorf("ID = %q, want 11007", c.ID)
	}
	if c.Name != "Margarita" {
		t.Errorf("Name = %q, want Margarita", c.Name)
	}
	if c.Ingredients != "Tequila:1 oz" {
		t.Errorf("Ingredients = %q, want Tequila:1 oz", c.Ingredients)
	}
	if c.Instructions != "Rub the rim." {
		t.Errorf("Instructions = %q, want Rub the rim.", c.Instructions)
	}
}

func TestToCocktailTruncatesInstructions(t *testing.T) {
	long := make([]byte, 250)
	for i := range long {
		long[i] = 'A'
	}
	d := rawDrink{
		IDDrink:         "x",
		StrInstructions: string(long),
	}
	c := toCocktail(d)
	if len(c.Instructions) != 203 {
		t.Errorf("Instructions len = %d, want 203 (200 + ...)", len(c.Instructions))
	}
	if c.Instructions[200:] != "..." {
		t.Errorf("Instructions suffix = %q, want ...", c.Instructions[200:])
	}
}
