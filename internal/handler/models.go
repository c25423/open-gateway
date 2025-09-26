package handler

import (
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

func NewModelsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get models from config
		models, err := config.GetOaiModels()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve models"})
			return
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

		// Construct response
		response := ModelsResponse{
			Object: "list",
			Data:   modelObjects,
		}

		c.JSON(http.StatusOK, response)
	}
}
