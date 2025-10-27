package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// Mediux_FetchAllUserSets fetches all user-defined sets from Mediux for a specific user.
//
// Takes in the username as a parameter.
//
// Returns a struct containing arrays of different set types (shows, movies, collections, boxsets).
// Example response structure:
/*
{
Data struct {
		ShowSets       []MediuxUserShowSet       `json:"show_sets"`
		MovieSets      []MediuxUserMovieSet      `json:"movie_sets"`
		CollectionSets []MediuxUserCollectionSet `json:"collection_sets"`
		Boxsets        []MediuxUserBoxset        `json:"boxsets"`
	} `json:"data"`
}
*/
func Mediux_FetchAllUserSets(username string) (MediuxUserAllSetsResponse, logging.StandardError) {
	requestBody := Mediux_GenerateAllUserSetsBody(username)
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
			"username":     username,
			"responseBody": string(response.Body()),
			"statusCode":   response.StatusCode(),
		}
		return MediuxUserAllSetsResponse{}, Err
	}
	// Check if the response status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		Err.Message = "Unexpected response from Mediux API"
		Err.HelpText = "Check if the Mediux API endpoint is correct and the token is valid."
		Err.Details = map[string]any{
			"statusCode":   response.StatusCode(),
			"responseBody": string(response.Body()),
		}
		return MediuxUserAllSetsResponse{}, Err
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody MediuxUserAllSetsResponse

	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		Err.Message = "Failed to unmarshal response from Mediux API"
		Err.HelpText = "Ensure the response format matches the expected structure."
		Err.Details = map[string]any{
			"error":        err.Error(),
			"responseBody": string(response.Body()),
		}
		return MediuxUserAllSetsResponse{}, Err
	}

	return responseBody, logging.StandardError{}
}
