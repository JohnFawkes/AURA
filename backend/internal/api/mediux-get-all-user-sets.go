package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
)

// Mediux_FetchAllUserSets fetches all user-defined sets from MediUX for a specific user.
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
func Mediux_FetchAllUserSets(ctx context.Context, username string) (MediuxUserAllSetsResponse, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Fetch All '%s' User Sets from MediUX", username), logging.LevelDebug)
	defer logAction.Complete()

	requestBody := Mediux_GenerateAllUserSetsBody(username)

	// Send the GraphQL request
	resp, Err := Mediux_SendGraphQLRequest(ctx, requestBody)
	if Err.Message != "" {
		Err.Detail["username"] = username
		return MediuxUserAllSetsResponse{}, Err
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody MediuxUserAllSetsResponse
	Err = DecodeJSONBody(ctx, resp.Body(), &responseBody, "MediuxUserAllSetsResponse")
	if Err.Message != "" {
		return MediuxUserAllSetsResponse{}, Err
	}

	return responseBody, logging.LogErrorInfo{}
}
