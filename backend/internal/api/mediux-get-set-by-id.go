package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func Mediux_FetchShowSetByID(librarySection, tmdbID, setID string) (PosterSet, logging.StandardError) {

	Err := logging.NewStandardError()

	requestBody := Mediux_GenerateShowSetByIDRequestBody(setID)

	// If cache is empty, return false
	if Global_Cache_LibraryStore.IsEmpty() {
		Err.Message = "Backend cache is empty"
		Err.HelpText = "Try refreshing the cache from the Home Page"
		Err.Details = "This typically happens when the backend cache is not initialized or has been cleared. Example on application restart."
		return PosterSet{}, Err
	}

	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", Global_Config.Mediux.Token)).
		SetBody(requestBody).
		Post("https://images.mediux.io/graphql")
	if err != nil {

		Err.Message = "Failed to make request to Mediux API"
		Err.HelpText = "Ensure the Mediux API is reachable and the token is valid."
		Err.Details = map[string]any{
			"error":          err.Error(),
			"responseBody":   string(response.Body()),
			"statusCode":     response.StatusCode(),
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
			"setID":          setID,
		}
		return PosterSet{}, Err
	}
	if response.StatusCode() != http.StatusOK {
		Err.Message = "Mediux API returned non-OK status"
		Err.HelpText = "Check the Mediux API status or your request parameters."
		Err.Details = map[string]any{
			"responseBody":   string(response.Body()),
			"statusCode":     response.StatusCode(),
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
			"setID":          setID,
		}
		return PosterSet{}, Err
	}

	// Parse the response body into a MediuxShowSetResponse struct
	var responseBody MediuxShowSetResponse
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		Err.Message = "Failed to parse Mediux API response"
		Err.HelpText = "Ensure the Mediux API is returning a valid JSON response."
		Err.Details = map[string]any{
			"error":          err.Error(),
			"responseBody":   string(response.Body()),
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
			"setID":          setID,
		}
		return PosterSet{}, Err
	}

	showSet := responseBody.Data.ShowSetID

	if showSet.ID == "" {
		Err.Message = "No show set found in the response"
		Err.HelpText = "Ensure the set ID is correct and the show set exists in the Mediux database."
		Err.Details = map[string]any{
			"setID":          setID,
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
		}
		return PosterSet{}, Err
	}

	logging.LOG.Trace(fmt.Sprintf("Processing show set: %s", showSet.SetTitle))
	logging.LOG.Trace(fmt.Sprintf("Date Created: %s", showSet.DateCreated))
	logging.LOG.Trace(fmt.Sprintf("Date Updated: %s", showSet.DateUpdated))

	// Process the response and return the poster sets
	posterSets := processShowResponse(librarySection, tmdbID, showSet.Show)

	if len(posterSets) == 0 {
		Err.Message = "No poster sets found for the provided show set ID"
		Err.HelpText = "Ensure the show set ID is correct and the show set exists in the Mediux database."
		Err.Details = map[string]any{
			"setID":          setID,
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
		}
		return PosterSet{}, Err
	}

	posterSet := posterSets[0]

	posterSet.ID = showSet.ID
	posterSet.Title = showSet.SetTitle
	posterSet.User.Name = showSet.UserCreated.Username
	posterSet.Type = "show"
	posterSet.DateCreated = showSet.DateCreated
	posterSet.DateUpdated = showSet.DateUpdated

	return posterSet, logging.StandardError{}
}

func Mediux_FetchMovieSetByID(librarySection, tmdbID, setID string) (PosterSet, logging.StandardError) {
	Err := logging.NewStandardError()

	requestBody := Mediux_GenerateMovieSetByIDRequestBody(setID)

	// If cache is empty, return false
	if Global_Cache_LibraryStore.IsEmpty() {
		Err.Message = "Backend cache is empty"
		Err.HelpText = "Try refreshing the cache from the Home Page"
		Err.Details = "This typically happens when the backend cache is not initialized or has been cleared. Example on application restart."
		return PosterSet{}, Err
	}

	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", Global_Config.Mediux.Token)).
		SetBody(requestBody).
		Post("https://images.mediux.io/graphql")
	if err != nil {

		Err.Message = "Failed to make request to Mediux API"
		Err.HelpText = "Ensure the Mediux API is reachable and the token is valid."
		Err.Details = map[string]any{
			"error":          err.Error(),
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
			"setID":          setID,
			"responseBody":   string(response.Body()),
			"statusCode":     response.StatusCode(),
		}
		return PosterSet{}, Err
	}
	if response.StatusCode() != http.StatusOK {
		Err.Message = "Mediux API returned non-OK status"
		Err.HelpText = "Check the Mediux API status or your request parameters."
		Err.Details = map[string]any{
			"responseBody":   string(response.Body()),
			"statusCode":     response.StatusCode(),
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
			"setID":          setID,
		}
		return PosterSet{}, Err
	}

	// Parse the response body into a MediuxMovieSetResponse struct
	var responseBody MediuxMovieSetResponse
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		Err.Message = "Failed to parse Mediux API response"
		Err.HelpText = "Ensure the Mediux API is returning a valid JSON response."
		Err.Details = map[string]any{
			"error":          err.Error(),
			"responseBody":   string(response.Body()),
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
			"setID":          setID,
		}
		return PosterSet{}, Err
	}

	movieSet := responseBody.Data.MovieSetID
	logging.LOG.Trace(fmt.Sprintf("Processing movie set: %s", movieSet.SetTitle))
	logging.LOG.Trace(fmt.Sprintf("Date Created: %s", movieSet.DateCreated))
	logging.LOG.Trace(fmt.Sprintf("Date Updated: %s", movieSet.DateUpdated))

	// Process the response and return the poster sets
	posterSets := processMovieSetPostersAndBackdrops(librarySection, tmdbID, movieSet.Movie)
	if len(posterSets) == 0 {
		Err.Message = "No poster sets found for the provided movie set ID"
		Err.HelpText = "Ensure the movie set ID is correct and the movie set exists in the Mediux database."
		Err.Details = map[string]any{
			"setID":          setID,
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
		}
		return PosterSet{}, Err
	}
	posterSet := posterSets[0]

	posterSet.ID = movieSet.ID
	posterSet.Title = movieSet.SetTitle
	posterSet.User.Name = movieSet.UserCreated.Username
	posterSet.Type = "movie"
	posterSet.DateCreated = movieSet.DateCreated
	posterSet.DateUpdated = movieSet.DateUpdated
	posterSet.Status = movieSet.Movie.Status

	return posterSet, logging.StandardError{}
}

