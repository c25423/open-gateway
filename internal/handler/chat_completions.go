package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/c25423/open-gateway/internal/config"
	"github.com/gin-gonic/gin"
)

func NewChatCompletionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Transform
		req, err := transformRequest(c.Request.Header, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Execute request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upstream unreachable"})
			return
		}
		defer resp.Body.Close()

		// Mirror response
		for k, vv := range resp.Header {
			for _, v := range vv {
				c.Writer.Header().Add(k, v)
			}
		}
		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}

func transformRequest(incomingHeaders http.Header, incomingBody io.Reader) (*http.Request, error) {
	// Read and unmarshal body
	incomingBodyBytes, err := io.ReadAll(incomingBody)
	if err != nil {
		return nil, err
	}
	var bodyMap map[string]any
	if err := json.Unmarshal(incomingBodyBytes, &bodyMap); err != nil {
		return nil, err
	}

	// Get OAI provider config and model config using the incoming OAI identifier
	oaiIdentifier, ok := bodyMap["model"].(string)
	if !ok {
		return nil, fmt.Errorf("malformed oai identifier %q", oaiIdentifier)
	}
	oaiProviderConfig, oaiModelConfig, err := config.GetOaiConfigByOaiIdentifier(oaiIdentifier)
	if err != nil {
		return nil, err
	}

	// Build URL
	upstreamUrl := string(oaiProviderConfig.BaseUrl) + "/chat/completions"
	// Overwite body <- extra body in config
	for k, v := range oaiModelConfig.ExtraBody {
		bodyMap[k] = v
	}
	// Overwrite body <- "model"
	bodyMap["model"] = oaiModelConfig.Identifier
	// Marshal body
	upstreamBody, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, err
	}
	// Build upstream request
	req, err := http.NewRequest(http.MethodPost, upstreamUrl, bytes.NewReader(upstreamBody))
	if err != nil {
		return nil, err
	}
	// Add orginal headers
	for k, vv := range incomingHeaders {
		for _, v := range vv {
			req.Header.Set(k, v)
		}
	}
	// Overwrite headers <- extra headers in config
	for k, v := range oaiModelConfig.ExtraHeaders {
		req.Header.Set(k, v)
	}
	// Overwrite headers <- "Accept", "Authorization", and "Content-Type"
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", "Bearer "+string(oaiProviderConfig.ApiKey))
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
