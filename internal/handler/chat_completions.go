package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/c25423/open-gateway/internal/config"
	"github.com/gin-gonic/gin"
)

func NewChatCompletionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		req, isStream, err := transformRequest(c.Request.Context(), c.Request.Header, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Execute request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return // On client disconnected
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upstream unreachable"})
			return
		}
		defer resp.Body.Close()

		// Process response headers
		for k, vv := range resp.Header {
			for _, v := range vv {
				c.Writer.Header().Add(k, v)
			}
		}
		c.Status(resp.StatusCode)

		// Process response body
		if isStream {
			// log.Println("Handling streamed response")
			// Use the manual flushing loop for streamed responses
			flusher, ok := c.Writer.(http.Flusher)
			if !ok {
				log.Println("Flusher unsupported: ResponseWriter does not implement http.Flusher")
				io.Copy(c.Writer, resp.Body) // Fallback to io.Copy
				return
			}
			buf := make([]byte, 4096)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					if _, writeErr := c.Writer.Write(buf[:n]); writeErr != nil {
						break // On client disconnected
					}
					flusher.Flush()
				}
				if err != nil {
					break // On EOF or other error
				}
			}
		} else {
			// log.Println("Handling non-streamed response")
			// Use io.Copy for non-streamed responses
			io.Copy(c.Writer, resp.Body)
		}
	}
}

func transformRequest(ctx context.Context, incomingHeaders http.Header, incomingBody io.Reader) (*http.Request, bool, error) {
	// Read and unmarshal body
	incomingBodyBytes, err := io.ReadAll(incomingBody)
	if err != nil {
		return nil, false, err
	}
	var bodyMap map[string]any
	if err := json.Unmarshal(incomingBodyBytes, &bodyMap); err != nil {
		return nil, false, err
	}

	// Check if requesting streamed response
	var isStream bool
	if streamVal, ok := bodyMap["stream"]; ok {
		if b, isBool := streamVal.(bool); isBool {
			isStream = b
		}
	}

	// Get OAI provider config and model config using the incoming OAI identifier
	oaiIdentifier, ok := bodyMap["model"].(string)
	if !ok {
		return nil, false, fmt.Errorf("malformed oai identifier %q", oaiIdentifier)
	}
	oaiProviderConfig, oaiModelConfig, err := config.GetOaiConfigByOaiIdentifier(oaiIdentifier)
	if err != nil {
		return nil, false, err
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
		return nil, false, err
	}
	// Build upstream request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, upstreamUrl, bytes.NewReader(upstreamBody))
	if err != nil {
		return nil, false, err
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
	// Overwrite headers <- "Accept", "Accept-Encoding", "Authorization", and "Content-Type"
	req.Header.Set("Accept", "*/*")
	req.Header.Del("Accept-Encoding") // Remove to avoid compression compatibility issues
	req.Header.Set("Authorization", "Bearer "+string(oaiProviderConfig.ApiKey))
	req.Header.Set("Content-Type", "application/json")

	return req, isStream, nil
}
