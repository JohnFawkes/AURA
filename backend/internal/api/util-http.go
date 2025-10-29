package api

import (
	"aura/internal/logging"
	"aura/internal/masking"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MakeHTTPRequest function to handle HTTP requests
func MakeHTTPRequest(ctx context.Context, url, method string, headers map[string]string, timeout int, body []byte, tokenType string) (*http.Response, []byte, logging.LogErrorInfo) {
	urlTitle := tokenType
	if urlTitle == "" {
		if tokenType == "MediaServer" {
			urlTitle = Global_Config.MediaServer.Type
		} else {
			urlTitle = getURLTitle(url)
		}
	}

	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Making %s request to %s", method, urlTitle), logging.LevelTrace)
	defer logAction.Complete()

	// Create a context with a timeout
	timeoutInterval := time.Duration(timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeoutInterval)
	defer cancel()

	// Create a new request with context
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		logAction.SetError("Failed to create HTTP request", err.Error(), map[string]any{
			"method": method,
			"url":    url,
			"error":  err.Error(),
		})
		return nil, nil, *logAction.Error
	}

	// Add a User-Agent header to the request
	req.Header.Set("User-Agent", "aura/1.0")
	req.Header.Set("X-Request", "mediux-aura")
	req.Header.Set("Accept", "application/json")

	// Add headers to the request
	if tokenType == "MediaServer" {
		if strings.ToLower(Global_Config.MediaServer.Type) == "plex" {
			req.Header.Set("X-Plex-Token", Global_Config.MediaServer.Token)
		} else if strings.ToLower(Global_Config.MediaServer.Type) == "emby" {
			req.Header.Set("X-Emby-Token", Global_Config.MediaServer.Token)
		} else if strings.ToLower(Global_Config.MediaServer.Type) == "jellyfin" {
			req.Header.Set("X-Emby-Token", Global_Config.MediaServer.Token)
		}
	} else if strings.ToLower(tokenType) == "tmdb" {
		req.Header.Set("Authorization", "Bearer "+Global_Config.TMDB.APIKey)
	} else if strings.ToLower(tokenType) == "mediux" {
		req.Header.Set("Authorization", "Bearer "+Global_Config.Mediux.Token)
	} else if strings.ToLower(tokenType) == "plex" {
		req.Header.Set("X-Plex-Token", Global_Config.MediaServer.Token)
	} else if strings.ToLower(tokenType) == "emby" {
		req.Header.Set("X-Emby-Token", Global_Config.MediaServer.Token)
	} else if strings.ToLower(tokenType) == "jellyfin" {
		req.Header.Set("X-Emby-Token", Global_Config.MediaServer.Token)
	}

	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Add(key, value)
			logAction.AppendResult("headers_added", map[string]any{key: masking.Masking_Token(value)})
		}
	}

	// Only set Content-Type to application/json if not already set
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Create a new HTTP client with both HTTP/1.1 and HTTP/2 support
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			ForceAttemptHTTP2: true, // Try HTTP/2 but fallback to HTTP/1.1 if needed
		},
		Timeout: timeoutInterval,
	}

	// Add common headers
	req.Header.Set("Connection", "keep-alive")

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		logAction.SetError("Failed to send HTTP request", "Check connection and try again", map[string]any{
			"method": method,
			"url":    url,
			"error":  err.Error(),
		})
		return nil, nil, *logAction.Error
	}

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		logAction.SetError("Failed to read HTTP response body", "Check response and try again", map[string]any{
			"method":      method,
			"url":         url,
			"error":       err.Error(),
			"status_code": resp.StatusCode,
		})
		return nil, nil, *logAction.Error
	}
	defer resp.Body.Close()

	// Add the Status Code the logging context Result
	logAction.AppendResult("status_code", resp.StatusCode)

	// Return the response
	return resp, respBody, logging.LogErrorInfo{}
}

func DecodeJSONBody(ctx context.Context, body []byte, v any, structName string) logging.LogErrorInfo {
	decoder := json.NewDecoder(bytes.NewReader(body))
	err := decoder.Decode(v)
	if err != nil {
		_, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Decoding request body into `%s` struct", structName), logging.LevelTrace)
		defer logAction.Complete()
		logAction.SetError("Failed to decode the JSON", fmt.Sprintf("Ensure that the JSON is correct for %s", structName),
			map[string]any{
				"requestBody": body,
				"error":       err.Error(),
			})
		return *logAction.Error
	}
	return logging.LogErrorInfo{}
}

func DecodeRequestBodyJSON(ctx context.Context, r io.ReadCloser, v any, structName string) logging.LogErrorInfo {
	body, err := io.ReadAll(r)
	if err != nil {
		_, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Decoding request body into `%s` struct", structName), logging.LevelTrace)
		defer logAction.Complete()
		logAction.SetError("Failed to read request body", fmt.Sprintf("Ensure that the request body can be read for %s", structName),
			map[string]any{
				"error": err.Error(),
			})
		return *logAction.Error
	}
	defer r.Close()

	decoder := json.NewDecoder(bytes.NewReader(body))
	err = decoder.Decode(v)
	if err != nil {
		_, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Decoding request body into `%s` struct", structName), logging.LevelTrace)
		defer logAction.Complete()
		logAction.SetError("Failed to decode the JSON", fmt.Sprintf("Ensure that the JSON is correct for %s", structName),
			map[string]any{
				"requestBody": body,
				"error":       err.Error(),
			})
		return *logAction.Error
	}
	return logging.LogErrorInfo{}
}

func getURLTitle(rawURL string) string {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return ""
	}
	return parsedURL.Host
}
