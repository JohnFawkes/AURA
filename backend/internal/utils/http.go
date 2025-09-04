package utils

import (
	"aura/internal/config"
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
		urlTitle = config.Global.MediaServer.Type
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
		Err.HelpText = fmt.Sprintf("Error creating HTTP request (%s) [%s]", urlTitle, ElapsedTime(startTime))
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

	// Add headers to the request
	if tokenType == "MediaServer" {
		if strings.ToLower(config.Global.MediaServer.Type) == "plex" {
			req.Header.Set("X-Plex-Token", config.Global.MediaServer.Token)
		} else if strings.ToLower(config.Global.MediaServer.Type) == "emby" {
			req.Header.Set("X-Emby-Token", config.Global.MediaServer.Token)
		} else if strings.ToLower(config.Global.MediaServer.Type) == "jellyfin" {
			req.Header.Set("X-Emby-Token", config.Global.MediaServer.Token)
		}
	} else if strings.ToLower(tokenType) == "tmdb" {
		req.Header.Set("Authorization", "Bearer "+config.Global.TMDB.ApiKey)
	} else if strings.ToLower(tokenType) == "mediux" {
		req.Header.Set("Authorization", "Bearer "+config.Global.Mediux.Token)
	} else if strings.ToLower(tokenType) == "plex" {
		req.Header.Set("X-Plex-Token", config.Global.MediaServer.Token)
	} else if strings.ToLower(tokenType) == "emby" {
		req.Header.Set("X-Emby-Token", config.Global.MediaServer.Token)
	} else if strings.ToLower(tokenType) == "jellyfin" {
		req.Header.Set("X-Emby-Token", config.Global.MediaServer.Token)
	}

	if headers != nil {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
		logging.LOG.Trace("Added custom headers to request")
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
	req.Header.Set("Accept", "*/*")

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		Err.Message = "Failed to send HTTP request"
		Err.HelpText = fmt.Sprintf("Error sending HTTP request (%s) [%s]", urlTitle, ElapsedTime(startTime))
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
		Err.HelpText = fmt.Sprintf("Error reading HTTP response body (%s) [%s]", urlTitle, ElapsedTime(startTime))
		Err.Details = map[string]any{
			"method":       method,
			"url":          url,
			"error":        err.Error(),
			"responseBody": string(respBody),
		}
		return nil, nil, Err
	}

	// Make sure response status is OK
	// if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
	// 	Err.Message = fmt.Sprintf("Received non-OK HTTP response: %d", resp.StatusCode)
	// 	Err.HelpText = fmt.Sprintf("Ensure the server is running and accessible at the configured URL (%s) [%s]", urlTitle, ElapsedTime(startTime))
	// 	Err.Details = fmt.Sprintf("Response Status Code: %d, Response Body: %s", resp.StatusCode, string(respBody))
	// 	return nil, nil, Err
	// }

	// Defer closing the response body
	defer resp.Body.Close()
	logging.LOG.Trace(fmt.Sprintf("Sent HTTP request to %s [%s]", urlTitle, ElapsedTime(startTime)))
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
		Err.HelpText = fmt.Sprintf("Ensure the request body is a valid JSON object matching the expected structure for `%s` [%s]", structName, ElapsedTime(startTime))
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

func ValidateMediUXToken(token string) logging.StandardError {
	// Create a new StandardError instance
	Err := logging.NewStandardError()
	logging.LOG.Debug("Validating MediUX token")
	if token == "" {
		Err.Message = "MediUX token is empty"
		Err.HelpText = "Please provide a valid MediUX token in the configuration file"
		return Err
	}

	// Make a GET request to the MediUX API to validate the token
	url := "https://images.mediux.io/users/me"
	response, body, err := MakeHTTPRequest(url, "GET", nil, 10, nil, "Mediux")
	if err.Message != "" {
		Err.Message = "Failed to validate MediUX token"
		Err.HelpText = "Ensure the MediUX service is reachable and the token is valid."
		Err.Details = fmt.Sprintf("Error making request to MediUX API: %s", err.Message)
		return Err
	}

	if response.StatusCode != http.StatusOK {
		Err.Message = "Invalid MediUX token"
		Err.HelpText = "Ensure the MediUX token is correct and has not expired."
		Err.Details = map[string]any{
			"status_code": response.StatusCode,
			"response":    string(body),
		}
		return Err
	}

	logging.LOG.Info("Successfully validated MediUX token")
	return logging.StandardError{}
}
