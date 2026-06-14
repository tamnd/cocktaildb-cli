package cocktaildb

import "strings"

// Ingredient is one cocktail ingredient with its measure.
type Ingredient struct {
	Name    string `json:"name"`
	Measure string `json:"measure"`
}

// Drink is one drink from TheCocktailDB.
type Drink struct {
	ID           string       `kit:"id" json:"id"`
	Name         string       `json:"name"`
	Category     string       `json:"category"`
	Alcoholic    string       `json:"alcoholic"`
	Glass        string       `json:"glass"`
	Instructions string       `json:"instructions"`
	Thumbnail    string       `json:"thumbnail"`
	Ingredients  []Ingredient `json:"ingredients"`
}

// Category is one list entry from TheCocktailDB (category, alcoholic type, glass, or ingredient).
type Category struct {
	Name string `kit:"id" json:"name"`
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
	// Ingredients 1-15
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
	// Measures 1-15
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

// listResponse handles all list.php responses where each entry is a map.
type listResponse struct {
	Drinks []map[string]string `json:"drinks"`
}

// toIngredients converts the flat strIngredientN / strMeasureN fields into
// a clean []Ingredient. Slots are filled consecutively; the loop stops on
// the first empty name.
func (r rawDrink) toIngredients() []Ingredient {
	names := [15]string{
		r.StrIngredient1, r.StrIngredient2, r.StrIngredient3,
		r.StrIngredient4, r.StrIngredient5, r.StrIngredient6,
		r.StrIngredient7, r.StrIngredient8, r.StrIngredient9,
		r.StrIngredient10, r.StrIngredient11, r.StrIngredient12,
		r.StrIngredient13, r.StrIngredient14, r.StrIngredient15,
	}
	measures := [15]string{
		r.StrMeasure1, r.StrMeasure2, r.StrMeasure3,
		r.StrMeasure4, r.StrMeasure5, r.StrMeasure6,
		r.StrMeasure7, r.StrMeasure8, r.StrMeasure9,
		r.StrMeasure10, r.StrMeasure11, r.StrMeasure12,
		r.StrMeasure13, r.StrMeasure14, r.StrMeasure15,
	}
	var ings []Ingredient
	for i, name := range names {
		if strings.TrimSpace(name) == "" {
			break
		}
		ings = append(ings, Ingredient{
			Name:    strings.TrimSpace(name),
			Measure: strings.TrimSpace(measures[i]),
		})
	}
	return ings
}

// toDrink converts a rawDrink to a Drink.
func toDrink(d rawDrink) Drink {
	return Drink{
		ID:           d.IDDrink,
		Name:         d.StrDrink,
		Category:     d.StrCategory,
		Alcoholic:    d.StrAlcoholic,
		Glass:        d.StrGlass,
		Instructions: d.StrInstructions,
		Thumbnail:    d.StrDrinkThumb,
		Ingredients:  d.toIngredients(),
	}
}
