package mediux

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
)

func GetAllUserSets(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	// Get the username from the URL
	username := chi.URLParam(r, "username")
	if username == "" {
		utils.SendErrorJSONResponse(w, http.StatusBadRequest, logging.ErrorLog{
			Log: logging.Log{
				Message: "Username is required",
				Elapsed: utils.ElapsedTime(startTime),
			},
			Err: errors.New("username is required"),
		})
		return
	}

	allSetsResponse, logErr := fetchAllUserSets(username)
	if logErr.Err != nil {
		logging.LOG.Error(logErr.Log.Message)
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: fmt.Sprintf("Fetched all sets for '%s'", username),
		Elapsed: utils.ElapsedTime(startTime),
		Data:    allSetsResponse.Data,
	})
}

func fetchAllUserSets(username string) (MediuxUserAllSetsResponse, logging.ErrorLog) {
	requestBody := generateAllUserSetsBody(username)

	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.Global.Mediux.Token)).
		SetBody(requestBody).
		Post("https://staged.mediux.io/graphql")
	if err != nil {
		return MediuxUserAllSetsResponse{}, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to send request to Mediux API"},
		}
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody MediuxUserAllSetsResponse

	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Response error: %s", response.Body()))
		return MediuxUserAllSetsResponse{}, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to parse response from Mediux API"},
		}
	}

	return responseBody, logging.ErrorLog{}
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
