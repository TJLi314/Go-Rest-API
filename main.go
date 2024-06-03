package main

import (
	"encoding/json"
	"net/http"
	"regexp"

	"Standard-Library-Rest-API/recipes"

	"github.com/gosimple/slug"
)

func main() {
	// Create the Store and Recipe Handler
	store := recipes.NewMemStore()
	recipeHandler := NewRecipesHandler(store)

	// Create a new request multiplexer
	// Take incoming requests and dispatch them to the matching handlers
	mux := http.NewServeMux()

	// Register the routes and handlers
	mux.Handle("/", &homeHandler{})
	mux.Handle("/recipes", recipeHandler)
	mux.Handle("/recipes/", recipeHandler)

	// Run the server
	http.ListenAndServe(":8080", mux)
}

type homeHandler struct{}

func (h *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is my home page"))
}

type recipeStore interface {
	Add(name string, recipe recipes.Recipe) error
	Get(name string) (recipes.Recipe, error)
	Update(name string, recipe recipes.Recipe) error
	List() (map[string]recipes.Recipe, error)
	Remove(name string) error
}

type RecipesHandler struct {
	store recipeStore
}

func NewRecipesHandler(s recipeStore) *RecipesHandler {
	return &RecipesHandler {
		store: s,
	}
}

var RecipeRe = regexp.MustCompile(`^/recipes/*$`)
var RecipeReWithID = regexp.MustCompile(`^/recipes/([a-z0-9]+(?:-[a-z0-9]+)+)$`)

func (h *RecipesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && RecipeRe.MatchString(r.URL.Path):
		h.CreateRecipe(w, r)
		return
	case r.Method == http.MethodGet && RecipeRe.MatchString(r.URL.Path):
        h.ListRecipes(w, r)
        return
    case r.Method == http.MethodGet && RecipeReWithID.MatchString(r.URL.Path):
        h.GetRecipe(w, r)
        return
    case r.Method == http.MethodPut && RecipeReWithID.MatchString(r.URL.Path):
        h.UpdateRecipe(w, r)
        return
    case r.Method == http.MethodDelete && RecipeReWithID.MatchString(r.URL.Path):
        h.DeleteRecipe(w, r)
        return
    default:
        return 
	}
}

func (h *RecipesHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	// Recipe object that will be populated from JSON payload
	var recipe recipes.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	// Convert the name of the recipe into URL friendly string
	resourceID := slug.Make(recipe.Name)
	
	// Call the store to add the recipe
	if err := h.store.Add(resourceID, recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}		

	// Set the status code to 200
	w.WriteHeader(http.StatusOK)
}

func (h *RecipesHandler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	resources, err := h.store.List()

	jsonBytes, err := json.Marshal(resources)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *RecipesHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	// Extract the resource ID/slug using a regex
	matches := RecipeReWithID.FindStringSubmatch(r.URL.Path)

	// Expect matches to be length >= 2 (full string + 1 matching group)
    if len(matches) < 2 {
        InternalServerErrorHandler(w, r)
        return
    }

	// Retrieve recipe from the store
	recipe, err := h.store.Get(matches[1])
	if err != nil {
		// Special case of NotFound Error
		if err == recipes.NotFoundErr {
			NotFoundHandler(w, r)
			return
		}
		
		InternalServerErrorHandler(w, r)
		return
	}

	// Convert to JSON payload and write results
	jsonBytes, err := json.Marshal(recipe)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *RecipesHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	// Extract the resource ID/slug using a regex
	matches := RecipeReWithID.FindStringSubmatch(r.URL.Path)
    if len(matches) < 2 {
        InternalServerErrorHandler(w, r)
        return
    }
	
	// Recipe object that will be populated from JSON payload
	var recipe recipes.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	
	// Call the store to update the recipe
	if err := h.store.Update(matches[1], recipe); err != nil {
		if err == recipes.NotFoundErr {
            NotFoundHandler(w, r)
            return
        }
        InternalServerErrorHandler(w, r)
        return
	}

	w.WriteHeader(http.StatusOK)

}

func (h *RecipesHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	// Extract the resource ID/slug using a regex
	matches := RecipeReWithID.FindStringSubmatch(r.URL.Path)
    if len(matches) < 2 {
        InternalServerErrorHandler(w, r)
        return
    }

	if err := h.store.Remove(matches[1]); err != nil {
		InternalServerErrorHandler(w, r)
        return
	}

	w.WriteHeader(http.StatusOK)
}

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 Internal Server Error"))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 Not Found"))
}