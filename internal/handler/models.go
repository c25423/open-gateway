package handler

import (
	"log"
	"net/http"

	"github.com/c25423/open-gateway/internal/config"
	"github.com/gin-gonic/gin"
)

// OpenAI Models Response Structure
type ModelObject struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type ModelsResponse struct {
	Object string        `json:"object"`
	Data   []ModelObject `json:"data"`
}

// Prebuilt models response to avoid reconstruction on each request
var prebuiltModelsResponse *ModelsResponse

func NewModelsHandler() gin.HandlerFunc {
	models, err := config.GetOaiModels()
	if err != nil {
		models = []string{}
	}

	// Convert to OpenAI format
	modelObjects := make([]ModelObject, 0, len(models))
	for _, modelName := range models {
		modelObjects = append(modelObjects, ModelObject{
			ID:      modelName,
			Object:  "model",
			Created: 0,
			OwnedBy: "open-gateway",
		})
	}

	// Construct prebuilt response
	prebuiltModelsResponse = &ModelsResponse{
		Object: "list",
		Data:   modelObjects,
	}

	log.Println("Prebuilt models response")

	return func(c *gin.Context) {
		// Return the prebuilt response
		if prebuiltModelsResponse != nil {
			c.JSON(http.StatusOK, prebuiltModelsResponse)
		} else {
			// Fallback in case of initialization error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize models"})
		}
	}
}
