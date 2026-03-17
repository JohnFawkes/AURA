package plex

import (
	"aura/config"
	"aura/logging"
	"aura/utils/httpx"
	"context"
	"net/http"
)

func makeRequest(ctx context.Context, msConfig config.Config_MediaServer, url string, method string, body []byte) (resp *http.Response, respBody []byte, Err logging.LogErrorInfo) {
	// Make the HTTP Headers for this request
	headers := make(map[string]string)
	headers = AddPlexHeaders(msConfig, headers)

	// Make the HTTP request to Plex
	resp, respBody, Err = httpx.MakeHTTPRequest(ctx, url, method, headers, 60, body, "Plex")
	if Err.Message != "" {
		return nil, nil, Err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, logging.LogErrorInfo{
			Message: "Plex Server returned a non-success status code",
			Help:    "Check the response from the server for more details",
			Detail:  map[string]any{"status_code": resp.StatusCode, "error_body": string(respBody)},
		}
	}

	return resp, respBody, logging.LogErrorInfo{}
}

func AddPlexHeaders(msConfig config.Config_MediaServer, headers map[string]string) map[string]string {
	headers["X-Plex-Token"] = msConfig.ApiToken
	headers["Accept"] = "application/json"
	headers["X-Plex-Product"] = "aura"
	headers["X-Plex-Client-Identifier"] = "aura"
	headers["X-Plex-Device"] = "aura"
	headers["X-Plex-Platform"] = "aura"
	headers["X-Plex-Version"] = "1.0"
	return headers
}
