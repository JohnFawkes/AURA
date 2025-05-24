package mediux

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

func GetUserFollowingAndHiding(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	// Fetch user following and hiding data from the Mediux API
	responseBody, logErr := fetchUserFollowingAndHiding()
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	data := responseBody.Data
	if data.Follows == nil && data.Hides == nil {
		utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
			Status:  "success",
			Message: "No user following and hiding data found",
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
		Message: "Retrieved user following and hiding data successfully",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    userFollowHide,
	})
}

func fetchUserFollowingAndHiding() (modals.MediuxUserFollowHideResponse, logging.ErrorLog) {
	requestBody := generateUserFollowingAndHidingBody()

	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.Global.Mediux.Token)).
		SetBody(requestBody).
		Post("https://staged.mediux.io/graphql")
	if err != nil {
		return modals.MediuxUserFollowHideResponse{}, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to send request to Mediux API"},
		}
	}

	var responseBody modals.MediuxUserFollowHideResponse
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		return modals.MediuxUserFollowHideResponse{}, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to unmarshal response from Mediux API"},
		}
	}
	if response.StatusCode() != http.StatusOK {
		return modals.MediuxUserFollowHideResponse{}, logging.ErrorLog{
			Err: errors.New("received non-200 response from Mediux API"),
			Log: logging.Log{Message: fmt.Sprintf("Received non-200 response from Mediux API: %s", response.String())},
		}
	}

	return responseBody, logging.ErrorLog{}
}
