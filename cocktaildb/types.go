package cocktaildb

import "strings"

// Ingredient is one cocktail ingredient with its measure.
type Ingredient struct {
	Name    string `json:"name"`
	Measure string `json:"measure"`
}

// Cocktail is one drink from TheCocktailDB.
type Cocktail struct {
	Rank         int          `json:"rank"`
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Category     string       `json:"category"`
	Alcoholic    string       `json:"alcoholic"` // "Alcoholic", "Non alcoholic", "Optional alcohol"
	Glass        string       `json:"glass"`
	Instructions string       `json:"instructions"`
	Ingredients  []Ingredient `json:"ingredients"`
	Thumbnail    string       `json:"thumbnail"` // strDrinkThumb URL
}

// Category is one drink category from TheCocktailDB.
type Category struct {
	Rank int    `json:"rank"`
	Name string `json:"name"`
}

// GlassType is one glass type from TheCocktailDB.
type GlassType struct {
	Rank int    `json:"rank"`
	Name string `json:"name"`
}

// FilterResult is a summary drink object returned by the filter endpoint.
// It does not include ingredients; use Get to fetch full details.
type FilterResult struct {
	Rank      int    `json:"rank"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
}

// --- wire types (unexported, only used for JSON decoding) ---

type rawDrink struct {
	IDDrink         string `json:"idDrink"`
	StrDrink        string `json:"strDrink"`
	StrCategory     string `json:"strCategory"`
	StrAlcoholic    string `json:"strAlcoholic"`
	StrGlass        string `json:"strGlass"`
	StrInstructions string `json:"strInstructions"`
	StrDrinkThumb   string `json:"strDrinkThumb"`
	// Ingredients 1–15
	StrIngredient1  string `json:"strIngredient1"`
	StrIngredient2  string `json:"strIngredient2"`
	StrIngredient3  string `json:"strIngredient3"`
	StrIngredient4  string `json:"strIngredient4"`
	StrIngredient5  string `json:"strIngredient5"`
	StrIngredient6  string `json:"strIngredient6"`
	StrIngredient7  string `json:"strIngredient7"`
	StrIngredient8  string `json:"strIngredient8"`
	StrIngredient9  string `json:"strIngredient9"`
	StrIngredient10 string `json:"strIngredient10"`
	StrIngredient11 string `json:"strIngredient11"`
	StrIngredient12 string `json:"strIngredient12"`
	StrIngredient13 string `json:"strIngredient13"`
	StrIngredient14 string `json:"strIngredient14"`
	StrIngredient15 string `json:"strIngredient15"`
	// Measures 1–15
	StrMeasure1  string `json:"strMeasure1"`
	StrMeasure2  string `json:"strMeasure2"`
	StrMeasure3  string `json:"strMeasure3"`
	StrMeasure4  string `json:"strMeasure4"`
	StrMeasure5  string `json:"strMeasure5"`
	StrMeasure6  string `json:"strMeasure6"`
	StrMeasure7  string `json:"strMeasure7"`
	StrMeasure8  string `json:"strMeasure8"`
	StrMeasure9  string `json:"strMeasure9"`
	StrMeasure10 string `json:"strMeasure10"`
	StrMeasure11 string `json:"strMeasure11"`
	StrMeasure12 string `json:"strMeasure12"`
	StrMeasure13 string `json:"strMeasure13"`
	StrMeasure14 string `json:"strMeasure14"`
	StrMeasure15 string `json:"strMeasure15"`
}

type drinksResponse struct {
	Drinks []rawDrink `json:"drinks"`
}

type categoriesResponse struct {
	Drinks []struct {
		StrCategory string `json:"strCategory"`
	} `json:"drinks"`
}

type glassesResponse struct {
	Drinks []struct {
		StrGlass string `json:"strGlass"`
	} `json:"drinks"`
}

type filterDrink struct {
	IDDrink       string `json:"idDrink"`
	StrDrink      string `json:"strDrink"`
	StrDrinkThumb string `json:"strDrinkThumb"`
}

type filterResponse struct {
	Drinks []filterDrink `json:"drinks"`
}

// parseIngredients converts the flat strIngredientN / strMeasureN fields of a
// rawDrink into a clean []Ingredient. Slots are filled consecutively: the loop
// stops on the first empty name.
func parseIngredients(d rawDrink) []Ingredient {
	names := [15]string{
		d.StrIngredient1, d.StrIngredient2, d.StrIngredient3,
		d.StrIngredient4, d.StrIngredient5, d.StrIngredient6,
		d.StrIngredient7, d.StrIngredient8, d.StrIngredient9,
		d.StrIngredient10, d.StrIngredient11, d.StrIngredient12,
		d.StrIngredient13, d.StrIngredient14, d.StrIngredient15,
	}
	measures := [15]string{
		d.StrMeasure1, d.StrMeasure2, d.StrMeasure3,
		d.StrMeasure4, d.StrMeasure5, d.StrMeasure6,
		d.StrMeasure7, d.StrMeasure8, d.StrMeasure9,
		d.StrMeasure10, d.StrMeasure11, d.StrMeasure12,
		d.StrMeasure13, d.StrMeasure14, d.StrMeasure15,
	}
	var ings []Ingredient
	for i, name := range names {
		if name == "" {
			break
		}
		ings = append(ings, Ingredient{
			Name:    strings.TrimSpace(name),
			Measure: strings.TrimSpace(measures[i]),
		})
	}
	return ings
}

// toCocktail converts a rawDrink to a Cocktail, assigning the given rank.
func toCocktail(d rawDrink, rank int) Cocktail {
	return Cocktail{
		Rank:         rank,
		ID:           d.IDDrink,
		Name:         d.StrDrink,
		Category:     d.StrCategory,
		Alcoholic:    d.StrAlcoholic,
		Glass:        d.StrGlass,
		Instructions: d.StrInstructions,
		Ingredients:  parseIngredients(d),
		Thumbnail:    d.StrDrinkThumb,
	}
}
