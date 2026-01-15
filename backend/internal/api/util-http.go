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
	"strings"
	"time"
)

// MakeHTTPRequest function to handle HTTP requests
func MakeHTTPRequest(ctx context.Context, url, method string, headers map[string]string, timeout int, body []byte, siteName string) (*http.Response, []byte, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Making %s request to %s", method, siteName), logging.LevelTrace)
	defer logAction.Complete()

	// Create a context with a timeout
	timeoutInterval := time.Duration(timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeoutInterval)
	defer cancel()

	// Create a new request with context
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		logAction.SetError(fmt.Sprintf("Failed to create %s request to %s", method, siteName),
			"Check error and try again",
			map[string]any{
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

	possibleSensitiveHeaders := []string{
		"authorization",
		"api-key",
		"x-api-key",
		"x-auth-token",
		"x-access-token",
		"access-token",
		"token",
		"authentication",
	}

	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Add(key, value)
			maskedValue := value
			lowerKey := strings.ToLower(key)
			for _, sensitive := range possibleSensitiveHeaders {
				if strings.Contains(lowerKey, sensitive) {
					maskedValue = masking.Masking_Token(value)
					break
				}
			}
			logAction.AppendResult("headers_added", map[string]any{key: maskedValue})
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
		logAction.SetError(fmt.Sprintf("Failed to send %s request to %s", method, siteName),
			"Check error and try again",
			map[string]any{
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
		logAction.SetError(fmt.Sprintf("Failed to read response body from %s", siteName),
			"Check error and try again",
			map[string]any{
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

func MakeAuthHeader(key string, value string) map[string]string {
	headers := make(map[string]string)
	if strings.ToLower(key) == "authorization" && !strings.HasPrefix(strings.ToLower(value), "bearer ") {
		value = "Bearer " + value
	}
	headers[key] = value
	return headers
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

		errorDetails := map[string]any{
			"requestBody": string(body),
			"error":       err.Error(),
			"errorType":   fmt.Sprintf("%T", err),
		}

		// Enhanced error reporting for JSON errors
		switch e := err.(type) {
		case *json.UnmarshalTypeError:
			errorDetails["field"] = e.Field
			errorDetails["value"] = e.Value
			errorDetails["type"] = e.Type.String()
			logAction.SetError(
				fmt.Sprintf("Type error for field '%s' in struct '%s'", e.Field, structName),
				fmt.Sprintf("Check the type of field '%s' (expected %s, got %s)", e.Field, e.Type.String(), e.Value),
				errorDetails,
			)
		case *json.SyntaxError:
			errorDetails["offset"] = e.Offset
			logAction.SetError(
				"JSON syntax error",
				fmt.Sprintf("Syntax error at offset %d in struct '%s'", e.Offset, structName),
				errorDetails,
			)
		default:
			logAction.SetError(
				"Failed to decode the JSON request body",
				fmt.Sprintf("Ensure that the JSON is correct for %s", structName),
				errorDetails,
			)
		}
		return *logAction.Error
	}
	return logging.LogErrorInfo{}
}
