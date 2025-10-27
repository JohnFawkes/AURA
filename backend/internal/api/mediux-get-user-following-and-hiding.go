package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func Mediux_FetchUserFollowingAndHiding() (UserFollowHide, logging.StandardError) {
	requestBody := Mediux_GenerateUserFollowingAndHidingBody()
	Err := logging.NewStandardError()

	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", Global_Config.Mediux.Token)).
		SetBody(requestBody).
		Post("https://images.mediux.io/graphql")
	if err != nil {
		Err.Message = "Failed to send request to Mediux API"
		Err.HelpText = "Check if the Mediux API is reachable and the token is valid."
		Err.Details = map[string]any{
			"error":        err.Error(),
			"responseBody": string(response.Body()),
			"statusCode":   response.StatusCode(),
		}
		return UserFollowHide{}, Err
	}
	// Check if the response status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		Err.Message = "Unexpected response from Mediux API"
		Err.HelpText = "Check if the Mediux API endpoint is correct and the token is valid."
		Err.Details = map[string]any{
			"statusCode":   response.StatusCode(),
			"responseBody": string(response.Body()),
		}
		return UserFollowHide{}, Err
	}

	var responseBody MediuxUserFollowHideResponse
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		Err.Message = "Failed to unmarshal response from Mediux API"
		Err.HelpText = "Ensure the response format matches the expected structure."
		Err.Details = map[string]any{
			"error":        err.Error(),
			"responseBody": string(response.Body()),
		}
		return UserFollowHide{}, Err
	}

	data := responseBody.Data
	if data.Follows == nil && data.Hides == nil {
		return UserFollowHide{}, logging.StandardError{}
	}

	// Create a new UserFollowHide object to store the response data
	userFollowHide := UserFollowHide{
		Follows: make([]MediuxUserFollowHideUserInfo, 0),
		Hides:   make([]MediuxUserFollowHideUserInfo, 0),
	}

	// Populate the UserFollowHide object with the response data
	if data.Follows != nil {
		for _, follow := range data.Follows {
			userFollowHide.Follows = append(userFollowHide.Follows, MediuxUserFollowHideUserInfo{
				ID:       follow.FolloweeID.ID,
				Username: follow.FolloweeID.Username,
			})
		}
	}
	if data.Hides != nil {
		for _, hide := range data.Hides {
			userFollowHide.Hides = append(userFollowHide.Hides, MediuxUserFollowHideUserInfo{
				ID:       hide.HidingID.ID,
				Username: hide.HidingID.Username,
			})
		}
	}

	return userFollowHide, logging.StandardError{}
}