func Mediux_FetchCollectionSetByID(librarySection, tmdbID, setID string) (PosterSet, logging.StandardError) {
	Err := logging.NewStandardError()

	// If cache is empty, return false
	if Global_Cache_LibraryStore.IsEmpty() {
		Err.Message = "Backend cache is empty"
		Err.HelpText = "Try refreshing the cache from the Home Page"
		Err.Details = "This typically happens when the backend cache is not initialized or has been cleared. Example on application restart."
		return PosterSet{}, Err
	}

	requestBody := Mediux_GenerateCollectionSetByIDRequestBody(setID, tmdbID)
	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", Global_Config.Mediux.Token)).
		SetBody(requestBody).
		Post("https://images.mediux.io/graphql")
	if err != nil {

		Err.Message = "Failed to make request to Mediux API"
		Err.HelpText = "Ensure the Mediux API is reachable and the token is valid."
		Err.Details = map[string]any{
			"error":          err.Error(),
			"tmdbID":         tmdbID,
			"librarySection": librarySection,
			"responseBody":   string(response.Body()),
			"statusCode":     response.StatusCode(),
		}
		return PosterSet{}, Err
	}
	if response.StatusCode() != http.StatusOK {
		Err.Message = "Mediux API returned non-OK status"
		Err.HelpText = "Check the Mediux API status or your request parameters."
		Err.Details = map[string]any{
			"responseBody":   string(response.Body()),
			"statusCode":     response.StatusCode(),
			"tmdbID":         tmdbID,
			"librarySection": librarySection,
			"setID":          setID,
		}
		return PosterSet{}, Err
	}

	// Parse the response body into a MediuxMovieSetResponse struct
	var responseBody MediuxCollectionSetResponse
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		Err.Message = "Failed to parse Mediux API response"
		Err.HelpText = "Ensure the Mediux API is returning a valid JSON response."
		Err.Details = map[string]any{
			"error":        err.Error(),
			"responseBody": string(response.Body()),
		}
		return PosterSet{}, Err
	}

	collectionSet := responseBody.Data.CollectionSetID
	logging.LOG.Trace(fmt.Sprintf("Processing collection set: %s", collectionSet.SetTitle))
	logging.LOG.Trace(fmt.Sprintf("Date Created: %s", collectionSet.DateCreated))
	logging.LOG.Trace(fmt.Sprintf("Date Updated: %s", collectionSet.DateUpdated))
	if collectionSet.ID == "" {

		Err.Message = "No collection set found for the provided ID"
		Err.HelpText = "Ensure the collection set ID is correct and the collection set exists in the Mediux database."
		Err.Details = map[string]any{
			"setID":          setID,
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
		}
		return PosterSet{}, Err
	}

	// Process the response and return the poster sets
	posterSets := processMovieCollection(tmdbID, librarySection, tmdbID, collectionSet.Collection)
	if len(posterSets) == 0 {
		Err.Message = "No poster sets found for the provided collection set ID"
		Err.HelpText = "Ensure the collection set ID is correct and the collection set exists in the Mediux database."
		Err.Details = map[string]any{
			"setID":          setID,
			"librarySection": librarySection,
			"tmdbID":         tmdbID,
		}
		return PosterSet{}, logging.StandardError{}
	}

	posterSet := posterSets[0]
	posterSet.ID = collectionSet.ID
	posterSet.Title = collectionSet.SetTitle
	posterSet.User.Name = collectionSet.UserCreated.Username
	posterSet.Type = "collection"
	posterSet.DateCreated = collectionSet.DateCreated
	posterSet.DateUpdated = collectionSet.DateUpdated

	return posterSet, logging.StandardError{}

}
