package main

import (
	"GoRestAPI/recipes"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

func main() {
	// Create Gin router
	router := gin.Default()

	// Instantiate recipe Handler and provide a data store implementation
	store := recipes.NewMemStore()
	recipesHandler := NewRecipesHandler(store)

	// Register Routes
	router.GET("/", homePage)
	router.GET("/recipes", recipesHandler.ListRecipes)
	router.POST("/recipes", recipesHandler.CreateRecipe)
    router.GET("/recipes/:id", recipesHandler.GetRecipe)
    router.PUT("/recipes/:id", recipesHandler.UpdateRecipe)
    router.DELETE("/recipes/:id", recipesHandler.DeleteRecipe)

	// Start the server
	router.Run()
}

func homePage(c *gin.Context) {
	c.String(http.StatusOK, "This is my home page")
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

// Define handler function signatures
func (h RecipesHandler) CreateRecipe(c *gin.Context) {
	// Get request body and covert it to recipes.Recipe
	var recipe recipes.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert the name of the recipe into URL friendly string
	id := slug.Make(recipe.Name)

	// Add to store
	h.store.Add(id, recipe)

	// Return success payload
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h RecipesHandler) ListRecipes(c *gin.Context)  {
	// Call the store to get the list of recipes
	r, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	// Return the list, JSON encoding is implicit
	c.JSON(200, r)
}

func (h RecipesHandler) GetRecipe(c *gin.Context) {
	// Retrieve the URL parameter
	id := c.Param("id")

	// Retrieve recipe from the store
	recipe, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}

	c.JSON(200, recipe)
}

func (h RecipesHandler) UpdateRecipe(c *gin.Context) {
	// Retrieve the URL parameter
	id := c.Param("id")

	// Get request body and covert it to recipes.Recipe
	var recipe recipes.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the store to update the recipe
	if err := h.store.Update(id, recipe); err != nil {
		if err == recipes.NotFoundErr {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h RecipesHandler) DeleteRecipe(c *gin.Context) {
	// Retrieve the URL parameter
	id := c.Param("id")

	if err := h.store.Remove(id); err != nil {
		if err == recipes.NotFoundErr {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}