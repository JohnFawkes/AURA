package api

import (
	"aura/internal/logging"
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
func MakeHTTPRequest(url, method string, headers map[string]string, timeout int, body []byte, tokenType string) (*http.Response, []byte, logging.StandardError) {
	startTime := time.Now()
	Err := logging.NewStandardError()
	var urlTitle string
	if tokenType == "MediaServer" {
		urlTitle = Global_Config.MediaServer.Type
	} else {
		urlTitle = getURLTitle(url)
	}

	// Create a context with a timeout
	timeoutInterval := time.Duration(timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeoutInterval)
	defer cancel()

	// Create a new request with context
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		Err.Message = "Failed to create HTTP request"
		Err.HelpText = fmt.Sprintf("Error creating HTTP request (%s) [%s]", urlTitle, Util_ElapsedTime(startTime))
		Err.Details = map[string]any{
			"method": method,
			"url":    url,
			"error":  err.Error(),
		}
		return nil, nil, Err
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

	if headers != nil {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
		if tokenType != "Sonarr" && tokenType != "Radarr" {
			logging.LOG.Trace("Added custom headers to request")
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
		Err.Message = "Failed to send HTTP request"
		Err.HelpText = fmt.Sprintf("Error sending HTTP request (%s) [%s]", urlTitle, Util_ElapsedTime(startTime))
		Err.Details = map[string]any{
			"method": method,
			"url":    url,
			"error":  err.Error(),
		}
		return nil, nil, Err
	}

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		Err.Message = "Failed to read HTTP response body"
		Err.HelpText = fmt.Sprintf("Error reading HTTP response body (%s) [%s]", urlTitle, Util_ElapsedTime(startTime))
		Err.Details = map[string]any{
			"method":       method,
			"url":          url,
			"error":        err.Error(),
			"responseBody": string(respBody),
		}
		return nil, nil, Err
	}
	defer resp.Body.Close()

	if tokenType != "Sonarr" && tokenType != "Radarr" {
		logging.LOG.Trace(fmt.Sprintf("Sent HTTP request to %s [%s]", urlTitle, Util_ElapsedTime(startTime)))
	}
	// Return the response
	return resp, respBody, logging.StandardError{}
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, v any, structName string, startTime time.Time) logging.StandardError {
	Err := logging.NewStandardError()
	logging.LOG.Debug(fmt.Sprintf("Decoding the request body into the `%s` struct", structName))
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(v)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to decode the request body into the `%s` struct --- `%s`", structName, err.Error())
		Err.Message = errorMsg
		Err.HelpText = fmt.Sprintf("Ensure the request body is a valid JSON object matching the expected structure for `%s` [%s]", structName, Util_ElapsedTime(startTime))
		Err.Details = fmt.Sprintf("Request Body: %s", r.Body)
		return Err
	}
	logging.LOG.Trace(fmt.Sprintf("Decoded the request body into the `%s` struct", structName))
	return logging.StandardError{}
}

func getURLTitle(rawURL string) string {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return ""
	}
	return parsedURL.Host
}
