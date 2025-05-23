package mediux

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
)

// GetAllSets handles HTTP requests to retrieve all poster sets for a given TMDB ID and item type.
// It extracts the TMDB ID and item type from the URL parameters, validates them, and fetches
// the corresponding poster sets from the Mediux service. If successful, it responds with the
// retrieved data in JSON format. In case of errors, it sends an appropriate error response.
//
// Parameters:
//   - w: The HTTP response writer used to send the response.
//   - r: The HTTP request containing the URL parameters.
//
// URL Parameters:
//   - tmdbID: The TMDB ID of the item for which poster sets are being retrieved.
//   - itemType: The type of the item (e.g., movie, show).
//
// Responses:
//   - 200 OK: If the poster sets are successfully retrieved, the response contains a JSON object
//     with the status, message, elapsed time, and the retrieved data.
//   - 500 Internal Server Error: If there is an error (e.g., missing parameters or a failure
//     during data retrieval), the response contains an error message and elapsed time.
func GetAllSets(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	// Get the TMDB ID from the URL
	tmdbID := chi.URLParam(r, "tmdbID")
	itemType := chi.URLParam(r, "itemType")
	if tmdbID == "" || itemType == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{Err: errors.New("missing tmdbID or itemType in URL"),
			Log: logging.Log{
				Message: "Missing TMDB ID or item type in URL",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	posterSets, logErr := fetchAllSets(tmdbID, itemType)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Retrieved all sets from Mediux",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    posterSets,
	})
}

func fetchAllSets(tmdbID string, itemType string) (modals.PosterSets, logging.ErrorLog) {
	logging.LOG.Trace(fmt.Sprintf("Fetching all sets for TMDB ID: %s", tmdbID))

	// Generate the request body
	var requestBody map[string]any
	if itemType == "movie" {
		requestBody = generateMovieRequestBody(tmdbID)
	} else if itemType == "show" {
		requestBody = generateShowRequestBody(tmdbID)
	} else {
		return modals.PosterSets{}, logging.ErrorLog{
			Err: errors.New("invalid item type"),
			Log: logging.Log{Message: "Invalid item type provided"},
		}
	}

	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.Global.Mediux.Token)).
		SetBody(requestBody).
		Post("https://staged.mediux.io/graphql")
	if err != nil {
		return modals.PosterSets{}, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to send request to Mediux API"},
		}
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody modals.MediuxResponse

	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		return modals.PosterSets{}, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to parse response from Mediux API"},
		}
	}

	// Check if the response status is OK
	if response.StatusCode() != http.StatusOK {
		return modals.PosterSets{}, logging.ErrorLog{
			Err: errors.New("received non-200 response from Mediux API"),
			Log: logging.Log{Message: fmt.Sprintf("Received non-200 response from Mediux API: %s", response.String())},
		}
	}

	if itemType == "movie" {
		if responseBody.Data.Movie == nil {
			return modals.PosterSets{}, logging.ErrorLog{
				Err: errors.New("no movies found in the response"),
				Log: logging.Log{Message: "No movies found in the response"},
			}
		}

		if (responseBody.Data.Movie.MovieSets == nil || len(*responseBody.Data.Movie.MovieSets) == 0) &&
			(responseBody.Data.Movie.CollectionID == nil || len(responseBody.Data.Movie.CollectionID.CollectionSets) == 0) {
			return modals.PosterSets{}, logging.ErrorLog{
				Err: errors.New("no movie sets found in the response"),
				Log: logging.Log{Message: "No movie sets found in the response"},
			}
		}
	} else if itemType == "show" {
		if responseBody.Data.Show == nil {
			return modals.PosterSets{}, logging.ErrorLog{
				Err: errors.New("no shows found in the response"),
				Log: logging.Log{Message: "No shows found in the response"},
			}
		}

		if responseBody.Data.Show.ShowSets == nil || len(*responseBody.Data.Show.ShowSets) == 0 {
			return modals.PosterSets{}, logging.ErrorLog{
				Err: errors.New("no show sets found in the response"),
				Log: logging.Log{Message: "No show sets found in the response"},
			}
		}
	}

	var posterSets modals.PosterSets
	if itemType == "movie" {
		logging.LOG.Info(fmt.Sprintf("Start response for item type: %s", itemType))
		posterSets = processResponse(itemType, (*responseBody.Data.Movie))
	} else if itemType == "show" {
		posterSets = processResponse(itemType, (*responseBody.Data.Show))
	}

	return posterSets, logging.ErrorLog{}
}

