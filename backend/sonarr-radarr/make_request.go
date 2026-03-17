package sonarr_radarr

import (
	"aura/config"
	"aura/logging"
	"aura/utils/httpx"
	"context"
	"fmt"
	"net/http"
)

func makeRequest(ctx context.Context, app config.Config_SonarrRadarrApp, url string, method string, body []byte) (resp *http.Response, respBody []byte, Err logging.LogErrorInfo) {
	// Make the HTTP Headers for this request
	headers := makeAuthHeader(app.ApiToken)

	// Make the HTTP request to Sonarr/Radarr
	resp, respBody, Err = httpx.MakeHTTPRequest(ctx, url, method, headers, 60, body, app.Type)
	if Err.Message != "" {
		return nil, nil, Err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, logging.LogErrorInfo{
			Message: fmt.Sprintf("%s Server returned a non-success status code", app.Type),
			Help:    "Check the response from the server for more details",
			Detail:  map[string]any{"status_code": resp.StatusCode, "error_body": string(respBody)},
		}
	}

	return resp, respBody, logging.LogErrorInfo{}
}

func makeAuthHeader(token string) (headers map[string]string) {
	headers = make(map[string]string)
	headers["X-Api-Key"] = token
	headers["Accept"] = "application/json"
	return headers
}
