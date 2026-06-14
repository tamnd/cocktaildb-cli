package cocktaildb_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/cocktaildb-cli/cocktaildb"
)

const fakeSearchJSON = `{"drinks":[
  {"idDrink":"11007","strDrink":"Margarita","strCategory":"Ordinary Drink","strAlcoholic":"Alcoholic","strGlass":"Cocktail glass","strInstructions":"Rub the rim with lime.","strDrinkThumb":"https://www.thecocktaildb.com/images/media/drink/5noda61589575158.jpg","strIngredient1":"Tequila","strMeasure1":"1 1/2 oz","strIngredient2":"Triple sec","strMeasure2":"1/2 oz","strIngredient3":"Lime juice","strMeasure3":"1 oz","strIngredient4":"Salt","strMeasure4":"","strIngredient5":"","strMeasure5":""},
  {"idDrink":"11118","strDrink":"Blue Margarita","strCategory":"Ordinary Drink","strAlcoholic":"Alcoholic","strGlass":"Cocktail glass","strInstructions":"Rub rim of cocktail glass with lime juice.","strDrinkThumb":"https://www.thecocktaildb.com/images/media/drink/qtvvyq1439905913.jpg","strIngredient1":"Tequila","strMeasure1":"1 1/2 oz","strIngredient2":"Blue Curacao","strMeasure2":"1 oz","strIngredient3":"Lime juice","strMeasure3":"1 oz","strIngredient4":"Salt","strMeasure4":"","strIngredient5":"","strMeasure5":""}
]}`

const fakeLookupJSON = `{"drinks":[
  {"idDrink":"11007","strDrink":"Margarita","strCategory":"Ordinary Drink","strAlcoholic":"Alcoholic","strGlass":"Cocktail glass","strInstructions":"Rub the rim with lime.","strDrinkThumb":"https://www.thecocktaildb.com/images/media/drink/5noda61589575158.jpg","strIngredient1":"Tequila","strMeasure1":"1 1/2 oz","strIngredient2":"Triple sec","strMeasure2":"1/2 oz","strIngredient3":"Lime juice","strMeasure3":"1 oz","strIngredient4":"Salt","strMeasure4":"","strIngredient5":"","strMeasure5":""}
]}`

const fakeRandomJSON = `{"drinks":[
  {"idDrink":"17222","strDrink":"Pisco Sour","strCategory":"Ordinary Drink","strAlcoholic":"Alcoholic","strGlass":"Whiskey sour glass","strInstructions":"Shake and strain into a chilled glass.","strDrinkThumb":"https://www.thecocktaildb.com/images/media/drink/tsssur1439907622.jpg","strIngredient1":"Pisco","strMeasure1":"2 oz","strIngredient2":"Lemon juice","strMeasure2":"1 oz","strIngredient3":"Syrup","strMeasure3":"1/2 oz","strIngredient4":"","strMeasure4":""}
]}`

const fakeCategoriesJSON = `{"drinks":[
  {"strCategory":"Beer"},
  {"strCategory":"Cocktail"},
  {"strCategory":"Ordinary Drink"},
  {"strCategory":"Shot"}
]}`

const fakeAlcoholicJSON = `{"drinks":[
  {"strAlcoholic":"Alcoholic"},
  {"strAlcoholic":"Non alcoholic"},
  {"strAlcoholic":"Optional alcohol"}
]}`

const fakeGlassJSON = `{"drinks":[
  {"strGlass":"Highball glass"},
  {"strGlass":"Cocktail glass"},
  {"strGlass":"Old-fashioned glass"}
]}`

func newTestClient(ts *httptest.Server) *cocktaildb.Client {
	cfg := cocktaildb.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return cocktaildb.NewClient(cfg)
}

func TestSearchSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeSearchJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Search(context.Background(), "margarita", 0)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent header not sent")
	}
}

func TestSearchParsesItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeSearchJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Search(context.Background(), "margarita", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	got := items[0]
	if got.Name != "Margarita" {
		t.Errorf("items[0].Name = %q, want Margarita", got.Name)
	}
	if got.Category != "Ordinary Drink" {
		t.Errorf("items[0].Category = %q, want Ordinary Drink", got.Category)
	}
	if len(got.Ingredients) != 4 {
		t.Errorf("len(items[0].Ingredients) = %d, want 4", len(got.Ingredients))
	}
	if got.Ingredients[0].Name != "Tequila" {
		t.Errorf("items[0].Ingredients[0].Name = %q, want Tequila", got.Ingredients[0].Name)
	}
	if got.Ingredients[0].Measure != "1 1/2 oz" {
		t.Errorf("items[0].Ingredients[0].Measure = %q, want 1 1/2 oz", got.Ingredients[0].Measure)
	}
	if items[1].Name != "Blue Margarita" {
		t.Errorf("items[1].Name = %q, want Blue Margarita", items[1].Name)
	}
}

func TestSearchLimitRespected(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeSearchJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Search(context.Background(), "margarita", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d, want 1", len(items))
	}
}

func TestSearchRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeSearchJSON)
	}))
	defer ts.Close()

	cfg := cocktaildb.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := cocktaildb.NewClient(cfg)

	_, err := c.Search(context.Background(), "margarita", 0)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

func TestLookupByID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("i") != "11007" {
			t.Errorf("expected i=11007, got %q", r.URL.Query().Get("i"))
		}
		_, _ = fmt.Fprint(w, fakeLookupJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	got, err := c.Lookup(context.Background(), "11007")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "11007" {
		t.Errorf("ID = %q, want 11007", got.ID)
	}
	if got.Name != "Margarita" {
		t.Errorf("Name = %q, want Margarita", got.Name)
	}
}

func TestLookupNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"drinks":null}`)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Lookup(context.Background(), "99999999")
	if err == nil {
		t.Error("expected error for not-found ID, got nil")
	}
}

func TestRandomParsesDrink(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeRandomJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	got, err := c.Random(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "17222" {
		t.Errorf("ID = %q, want 17222", got.ID)
	}
	if got.Name != "Pisco Sour" {
		t.Errorf("Name = %q, want Pisco Sour", got.Name)
	}
	if len(got.Ingredients) != 3 {
		t.Errorf("len(Ingredients) = %d, want 3", len(got.Ingredients))
	}
}

func TestRandomEmptyResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"drinks":null}`)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Random(context.Background())
	if err == nil {
		t.Error("expected error for empty random response, got nil")
	}
}

func TestListCategoriesCategories(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("c") != "list" {
			t.Errorf("expected c=list param, got %q", r.URL.RawQuery)
		}
		_, _ = fmt.Fprint(w, fakeCategoriesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	cats, err := c.ListCategories(context.Background(), "categories")
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 4 {
		t.Fatalf("len(cats) = %d, want 4", len(cats))
	}
	if cats[0].Name != "Beer" {
		t.Errorf("cats[0].Name = %q, want Beer", cats[0].Name)
	}
}

func TestListCategoriesAlcoholic(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("a") != "list" {
			t.Errorf("expected a=list param, got %q", r.URL.RawQuery)
		}
		_, _ = fmt.Fprint(w, fakeAlcoholicJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	cats, err := c.ListCategories(context.Background(), "alcoholic")
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 3 {
		t.Fatalf("len(cats) = %d, want 3", len(cats))
	}
	if cats[0].Name != "Alcoholic" {
		t.Errorf("cats[0].Name = %q, want Alcoholic", cats[0].Name)
	}
}

func TestListCategoriesGlass(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("g") != "list" {
			t.Errorf("expected g=list param, got %q", r.URL.RawQuery)
		}
		_, _ = fmt.Fprint(w, fakeGlassJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	cats, err := c.ListCategories(context.Background(), "glass")
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 3 {
		t.Fatalf("len(cats) = %d, want 3", len(cats))
	}
	if cats[0].Name != "Highball glass" {
		t.Errorf("cats[0].Name = %q, want Highball glass", cats[0].Name)
	}
}

func TestSearchQuerySentAsParam(t *testing.T) {
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("s")
		_, _ = fmt.Fprint(w, fakeSearchJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Search(context.Background(), "margarita", 0)
	if err != nil {
		t.Fatal(err)
	}
	if gotQuery != "margarita" {
		t.Errorf("search query param s = %q, want margarita", gotQuery)
	}
}

func TestSearchNoResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"drinks":null}`)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Search(context.Background(), "xyzzy_notreal", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("len(items) = %d, want 0", len(items))
	}
}
