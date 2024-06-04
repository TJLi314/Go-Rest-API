package main

import (
	"GoRestAPI/recipes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gosimple/slug"
)

func main() {
	// Create the Store and Recipe Handler
	store := recipes.NewMemStore()
	recipesHandler := NewRecipesHandler(store)
	home := homeHandler{}

	// create the router
	router := mux.NewRouter()

	// Register the routes
	router.HandleFunc("/", home.ServeHTTP)
	router.HandleFunc("/recipes", recipesHandler.ListRecipes).Methods("GET")
	router.HandleFunc("/recipes", recipesHandler.CreateRecipe).Methods("POST")
	router.HandleFunc("/recipes/{id}", recipesHandler.GetRecipe).Methods("GET")
	router.HandleFunc("/recipes/{id}", recipesHandler.UpdateRecipe).Methods("PUT")
    router.HandleFunc("/recipes/{id}", recipesHandler.DeleteRecipe).Methods("DELETE")

	// Start the server
	http.ListenAndServe(":8010", router)
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

func NewRecipesHandler(s recipeStore) RecipesHandler {
	return RecipesHandler {
		store: s,
	}
}

func (h RecipesHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
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

func (h RecipesHandler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	resources, err := h.store.List()

	jsonBytes, err := json.Marshal(resources)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h RecipesHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	// Extract the resource ID/slug using mux func
	id := mux.Vars(r)["id"]

	// Retrieve recipe from the store
	recipe, err := h.store.Get(id)
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

func (h RecipesHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	// Extract the resource ID/slug using mux func
	id := mux.Vars(r)["id"]
	
	// Recipe object that will be populated from JSON payload
	var recipe recipes.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	
	// Call the store to update the recipe
	if err := h.store.Update(id, recipe); err != nil {
		if err == recipes.NotFoundErr {
            NotFoundHandler(w, r)
            return
        }
        InternalServerErrorHandler(w, r)
        return
	}

	w.WriteHeader(http.StatusOK)

}

func (h RecipesHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	// Extract the resource ID/slug using mux func
	id := mux.Vars(r)["id"]

	if err := h.store.Remove(id); err != nil {
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