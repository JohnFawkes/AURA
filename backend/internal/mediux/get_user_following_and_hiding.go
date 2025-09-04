package mediux

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

func GetUserFollowingAndHiding(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	// Fetch user following and hiding data from the Mediux API
	responseBody, Err := fetchUserFollowingAndHiding()
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	data := responseBody.Data
	if data.Follows == nil && data.Hides == nil {
		utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
			Status:  "success",
			Elapsed: utils.ElapsedTime(startTime),
			Data:    nil,
		})
		return
	}

	// Create a new UserFollowHide object to store the response data
	userFollowHide := modals.UserFollowHide{
		Follows: make([]modals.MediuxUserFollowHideUserInfo, 0),
		Hides:   make([]modals.MediuxUserFollowHideUserInfo, 0),
	}

	// Populate the UserFollowHide object with the response data
	if data.Follows != nil {
		for _, follow := range data.Follows {
			userFollowHide.Follows = append(userFollowHide.Follows, modals.MediuxUserFollowHideUserInfo{
				ID:       follow.FolloweeID.ID,
				Username: follow.FolloweeID.Username,
			})
		}
	}
	if data.Hides != nil {
		for _, hide := range data.Hides {
			userFollowHide.Hides = append(userFollowHide.Hides, modals.MediuxUserFollowHideUserInfo{
				ID:       hide.HidingID.ID,
				Username: hide.HidingID.Username,
			})
		}
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    userFollowHide,
	})
}

func fetchUserFollowingAndHiding() (modals.MediuxUserFollowHideResponse, logging.StandardError) {
	requestBody := generateUserFollowingAndHidingBody()
	Err := logging.NewStandardError()

	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.Global.Mediux.Token)).
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
		return modals.MediuxUserFollowHideResponse{}, Err
	}
	// Check if the response status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		Err.Message = "Unexpected response from Mediux API"
		Err.HelpText = "Check if the Mediux API endpoint is correct and the token is valid."
		Err.Details = map[string]any{
			"statusCode":   response.StatusCode(),
			"responseBody": string(response.Body()),
		}
		return modals.MediuxUserFollowHideResponse{}, Err
	}

	var responseBody modals.MediuxUserFollowHideResponse
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {

		Err.Message = "Failed to unmarshal response from Mediux API"
		Err.HelpText = "Ensure the response format matches the expected structure."
		Err.Details = fmt.Sprintf("Error: %s, Response Body: %s", err.Error(), string(response.Body()))
		return modals.MediuxUserFollowHideResponse{}, Err
	}

	return responseBody, logging.StandardError{}
}
