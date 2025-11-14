package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

var mediuxRestyClient = resty.New().
	SetHeader("Content-Type", "application/json").
	SetHeader("User-Agent", "aura/1.0").
	SetHeader("X-Request", "mediux-aura")

func Mediux_SendGraphQLRequest(ctx context.Context, requestBody map[string]any) (*resty.Response, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Send GraphQL Request to MediUX", logging.LevelTrace)
	defer logAction.Complete()

	resp, err := mediuxRestyClient.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", Global_Config.Mediux.Token)).
		SetBody(requestBody).
		Post("https://images.mediux.io/graphql")
	if err != nil {
		logAction.SetError("Failed to send GraphQL request to MediUX", "Ensure the MediUX API is reachable and the token is valid.",
			map[string]any{
				"error":        err.Error(),
				"statusCode":   resp.StatusCode(),
				"responseBody": resp.String(),
			})
		return nil, *logAction.Error
	}
	if resp.StatusCode() != http.StatusOK {
		logAction.SetError("MediUX GraphQL request returned non-200 status", "Check the MediUX API status or your request parameters.",
			map[string]any{
				"statusCode":   resp.StatusCode(),
				"responseBody": resp.String(),
			})
		return nil, *logAction.Error
	}
	return resp, logging.LogErrorInfo{}
}