func processResponse(itemType string, response modals.MediuxItem) modals.PosterSets {
	var posterSets modals.PosterSets
	posterSets.Type = itemType
	posterSets.Item.ID = response.ID
	posterSets.Item.Title = response.Title
	posterSets.Item.Status = response.Status
	posterSets.Item.Tagline = response.Tagline
	posterSets.Item.Slug = response.Slug
	posterSets.Item.DateUpdated = response.DateUpdated
	posterSets.Item.TvdbID = response.TvdbID
	posterSets.Item.ImdbID = response.ImdbID
	posterSets.Item.TraktID = response.TraktID
	if itemType == "movie" {
		posterSets.Item.ReleaseDate = response.ReleaseDate
		// Process CollectionID if it is not nil
		if response.CollectionID != nil {
			collectionSets := processCollectionSets(*response.CollectionID)
			posterSets.Sets = append(posterSets.Sets, collectionSets...)
		}

		// Process MovieSets if it is not nil
		if response.MovieSets != nil {
			movieSets := processPosterSets(itemType, *response.MovieSets)
			posterSets.Sets = append(posterSets.Sets, movieSets...)
		}
	} else if itemType == "show" {
		posterSets.Item.FirstAirDate = response.FirstAirDate
		showSets := processPosterSets(itemType, *response.ShowSets)
		posterSets.Sets = append(posterSets.Sets, showSets...)
	}

	return posterSets
}

func processPosterSets(setType string, mediuxPosterSets []modals.MediuxPosterSet) []modals.PosterSet {
	var posterSets []modals.PosterSet
	for _, set := range mediuxPosterSets {
		if len(set.Files) == 0 {
			continue
		}
		var posterSet modals.PosterSet
		posterSet.ID = set.ID
		posterSet.Type = setType
		posterSet.User.Name = set.UserCreated.Username
		posterSet.DateCreated = set.DateCreated
		posterSet.DateUpdated = set.DateUpdated
		files := processFiles(set.Files)
		posterSet.Files = files
		posterSets = append(posterSets, posterSet)
	}

	return posterSets
}

func processCollectionSets(collection modals.MediuxCollectionID) []modals.PosterSet {
	var posterSets []modals.PosterSet
	for _, set := range collection.CollectionSets {
		if len(set.Files) == 0 {
			continue
		}
		var posterSet modals.PosterSet
		posterSet.ID = set.ID
		posterSet.Type = "collection"
		posterSet.User.Name = set.UserCreated.Username
		posterSet.DateCreated = set.DateCreated
		posterSet.DateUpdated = set.DateUpdated
		files := processFiles(set.Files)
		posterSet.Files = files
		posterSets = append(posterSets, posterSet)
	}

	return posterSets
}

func parseFileSize(fileSize string) int64 {
	size, err := strconv.Atoi(fileSize)
	if err != nil {
		return 0
	}
	return int64(size)
}

