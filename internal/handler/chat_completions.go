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

func transformRequest(incomingHeader http.Header, incomingBody io.Reader) (*http.Request, error) {
	// Read and unmarshall body
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

	// Build the upstream request
	upstreamUrl := string(oaiProviderConfig.BaseUrl) + "/chat/completions"
	for k, v := range oaiModelConfig.ExtraBody {
		bodyMap[k] = v
	}
	bodyMap["model"] = oaiModelConfig.Identifier
	upstreamBody, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, upstreamUrl, bytes.NewReader(upstreamBody))
	if err != nil {
		return nil, err
	}
	for k, v := range oaiModelConfig.ExtraHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", "Bearer "+string(oaiProviderConfig.ApiKey))
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
