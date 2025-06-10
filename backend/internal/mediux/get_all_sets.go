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
	librarySection := chi.URLParam(r, "librarySection")
	itemRatingKey := chi.URLParam(r, "ratingKey")
	if tmdbID == "" || itemType == "" || librarySection == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{Err: errors.New("missing tmdbID or itemType in URL"),
			Log: logging.Log{
				Message: "Missing TMDB ID, item type, or library section in URL",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	logging.LOG.Debug(fmt.Sprintf("Fetching all sets for TMDB ID: %s, item type: %s, library section: %s", tmdbID, itemType, librarySection))

	posterSets, logErr := fetchAllSets(tmdbID, itemType, librarySection, itemRatingKey)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	if len(posterSets) == 0 {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: errors.New("no sets found"),
			Log: logging.Log{
				Message: "No sets found for the provided TMDB ID and item type",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
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

func fetchAllSets(tmdbID, itemType, librarySection, itemRatingKey string) ([]modals.PosterSet, logging.ErrorLog) {
	// Generate the request body
	var requestBody map[string]any
	if itemType == "movie" {
		requestBody = generateMovieRequestBody(tmdbID)
	} else if itemType == "show" {
		requestBody = generateShowRequestBody(tmdbID)
	} else {
		return nil, logging.ErrorLog{
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
		return nil, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to send request to Mediux API"},
		}
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody modals.MediuxResponse

	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		logging.LOG.Error(fmt.Sprintf("Response error: %s", response.Body()))
		return nil, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to parse response from Mediux API"},
		}
	}

	// Check if the response status is OK
	if response.StatusCode() != http.StatusOK {
		return nil, logging.ErrorLog{
			Err: errors.New("received non-200 response from Mediux API"),
			Log: logging.Log{Message: fmt.Sprintf("Received non-200 response from Mediux API: %s", response.String())},
		}
	}

	// Check if the response is nil on all fields
	if itemType == "movie" {
		if responseBody.Data.Movie.ID == "" {
			return nil, logging.ErrorLog{
				Err: fmt.Errorf("no movies found in the response for TMDB ID: %s", tmdbID),
				Log: logging.Log{Message: "No movies found in the response"},
			}
		}

		if responseBody.Data.Movie.CollectionID == nil &&
			responseBody.Data.Movie.Posters == nil &&
			responseBody.Data.Movie.Backdrops == nil {
			return nil, logging.ErrorLog{
				Err: fmt.Errorf("no movie sets or collection found in the response for TMDB ID: %s", tmdbID),
				Log: logging.Log{Message: "No movie sets or collection found in the response"},
			}
		}

	} else if itemType == "show" {
		if responseBody.Data.Show.ID == "" {
			return nil, logging.ErrorLog{
				Err: fmt.Errorf("no shows found in the response for TMDB ID: %s", tmdbID),
				Log: logging.Log{Message: "No shows found in the response"},
			}
		}

		if responseBody.Data.Show.Posters == nil &&
			responseBody.Data.Show.Backdrops == nil &&
			responseBody.Data.Show.Seasons == nil {
			return nil, logging.ErrorLog{
				Err: fmt.Errorf("no show sets found in the response for TMDB ID: %s", tmdbID),
				Log: logging.Log{Message: "No show sets found in the response"},
			}
		}

	}

	var posterSets []modals.PosterSet
	if itemType == "movie" {
		posterSets = processMovieResponse(librarySection, itemRatingKey, (*responseBody.Data.Movie))
	} else if itemType == "show" {
		posterSets = processShowResponse((*responseBody.Data.Show))
	}

	return posterSets, logging.ErrorLog{}
}

func processShowResponse(show modals.MediuxShowByID) []modals.PosterSet {
	logging.LOG.Trace(fmt.Sprintf("Processing show: %s", show.Title))
	showSetMap := make(map[string]*modals.PosterSet)

	if len(show.Posters) > 0 {
		logging.LOG.Trace(fmt.Sprintf("Found %d posters", len(show.Posters)))
		for _, poster := range show.Posters {
			if poster.ShowSet != nil && poster.ShowSet.ID != "" {
				setInfo := poster.ShowSet
				newPoster := modals.PosterFile{
					ID:       poster.ID,
					Type:     "poster",
					Modified: poster.ModifiedOn,
					FileSize: parseFileSize(poster.FileSize),
				}

				if ps, exists := showSetMap[setInfo.ID]; exists {
					ps.Poster = &newPoster
				} else {
					newPosterSet := &modals.PosterSet{
						ID:          setInfo.ID,
						Title:       setInfo.SetTitle,
						Type:        "show",
						User:        modals.SetUser{Name: setInfo.UserCreated.Username},
						DateCreated: setInfo.DateCreated,
						DateUpdated: setInfo.DateUpdated,
						Status:      show.Status,
					}
					newPosterSet.Poster = &newPoster
					showSetMap[setInfo.ID] = newPosterSet
				}
			}
		}
	}
	if len(show.Backdrops) > 0 {
		logging.LOG.Trace(fmt.Sprintf("Found %d backdrops", len(show.Backdrops)))
		for _, backdrop := range show.Backdrops {
			if backdrop.ShowSet != nil && backdrop.ShowSet.ID != "" {
				setInfo := backdrop.ShowSet
				newBackdrop := modals.PosterFile{
					ID:       backdrop.ID,
					Type:     "backdrop",
					Modified: backdrop.ModifiedOn,
					FileSize: parseFileSize(backdrop.FileSize),
				}
				if ps, exists := showSetMap[setInfo.ID]; exists {
					ps.Backdrop = &newBackdrop
				} else {
					newPosterSet := &modals.PosterSet{
						ID:          setInfo.ID,
						Title:       setInfo.SetTitle,
						Type:        "show",
						User:        modals.SetUser{Name: setInfo.UserCreated.Username},
						DateCreated: setInfo.DateCreated,
						DateUpdated: setInfo.DateUpdated,
						Status:      show.Status,
					}
					newPosterSet.Backdrop = &newBackdrop
					showSetMap[setInfo.ID] = newPosterSet
				}
			}
		}
	}
	if len(show.Seasons) > 0 {
		logging.LOG.Trace(fmt.Sprintf("Found %d seasons", len(show.Seasons)))
		for _, season := range show.Seasons {
			for _, poster := range season.Posters {
				if poster.ShowSet != nil && poster.ShowSet.ID != "" {
					setInfo := poster.ShowSet
					var seasonType string
					if season.SeasonNumber == 0 {
						seasonType = "specialSeasonPoster"
					} else {
						seasonType = "seasonPoster"
					}
					newPoster := modals.PosterFile{
						ID:       poster.ID,
						Type:     seasonType,
						Modified: poster.ModifiedOn,
						FileSize: parseFileSize(poster.FileSize),
						Season: &modals.PosterFileSeason{
							Number: season.SeasonNumber,
						},
					}
					if ps, exists := showSetMap[setInfo.ID]; exists {
						ps.SeasonPosters = append(ps.SeasonPosters, newPoster)
					} else {
						newPosterSet := &modals.PosterSet{
							ID:          setInfo.ID,
							Title:       setInfo.SetTitle,
							Type:        "show",
							User:        modals.SetUser{Name: setInfo.UserCreated.Username},
							DateCreated: setInfo.DateCreated,
							DateUpdated: setInfo.DateUpdated,
							Status:      show.Status,
						}
						newPosterSet.SeasonPosters = []modals.PosterFile{newPoster}
						showSetMap[setInfo.ID] = newPosterSet
					}
				}
			}
			for _, episode := range season.Episodes {
				for _, titlecard := range episode.Titlecards {
					if titlecard.ShowSet != nil && titlecard.ShowSet.ID != "" {
						setInfo := titlecard.ShowSet
						newTitlecard := modals.PosterFile{
							ID:       titlecard.ID,
							Type:     "titlecard",
							Modified: titlecard.ModifiedOn,
							FileSize: parseFileSize(titlecard.FileSize),
							Episode: &modals.PosterFileEpisode{
								Title:         episode.EpisodeTitle,
								EpisodeNumber: episode.EpisodeNumber,
								SeasonNumber:  episode.Season.SeasonNumber,
							},
						}
						if ps, exists := showSetMap[setInfo.ID]; exists {
							ps.TitleCards = append(ps.TitleCards, newTitlecard)
						} else {
							newPosterSet := &modals.PosterSet{
								ID:          setInfo.ID,
								Title:       setInfo.SetTitle,
								Type:        "show",
								User:        modals.SetUser{Name: setInfo.UserCreated.Username},
								DateCreated: setInfo.DateCreated,
								DateUpdated: setInfo.DateUpdated,
								Status:      show.Status,
							}
							newPosterSet.TitleCards = []modals.PosterFile{newTitlecard}
							showSetMap[setInfo.ID] = newPosterSet
						}

					}

				}
			}
		}
	}

	// Convert the map to a slice
	var posterSets []modals.PosterSet
	for _, set := range showSetMap {
		posterSets = append(posterSets, *set)
	}

	return posterSets

}

func processMovieResponse(librarySection, itemRatingKey string, movie modals.MediuxMovieByID) []modals.PosterSet {
	var posterSets []modals.PosterSet

	if movie.CollectionID != nil {
		posterSets = append(posterSets, processMovieCollection(librarySection, movie.ID, movie.CollectionID.Movies)...)
	}
	posterSets = append(posterSets, processMovieSetPostersAndBackdrops(librarySection, itemRatingKey, movie)...)

	return posterSets
}

func processMovieSetPostersAndBackdrops(librarySection string, itemRatingKey string, movie modals.MediuxMovieByID) []modals.PosterSet {
	logging.LOG.Trace(fmt.Sprintf("Processing Movie Set for - %s", movie.Title))
	var posterSets []modals.PosterSet
	movieSetMap := make(map[string]*modals.PosterSet)

	if len(movie.Posters) > 0 {
		logging.LOG.Trace(fmt.Sprintf("Found %d posters", len(movie.Posters)))
		for _, poster := range movie.Posters {
			if poster.MovieSet != nil && poster.MovieSet.ID != "" {
				setInfo := poster.MovieSet
				newPoster := modals.PosterFile{
					ID:       poster.ID,
					Type:     "poster",
					Modified: poster.ModifiedOn,
					FileSize: parseFileSize(poster.FileSize),
					Movie: &modals.PosterFileMovie{
						ID:             movie.ID,
						Title:          movie.Title,
						Status:         movie.Status,
						Tagline:        movie.Tagline,
						Slug:           movie.Slug,
						DateUpdated:    movie.DateUpdated,
						TvdbID:         movie.TvdbID,
						ImdbID:         movie.ImdbID,
						TraktID:        movie.TraktID,
						ReleaseDate:    movie.ReleaseDate,
						RatingKey:      itemRatingKey,
						LibrarySection: librarySection,
					},
				}

				// Check to see this set already exists in the map
				if ps, exists := movieSetMap[setInfo.ID]; exists {
					ps.Poster = &newPoster
				} else {
					newPosterSet := &modals.PosterSet{
						ID:          setInfo.ID,
						Title:       setInfo.SetTitle,
						Type:        "movie",
						User:        modals.SetUser{Name: setInfo.UserCreated.Username},
						DateCreated: setInfo.DateCreated,
						DateUpdated: setInfo.DateUpdated,
						Status:      movie.Status,
					}
					newPosterSet.Poster = &newPoster
					movieSetMap[setInfo.ID] = newPosterSet
				}
			}
		}
	}

	if len(movie.Backdrops) > 0 {
		logging.LOG.Trace(fmt.Sprintf("Found %d backdrops", len(movie.Backdrops)))
		for _, backdrop := range movie.Backdrops {
			if backdrop.MovieSet != nil && backdrop.MovieSet.ID != "" {
				setInfo := backdrop.MovieSet
				newBackdrop := modals.PosterFile{
					ID:       backdrop.ID,
					Type:     "backdrop",
					Modified: backdrop.ModifiedOn,
					FileSize: parseFileSize(backdrop.FileSize),
					Movie: &modals.PosterFileMovie{
						ID:             movie.ID,
						Title:          movie.Title,
						Status:         movie.Status,
						Tagline:        movie.Tagline,
						Slug:           movie.Slug,
						DateUpdated:    movie.DateUpdated,
						TvdbID:         movie.TvdbID,
						ImdbID:         movie.ImdbID,
						TraktID:        movie.TraktID,
						ReleaseDate:    movie.ReleaseDate,
						RatingKey:      itemRatingKey,
						LibrarySection: librarySection,
					},
				}
				if ps, exists := movieSetMap[setInfo.ID]; exists {
					ps.Backdrop = &newBackdrop
				} else {
					newPosterSet := &modals.PosterSet{
						ID:          setInfo.ID,
						Title:       setInfo.SetTitle,
						Type:        "movie",
						User:        modals.SetUser{Name: setInfo.UserCreated.Username},
						DateCreated: setInfo.DateCreated,
						DateUpdated: setInfo.DateUpdated,
						Status:      movie.Status,
					}
					newPosterSet.Backdrop = &newBackdrop
					movieSetMap[setInfo.ID] = newPosterSet
				}
			}
		}
	}

	// Convert the map to a slice
	for _, set := range movieSetMap {
		posterSets = append(posterSets, *set)
	}

	return posterSets
}

func processMovieCollection(librarySection, mainMovieID string, movies []modals.MediuxMovieCollectionMovie) []modals.PosterSet {
	if len(movies) == 0 {
		return nil
	}
	logging.LOG.Trace("Processing movie collection")
	collectionSetMap := make(map[string]*modals.PosterSet)
	for _, movie := range movies {
		// fetchedRatingKey, _ := SearchForItemAndGetRatingKey(
		// 	movie.ID, "movie",
		// 	movie.Title, librarySection)
		logging.LOG.Trace(fmt.Sprintf("Processing movie: %s", movie.Title))
		if len(movie.Posters) > 0 {
			logging.LOG.Trace(fmt.Sprintf("Found %d posters", len(movie.Posters)))
			for _, poster := range movie.Posters {
				if poster.CollectionSet.ID != "" {
					setInfo := poster.CollectionSet

					newPoster := modals.PosterFile{
						ID:       poster.ID,
						Type:     "poster",
						Modified: poster.ModifiedOn,
						FileSize: parseFileSize(poster.FileSize),
						Movie: &modals.PosterFileMovie{
							ID:          movie.ID,
							Title:       movie.Title,
							Status:      movie.Status,
							Tagline:     movie.Tagline,
							Slug:        movie.Slug,
							DateUpdated: movie.DateUpdated,
							TvdbID:      movie.TvdbID,
							ImdbID:      movie.ImdbID,
							TraktID:     movie.TraktID,
							ReleaseDate: movie.ReleaseDate,
							//RatingKey:      fetchedRatingKey,
							LibrarySection: librarySection,
						},
					}

					// Check to see this set already exists in the map
					if cs, exists := collectionSetMap[setInfo.ID]; exists {
						if mainMovieID == movie.ID {
							cs.Poster = &newPoster
						} else {
							cs.OtherPosters = append(cs.OtherPosters, newPoster)
						}
					} else {
						// Create a new PosterSet
						newPosterSet := &modals.PosterSet{
							ID:          setInfo.ID,
							Title:       setInfo.SetTitle,
							Type:        "movie",
							User:        modals.SetUser{Name: setInfo.UserCreated.Username},
							DateCreated: setInfo.DateCreated,
							DateUpdated: setInfo.DateUpdated,
							Status:      movie.Status,
						}
						if mainMovieID == movie.ID {
							newPosterSet.Poster = &newPoster
						} else {
							newPosterSet.OtherPosters = append(newPosterSet.OtherPosters, newPoster)
						}
						collectionSetMap[setInfo.ID] = newPosterSet
					}

				}
			}
		}
		if len(movie.Backdrops) > 0 {
			logging.LOG.Trace(fmt.Sprintf("Found %d backdrops", len(movie.Backdrops)))
			for _, backdrop := range movie.Backdrops {
				if backdrop.CollectionSet.ID != "" {
					setInfo := backdrop.CollectionSet

					newBackdrop := modals.PosterFile{
						ID:       backdrop.ID,
						Type:     "backdrop",
						Modified: backdrop.ModifiedOn,
						FileSize: parseFileSize(backdrop.FileSize),
						Movie: &modals.PosterFileMovie{
							ID:          movie.ID,
							Title:       movie.Title,
							Status:      movie.Status,
							Tagline:     movie.Tagline,
							Slug:        movie.Slug,
							DateUpdated: movie.DateUpdated,
							TvdbID:      movie.TvdbID,
							ImdbID:      movie.ImdbID,
							TraktID:     movie.TraktID,
							ReleaseDate: movie.ReleaseDate,
							//RatingKey:      fetchedRatingKey,
							LibrarySection: librarySection,
						},
					}

					if cs, exists := collectionSetMap[setInfo.ID]; exists {
						if mainMovieID == movie.ID {
							cs.Backdrop = &newBackdrop
						} else {
							cs.OtherBackdrops = append(cs.OtherBackdrops, newBackdrop)
						}
					} else {
						newPosterSet := &modals.PosterSet{
							ID:          setInfo.ID,
							Title:       setInfo.SetTitle,
							Type:        "movie",
							User:        modals.SetUser{Name: setInfo.UserCreated.Username},
							DateCreated: setInfo.DateCreated,
							DateUpdated: setInfo.DateUpdated,
							Status:      movie.Status,
						}
						if mainMovieID == movie.ID {
							newPosterSet.Backdrop = &newBackdrop
						} else {
							newPosterSet.OtherBackdrops = append(newPosterSet.OtherBackdrops, newBackdrop)
						}
						collectionSetMap[setInfo.ID] = newPosterSet
					}
				}
			}
		}
	}

	// Convert the map to a slice
	var posterSets []modals.PosterSet
	for _, set := range collectionSetMap {
		posterSets = append(posterSets, *set)

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

func GetShowSetByID(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	// Get the set ID from the URL
	setID := chi.URLParam(r, "setID")
	if setID == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: errors.New("missing setID in URL"),
			Log: logging.Log{
				Message: "Missing setID in URL",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	logging.LOG.Debug(fmt.Sprintf("Fetching show set by ID: %s", setID))

	showSet, logErr := FetchShowSetByID(setID)
	if logErr.Err != nil {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logErr)
		return
	}

	if showSet.ID == "" {
		utils.SendErrorJSONResponse(w, http.StatusInternalServerError, logging.ErrorLog{
			Err: errors.New("no show set found"),
			Log: logging.Log{
				Message: "No show set found for the provided ID",
				Elapsed: utils.ElapsedTime(startTime),
			},
		})
		return
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Message: "Retrieved show set from Mediux",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    showSet,
	})
}

func FetchShowSetByID(setID string) (modals.PosterSet, logging.ErrorLog) {

	requestBody := generateShowSetByIDRequestBody(setID)

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

	// Check if the response status is OK
	if response.StatusCode() != http.StatusOK {
		return modals.PosterSet{}, logging.ErrorLog{
			Err: errors.New("received non-200 response from Mediux API"),
			Log: logging.Log{Message: fmt.Sprintf("Received non-200 response from Mediux API: %s", response.String())},
		}
	}

	// Parse the response body into a MediuxShowSetResponse struct
	var responseBody modals.MediuxShowSetResponse
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		return modals.PosterSet{}, logging.ErrorLog{
			Err: err,
			Log: logging.Log{Message: "Failed to parse response from Mediux API"},
		}
	}

	showSet := responseBody.Data.ShowSetID

	logging.LOG.Trace(fmt.Sprintf("Processing show set: %s", showSet.SetTitle))
	logging.LOG.Trace(fmt.Sprintf("Date Created: %s", showSet.DateCreated))
	logging.LOG.Trace(fmt.Sprintf("Date Updated: %s", showSet.DateUpdated))

	// Process the response and return the poster sets
	posterSets := processShowResponse(showSet.Show)
	posterSet := posterSets[0]

	posterSet.ID = showSet.ID
	posterSet.Title = showSet.SetTitle
	posterSet.User.Name = showSet.UserCreated.Username
	posterSet.Type = "show"
	posterSet.DateCreated = showSet.DateCreated
	posterSet.DateUpdated = showSet.DateUpdated

	return posterSet, logging.ErrorLog{}
}
