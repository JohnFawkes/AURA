package mediux

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
)

func GetAllUserSets(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	// Get the username from the URL
	username := chi.URLParam(r, "username")
	if username == "" {

		Err.Message = "Missing username in URL"
		Err.HelpText = "Ensure the username is provided in the URL path."
		Err.Details = fmt.Sprintf("URL Path: %s", r.URL.Path)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	allSetsResponse, Err := fetchAllUserSets(username)
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    allSetsResponse.Data,
	})
}

func fetchAllUserSets(username string) (MediuxUserAllSetsResponse, logging.StandardError) {
	requestBody := generateAllUserSetsBody(username)
	Err := logging.NewStandardError()

	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.Global.Mediux.Token)).
		SetBody(requestBody).
		Post("https://staged.mediux.io/graphql")
	if err != nil {
		Err.Message = "Failed to send request to Mediux API"
		Err.HelpText = "Check if the Mediux API is reachable and the token is valid."
		Err.Details = map[string]any{
			"error":        err.Error(),
			"requestBody":  requestBody,
			"username":     username,
			"responseBody": string(response.Body()),
			"statusCode":   response.StatusCode(),
		}
		return MediuxUserAllSetsResponse{}, Err
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody MediuxUserAllSetsResponse

	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {

		Err.Message = "Failed to unmarshal response from Mediux API"
		Err.HelpText = "Ensure the response format matches the expected structure."
		Err.Details = fmt.Sprintf("Error: %s, Response Body: %s", err.Error(), string(response.Body()))
		return MediuxUserAllSetsResponse{}, Err
	}

	return responseBody, logging.StandardError{}
}

type MediuxUserAllSetsResponse struct {
	Data struct {
		ShowSets       []MediuxUserShowSet       `json:"show_sets"`
		MovieSets      []MediuxUserMovieSet      `json:"movie_sets"`
		CollectionSets []MediuxUserCollectionSet `json:"collection_sets"`
		Boxsets        []MediuxUserBoxset        `json:"boxsets"`
	} `json:"data"`
}

type MediuxUserShowSet struct {
	ID            string                   `json:"id"`
	UserCreated   MediuxUserCreated        `json:"user_created,omitempty"`
	SetTitle      string                   `json:"set_title"`
	DateCreated   time.Time                `json:"date_created"`
	DateUpdated   time.Time                `json:"date_updated"`
	ShowID        MediuxUserShow           `json:"show_id"`
	ShowPoster    []MediuxUserImage        `json:"show_poster"`
	ShowBackdrop  []MediuxUserImage        `json:"show_backdrop"`
	SeasonPosters []MediuxUserSeasonPoster `json:"season_posters"`
	Titlecards    []MediuxUserTitlecard    `json:"titlecards"`
}

type MediuxUserMovieSet struct {
	ID            string            `json:"id"`
	UserCreated   MediuxUserCreated `json:"user_created"`
	SetTitle      string            `json:"set_title"`
	DateCreated   time.Time         `json:"date_created"`
	DateUpdated   time.Time         `json:"date_updated"`
	MovieID       MediuxUserMovie   `json:"movie_id"`
	MoviePoster   []MediuxUserImage `json:"movie_poster"`
	MovieBackdrop []MediuxUserImage `json:"movie_backdrop"`
}

type MediuxUserCollectionSet struct {
	ID             string                      `json:"id"`
	UserCreated    MediuxUserCreated           `json:"user_created"`
	SetTitle       string                      `json:"set_title"`
	DateCreated    time.Time                   `json:"date_created"`
	DateUpdated    time.Time                   `json:"date_updated"`
	MoviePosters   []MediuxUserCollectionMovie `json:"movie_posters"`
	MovieBackdrops []MediuxUserCollectionMovie `json:"movie_backdrops"`
}

type MediuxUserBoxset struct {
	ID             string                    `json:"id"`
	UserCreated    MediuxUserCreated         `json:"user_created"`
	BoxsetTitle    string                    `json:"boxset_title"`
	DateCreated    time.Time                 `json:"date_created"`
	DateUpdated    time.Time                 `json:"date_updated"`
	MovieSets      []MediuxUserMovieSet      `json:"movie_sets"`
	ShowSets       []MediuxUserShowSet       `json:"show_sets"`
	CollectionSets []MediuxUserCollectionSet `json:"collection_sets"`
}

// Reusable subtypes

type MediuxUserCreated struct {
	Username string `json:"username"`
}

type MediuxUserShow struct {
	ID           string    `json:"id"`
	DateUpdated  time.Time `json:"date_updated"`
	Status       string    `json:"status"`
	Title        string    `json:"title"`
	Tagline      string    `json:"tagline"`
	FirstAirDate string    `json:"first_air_date"`
	TvdbID       string    `json:"tvdb_id"`
	ImdbID       string    `json:"imdb_id"`
	TraktID      string    `json:"trakt_id"`
	Slug         string    `json:"slug"`
}

type MediuxUserMovie struct {
	ID          string    `json:"id"`
	DateUpdated time.Time `json:"date_updated"`
	Status      string    `json:"status"`
	Title       string    `json:"title"`
	Tagline     string    `json:"tagline"`
	ReleaseDate string    `json:"release_date"`
	TvdbID      string    `json:"tvdb_id"`
	ImdbID      string    `json:"imdb_id"`
	TraktID     string    `json:"trakt_id"`
	Slug        string    `json:"slug"`
}

type MediuxUserImage struct {
	ID         string    `json:"id"`
	ModifiedOn time.Time `json:"modified_on"`
	Filesize   string    `json:"filesize"`
}

type MediuxUserSeasonPoster struct {
	ID         string    `json:"id"`
	ModifiedOn time.Time `json:"modified_on"`
	Filesize   string    `json:"filesize"`
	Season     struct {
		SeasonNumber int `json:"season_number"`
	} `json:"season"`
}

type MediuxUserTitlecard struct {
	ID         string    `json:"id"`
	ModifiedOn time.Time `json:"modified_on"`
	Filesize   string    `json:"filesize"`
	Episode    struct {
		EpisodeTitle  string `json:"episode_title"`
		EpisodeNumber int    `json:"episode_number"`
		SeasonID      struct {
			SeasonNumber int `json:"season_number"`
		} `json:"season_id"`
	} `json:"episode"`
}

type MediuxUserCollectionMovie struct {
	ID         string    `json:"id"`
	ModifiedOn time.Time `json:"modified_on"`
	Filesize   string    `json:"filesize"`
	Movie      struct {
		ID          string    `json:"id"`
		DateUpdated time.Time `json:"date_updated"`
		Status      string    `json:"status"`
		Title       string    `json:"title"`
		Tagline     string    `json:"tagline"`
		ReleaseDate string    `json:"release_date"`
		TvdbID      string    `json:"tvdb_id"`
		ImdbID      string    `json:"imdb_id"`
		TraktID     string    `json:"trakt_id"`
		Slug        string    `json:"slug"`
	} `json:"movie"`
}
