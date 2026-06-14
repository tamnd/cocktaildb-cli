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
  {"idDrink":"11007","strDrink":"Margarita","strCategory":"Ordinary Drink","strAlcoholic":"Alcoholic","strGlass":"Cocktail glass","strInstructions":"Rub the rim with lime.","strDrinkThumb":"https://www.thecocktaildb.com/images/media/drink/5noda61589575158.jpg","strIngredient1":"Tequila","strMeasure1":"1 1/2 oz","strIngredient2":"Triple sec","strMeasure2":"1/2 oz","strIngredient3":"Lime juice","strMeasure3":"1 oz","strIngredient4":"Salt","strMeasure4":"","strIngredient5":"","strMeasure5":"","strIngredient6":"","strMeasure6":"","strIngredient7":"","strMeasure7":"","strIngredient8":"","strMeasure8":"","strIngredient9":"","strMeasure9":"","strIngredient10":"","strMeasure10":"","strIngredient11":"","strMeasure11":"","strIngredient12":"","strMeasure12":"","strIngredient13":"","strMeasure13":"","strIngredient14":"","strMeasure14":"","strIngredient15":"","strMeasure15":""},
  {"idDrink":"11118","strDrink":"Blue Margarita","strCategory":"Ordinary Drink","strAlcoholic":"Alcoholic","strGlass":"Cocktail glass","strInstructions":"Rub rim of cocktail glass with lime juice.","strDrinkThumb":"https://www.thecocktaildb.com/images/media/drink/qtvvyq1439905913.jpg","strIngredient1":"Tequila","strMeasure1":"1 1/2 oz","strIngredient2":"Blue Curacao","strMeasure2":"1 oz","strIngredient3":"Lime juice","strMeasure3":"1 oz","strIngredient4":"Salt","strMeasure4":"","strIngredient5":"","strMeasure5":"","strIngredient6":"","strMeasure6":"","strIngredient7":"","strMeasure7":"","strIngredient8":"","strMeasure8":"","strIngredient9":"","strMeasure9":"","strIngredient10":"","strMeasure10":"","strIngredient11":"","strMeasure11":"","strIngredient12":"","strMeasure12":"","strIngredient13":"","strMeasure13":"","strIngredient14":"","strMeasure14":"","strIngredient15":"","strMeasure15":""}
]}`

const fakeRandomJSON = `{"drinks":[
  {"idDrink":"11007","strDrink":"Margarita","strCategory":"Ordinary Drink","strAlcoholic":"Alcoholic","strGlass":"Cocktail glass","strInstructions":"Rub the rim with lime.","strDrinkThumb":"https://www.thecocktaildb.com/images/media/drink/5noda61589575158.jpg","strIngredient1":"Tequila","strMeasure1":"1 1/2 oz","strIngredient2":"Triple sec","strMeasure2":"1/2 oz","strIngredient3":"Lime juice","strMeasure3":"1 oz","strIngredient4":"Salt","strMeasure4":"","strIngredient5":"","strMeasure5":"","strIngredient6":"","strMeasure6":"","strIngredient7":"","strMeasure7":"","strIngredient8":"","strMeasure8":"","strIngredient9":"","strMeasure9":"","strIngredient10":"","strMeasure10":"","strIngredient11":"","strMeasure11":"","strIngredient12":"","strMeasure12":"","strIngredient13":"","strMeasure13":"","strIngredient14":"","strMeasure14":"","strIngredient15":"","strMeasure15":""}
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

func TestRandomParsesCocktail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeRandomJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	got, err := c.Random(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "11007" {
		t.Errorf("ID = %q, want 11007", got.ID)
	}
	if got.Name != "Margarita" {
		t.Errorf("Name = %q, want Margarita", got.Name)
	}
	if got.Rank != 1 {
		t.Errorf("Rank = %d, want 1", got.Rank)
	}
	if len(got.Ingredients) != 4 {
		t.Errorf("len(Ingredients) = %d, want 4", len(got.Ingredients))
	}
}
