package recipes

// Represents a recipe
type Recipe struct {
	Name 		string		  `json:"name"`
	Ingredients []Ingredient  `json: "ingredients"`
}

// Represents individual ingredients
type Ingredient struct {
	Name string `json: "name"`
}