func processFiles(files []modals.MediuxFileItem) []modals.PosterFile {
	var posterFiles []modals.PosterFile
	for _, file := range files {
		if file.FileType == "album" {
			continue
		}
		if file.FileType == "misc" {
			file.FileType = "backdrop"
		}
		if file.Season != nil {
			file.FileType = "seasonPoster"
		}

		posterFile := modals.PosterFile{
			ID:       file.ID,
			Type:     file.FileType,
			Modified: file.ModifiedOn,
			FileSize: parseFileSize(file.FileSize),
		}

		// Safely assign Movie if it is not nil
		if file.Movie != nil {
			posterFile.Movie = &modals.PosterFileMovie{
				ID: file.Movie.ID,
			}
		}

		// Safely assign Season if it is not nil
		if file.Season != nil {
			posterFile.Season = &modals.PosterFileSeason{
				Number: file.Season.SeasonNumber,
			}
		}

		// Safely assign Episode if it is not nil
		if file.Episode != nil {
			posterFile.Episode = &modals.PosterFileEpisode{
				Title:         file.Episode.EpisodeTitle,
				EpisodeNumber: file.Episode.EpisodeNumber,
			}
			// Safely assign SeasonNumber if Episode.Season is not nil
			if file.Episode.Season != nil {
				posterFile.Episode.SeasonNumber = file.Episode.Season.SeasonNumber
			}
		}

		posterFiles = append(posterFiles, posterFile)
	}

	return posterFiles
}

func FetchSetByID(set modals.PosterSet, tmdbID string) (modals.PosterSet, logging.ErrorLog) {

	setType := set.Type
	setID := set.ID

	var requestBody map[string]any
	if setType == "collection" {
		requestBody = generateCollectionSetByIDRequestBody(setID, tmdbID)
	} else if setType == "movie" {
		requestBody = generateMovieSetByIDRequestBody(setID, tmdbID)
	} else if setType == "show" {
		requestBody = generateShowSetByIDRequestBody(setID)
	}

	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.Global.Mediux.Token)).
		SetBody(requestBody).
		Post("https://staged.mediux.io/graphql")
	if err != nil {
		return modals.PosterSet{}, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to send request to Mediux API"},
		}
	}

	if response.StatusCode() != http.StatusOK {
		return modals.PosterSet{}, logging.ErrorLog{
			Err: errors.New("received non-200 response from Mediux API"),
			Log: logging.Log{Message: fmt.Sprintf("Received non-200 response from Mediux API: %s", response.String())},
		}
	}

	var responseBody modals.MediuxSetResponse
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		return modals.PosterSet{}, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to parse response from Mediux API"},
		}
	}

	if setType == "movie" {
		if responseBody.Data.MovieSet == nil {
			return modals.PosterSet{}, logging.ErrorLog{
				Err: errors.New("no movies found in the response"),
				Log: logging.Log{Message: "No movie found in the response"},
			}
		}
	} else if setType == "show" {
		if responseBody.Data.ShowSet == nil {
			return modals.PosterSet{}, logging.ErrorLog{
				Err: errors.New("no shows found in the response"),
				Log: logging.Log{Message: "No show found in the response"},
			}
		}
	} else if setType == "collection" {
		if responseBody.Data.CollectionSet == nil {
			return modals.PosterSet{}, logging.ErrorLog{
				Err: errors.New("no collection found in the response"),
				Log: logging.Log{Message: "No collection found in the response"},
			}
		}
	}

	var posterSet modals.PosterSet
	if setType == "show" {
		posterSet = modals.PosterSet{
			ID:          responseBody.Data.ShowSet.ID,
			Type:        "show",
			User:        modals.SetUser{Name: responseBody.Data.ShowSet.UserCreated.Username},
			DateCreated: responseBody.Data.ShowSet.DateCreated,
			DateUpdated: responseBody.Data.ShowSet.DateUpdated,
			Files:       processFiles(responseBody.Data.ShowSet.Files),
		}
	} else if setType == "movie" {
		posterSet = modals.PosterSet{
			ID:          responseBody.Data.MovieSet.ID,
			Type:        "movie",
			User:        modals.SetUser{Name: responseBody.Data.MovieSet.UserCreated.Username},
			DateCreated: responseBody.Data.MovieSet.DateCreated,
			DateUpdated: responseBody.Data.MovieSet.DateUpdated,
			Files:       processFiles(responseBody.Data.MovieSet.Files),
		}
	}

	return posterSet, logging.ErrorLog{}
}
