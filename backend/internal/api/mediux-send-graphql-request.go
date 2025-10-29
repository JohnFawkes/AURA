package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func Mediux_SendGraphQLRequest(ctx context.Context, requestBody map[string]any) (*resty.Response, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Send GraphQL Request to Mediux", logging.LevelTrace)
	defer logAction.Complete()

	client := resty.New()

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", Global_Config.Mediux.Token)).
		SetBody(requestBody).
		Post("https://images.mediux.io/graphql")
	if err != nil {
		logAction.SetError("Failed to send GraphQL request to Mediux", "Ensure the Mediux API is reachable and the token is valid.",
			map[string]any{
				"error":        err.Error(),
				"statusCode":   resp.StatusCode(),
				"responseBody": resp.String(),
			})
		return nil, *logAction.Error
	}
	if resp.StatusCode() != http.StatusOK {
		logAction.SetError("Mediux GraphQL request returned non-200 status", "Check the Mediux API status or your request parameters.",
			map[string]any{
				"statusCode":   resp.StatusCode(),
				"responseBody": resp.String(),
			})
		return nil, *logAction.Error
	}
	return resp, logging.LogErrorInfo{}
}
