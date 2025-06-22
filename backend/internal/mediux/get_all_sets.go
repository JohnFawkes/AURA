package mediux

import (
	"aura/internal/cache"
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
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

	// Create a new StandardError with details about the error
	Err := logging.NewStandardError()

	// Get the TMDB ID from the URL
	tmdbID := chi.URLParam(r, "tmdbID")
	itemType := chi.URLParam(r, "itemType")
	librarySection := chi.URLParam(r, "librarySection")
	itemRatingKey := chi.URLParam(r, "ratingKey")
	if tmdbID == "" || itemType == "" || librarySection == "" {

		Err.Message = "Missing TMDB ID, Item Type or Library Section in URL Parameters"
		Err.HelpText = "Ensure the TMDB ID, Item Type and Library Section are provided in path parameters."
		Err.Details = fmt.Sprintf("Params: tmdbID=%s, itemType=%s, librarySection=%s", tmdbID, itemType, librarySection)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// If cache is empty, return false
	if cache.LibraryCacheStore.IsEmpty() {

		Err.Message = "Backend cache is empty"
		Err.HelpText = "Try refreshing the cache from the Home Page"
		Err.Details = "This typically happens when the backend cache is not initialized or has been cleared. Example on application restart."
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	logging.LOG.Debug(fmt.Sprintf("Fetching all sets for TMDB ID: %s, item type: %s, library section: %s", tmdbID, itemType, librarySection))

	posterSets, Err := fetchAllSets(tmdbID, itemType, librarySection, itemRatingKey)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	if len(posterSets) == 0 {
		Err.Message = "No sets found for the provided TMDB ID and Item Type"
		Err.HelpText = "Ensure the TMDB ID and Item Type are correct and that sets exist for this item."
		Err.Details = map[string]any{
			"tmdbID":         tmdbID,
			"itemType":       itemType,
			"librarySection": librarySection,
		}
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    posterSets,
	})
}

func fetchAllSets(tmdbID, itemType, librarySection, itemRatingKey string) ([]modals.PosterSet, logging.StandardError) {
	Err := logging.NewStandardError()
	Err.Function = utils.GetFunctionName()

	// Generate the request body
	var requestBody map[string]any
	switch itemType {
	case "movie":
		requestBody = generateMovieRequestBody(tmdbID)
	case "show":
		requestBody = generateShowRequestBody(tmdbID)
	default:
		Err.Message = "Invalid item type provided"
		Err.HelpText = "Ensure the item type is either 'movie' or 'show'."
		Err.Details = fmt.Sprintf("Provided item type: %s", itemType)
		return nil, Err
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

		Err.Message = "Failed to make request to Mediux API"
		Err.HelpText = "Ensure the Mediux API is reachable and the token is valid."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody modals.MediuxResponse

	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {

		Err.Message = "Failed to parse Mediux API response"
		Err.HelpText = "Ensure the Mediux API is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s, Response: %s", err.Error(), response.Body())
		return nil, Err
	}

	// Check if the response is nil on all fields
	switch itemType {
	case "movie":
		if responseBody.Data.Movie.ID == "" {

			Err.Message = "Movie not found in the response"
			Err.HelpText = "Ensure the TMDB ID is correct and the movie exists in the Mediux database."
			Err.Details = fmt.Sprintf("TMDB ID: %s", tmdbID)
			return nil, Err
		}

		if responseBody.Data.Movie.CollectionID == nil &&
			responseBody.Data.Movie.Posters == nil &&
			responseBody.Data.Movie.Backdrops == nil {

			Err.Message = "Movie sets not found in the response"
			Err.HelpText = "Ensure the TMDB ID is correct and the movie has sets in the Mediux database."
			Err.Details = fmt.Sprintf("TMDB ID: %s", tmdbID)
			return nil, Err
		}

	case "show":
		if responseBody.Data.Show.ID == "" {

			Err.Message = "Show not found in the response"
			Err.HelpText = "Ensure the TMDB ID is correct and the show exists in the Mediux database."
			Err.Details = fmt.Sprintf("TMDB ID: %s", tmdbID)
			return nil, Err
		}

		if responseBody.Data.Show.Posters == nil &&
			responseBody.Data.Show.Backdrops == nil &&
			responseBody.Data.Show.Seasons == nil {

			Err.Message = "Show sets not found in the response"
			Err.HelpText = "Ensure the TMDB ID is correct and the show has sets in the Mediux database."
			Err.Details = fmt.Sprintf("TMDB ID: %s", tmdbID)
			return nil, Err
		}
	}

	var posterSets []modals.PosterSet
	switch itemType {
	case "movie":
		posterSets = processMovieResponse(librarySection, itemRatingKey, (*responseBody.Data.Movie))
	case "show":
		posterSets = processShowResponse(librarySection, itemRatingKey, (*responseBody.Data.Show))
	}

	return posterSets, logging.StandardError{}
}

func processShowResponse(librarySection, itemRatingKey string, show modals.MediuxShowByID) []modals.PosterSet {
	logging.LOG.Trace(fmt.Sprintf("Processing Show Set for - %s", show.Title))
	showSetMap := make(map[string]*modals.PosterSet)

	cachedItem, exists := cache.LibraryCacheStore.GetMediaItemFromSection(librarySection, itemRatingKey)
	if !exists {
		logging.LOG.Error(fmt.Sprintf("\tCould not find '%s' in cache", show.Title))
		return nil
	}

	if len(show.Posters) > 0 {
		logging.LOG.Trace(fmt.Sprintf("\tFound %d posters", len(show.Posters)))
		for _, poster := range show.Posters {
			if poster.ShowSet != nil && poster.ShowSet.ID != "" {
				setInfo := poster.ShowSet
				newPoster := modals.PosterFile{
					ID:       poster.ID,
					Type:     "poster",
					Modified: poster.ModifiedOn,
					FileSize: parseFileSize(poster.FileSize),
					Show: &modals.PosterFileShow{
						ID:             show.ID,
						Title:          show.Title,
						RatingKey:      itemRatingKey,
						LibrarySection: librarySection,
						MediaItem:      *cachedItem,
					},
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
		logging.LOG.Trace(fmt.Sprintf("\tFound %d backdrops", len(show.Backdrops)))
		for _, backdrop := range show.Backdrops {
			if backdrop.ShowSet != nil && backdrop.ShowSet.ID != "" {
				setInfo := backdrop.ShowSet
				newBackdrop := modals.PosterFile{
					ID:       backdrop.ID,
					Type:     "backdrop",
					Modified: backdrop.ModifiedOn,
					FileSize: parseFileSize(backdrop.FileSize),
					Show: &modals.PosterFileShow{
						ID:             show.ID,
						Title:          show.Title,
						RatingKey:      itemRatingKey,
						LibrarySection: librarySection,
						MediaItem:      *cachedItem,
					},
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
		var totalSeasonPosters, totalTitlecards int
		for _, season := range show.Seasons {
			totalSeasonPosters += len(season.Posters)
			for _, episode := range season.Episodes {
				totalTitlecards += len(episode.Titlecards)
			}
		}
		logging.LOG.Trace(fmt.Sprintf("\tFound %d season posters", totalSeasonPosters))
		logging.LOG.Trace(fmt.Sprintf("\tFound %d titlecards", totalTitlecards))
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
					// Check if the Season exists in the cachedItem
					seasonExists := false
					for _, cachedSeason := range cachedItem.Series.Seasons {
						if cachedSeason.SeasonNumber == season.SeasonNumber {
							seasonExists = true
							break
						}
					}
					var seasonItem *modals.MediaItem
					if !seasonExists {
						seasonItem = &modals.MediaItem{} // Create a new empty MediaItem
					} else {
						seasonItem = cachedItem // Use existing cached item
					}
					newPoster := modals.PosterFile{
						ID:       poster.ID,
						Type:     seasonType,
						Modified: poster.ModifiedOn,
						FileSize: parseFileSize(poster.FileSize),
						Show: &modals.PosterFileShow{
							ID:             show.ID,
							Title:          show.Title,
							RatingKey:      itemRatingKey,
							LibrarySection: librarySection,
							MediaItem:      *seasonItem,
						},
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

						// Check if the Episode exists in the cachedItem
						episodeExists := false
						for _, cachedSeason := range cachedItem.Series.Seasons {
							for _, cachedEpisode := range cachedSeason.Episodes {
								if cachedEpisode.SeasonNumber == episode.Season.SeasonNumber &&
									cachedEpisode.EpisodeNumber == episode.EpisodeNumber {
									episodeExists = true
									break
								}
							}
						}
						var episodeItem *modals.MediaItem
						if !episodeExists {
							episodeItem = &modals.MediaItem{} // Create a new empty MediaItem
						} else {
							episodeItem = cachedItem // Use existing cached item
						}

						newTitlecard := modals.PosterFile{
							ID:       titlecard.ID,
							Type:     "titlecard",
							Modified: titlecard.ModifiedOn,
							FileSize: parseFileSize(titlecard.FileSize),
							Show: &modals.PosterFileShow{
								ID:             show.ID,
								Title:          show.Title,
								RatingKey:      itemRatingKey,
								LibrarySection: librarySection,
								MediaItem:      *episodeItem,
							},
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
		posterSets = append(posterSets, processMovieCollection(itemRatingKey, librarySection, movie.ID, *movie.CollectionID)...)
	}
	posterSets = append(posterSets, processMovieSetPostersAndBackdrops(librarySection, itemRatingKey, movie)...)

	return posterSets
}

func processMovieSetPostersAndBackdrops(librarySection string, itemRatingKey string, movie modals.MediuxMovieByID) []modals.PosterSet {
	logging.LOG.Trace(fmt.Sprintf("Processing Movie Set for - %s", movie.Title))
	var posterSets []modals.PosterSet
	movieSetMap := make(map[string]*modals.PosterSet)

	cachedItem, exists := cache.LibraryCacheStore.GetMediaItemFromSection(librarySection, itemRatingKey)
	if !exists {
		logging.LOG.Error(fmt.Sprintf("\tCould not find '%s' in cache", movie.Title))
		return nil
	}

	if len(movie.Posters) > 0 {
		logging.LOG.Trace(fmt.Sprintf("\tFound %d posters", len(movie.Posters)))
		for _, poster := range movie.Posters {
			if poster.MovieSet != nil && poster.MovieSet.ID != "" {
				setInfo := poster.MovieSet
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
						MediaItem:   *cachedItem,
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
		logging.LOG.Trace(fmt.Sprintf("\tFound %d backdrops", len(movie.Backdrops)))
		for _, backdrop := range movie.Backdrops {
			if backdrop.MovieSet != nil && backdrop.MovieSet.ID != "" {
				setInfo := backdrop.MovieSet
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
						MediaItem:   *cachedItem,
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

func processMovieCollection(itemRatingKey, librarySection, mainMovieID string, collection modals.MediuxMovieCollectionID) []modals.PosterSet {
	movies := collection.Movies
	if len(movies) == 0 {
		return nil
	}
	logging.LOG.Trace(fmt.Sprintf("Processing %d movies in '%s' collection", len(movies), collection.CollectionName))
	collectionSetMap := make(map[string]*modals.PosterSet)
	for _, movie := range movies {
		logging.LOG.Trace(fmt.Sprintf("\tProcessing movie: %s", movie.Title))

		cachedItem, _ := cache.LibraryCacheStore.GetMediaItemFromSectionByTMDBID(librarySection, movie.ID)
		if cachedItem == nil {
			logging.LOG.Error(fmt.Sprintf("\t\tCould not find '%s' in cache", movie.Title))
		}

		if len(movie.Posters) > 0 {
			logging.LOG.Trace(fmt.Sprintf("\t\tFound %d posters", len(movie.Posters)))
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
							MediaItem:   *cachedItem,
						},
					}

					// Check to see this set already exists in the map
					if cs, exists := collectionSetMap[setInfo.ID]; exists {
						if mainMovieID == movie.ID {
							newPoster.Movie.RatingKey = itemRatingKey
							cs.Poster = &newPoster
						} else {
							cs.OtherPosters = append(cs.OtherPosters, newPoster)
						}
					} else {
						// Create a new PosterSet
						newPosterSet := &modals.PosterSet{
							ID:          setInfo.ID,
							Title:       setInfo.SetTitle,
							Type:        "collection",
							User:        modals.SetUser{Name: setInfo.UserCreated.Username},
							DateCreated: setInfo.DateCreated,
							DateUpdated: setInfo.DateUpdated,
							Status:      movie.Status,
						}
						if mainMovieID == movie.ID {
							newPoster.Movie.RatingKey = itemRatingKey
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
			logging.LOG.Trace(fmt.Sprintf("\t\tFound %d backdrops", len(movie.Backdrops)))
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
							MediaItem:   *cachedItem,
						},
					}

					if cs, exists := collectionSetMap[setInfo.ID]; exists {
						if mainMovieID == movie.ID {
							newBackdrop.Movie.RatingKey = itemRatingKey
							cs.Backdrop = &newBackdrop
						} else {
							cs.OtherBackdrops = append(cs.OtherBackdrops, newBackdrop)
						}
					} else {
						newPosterSet := &modals.PosterSet{
							ID:          setInfo.ID,
							Title:       setInfo.SetTitle,
							Type:        "collection",
							User:        modals.SetUser{Name: setInfo.UserCreated.Username},
							DateCreated: setInfo.DateCreated,
							DateUpdated: setInfo.DateUpdated,
							Status:      movie.Status,
						}
						if mainMovieID == movie.ID {
							newBackdrop.Movie.RatingKey = itemRatingKey
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

func GetSetByID(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	// Get the set ID from the URL
	setID := chi.URLParam(r, "setID")
	if setID == "" {

		Err.Message = "Missing setID in URL"
		Err.HelpText = "Ensure the setID is provided in the URL path."
		Err.Details = fmt.Sprintf("URL Path: %s", r.URL.Path)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	// Get the librarySection and itemRatingKey from the query parameters
	librarySection := r.URL.Query().Get("librarySection")
	itemRatingKey := r.URL.Query().Get("itemRatingKey")
	itemType := r.URL.Query().Get("itemType")
	if librarySection == "" || itemRatingKey == "" || itemType == "" {
		Err.Function = utils.GetFunctionName()

		Err.Message = "Missing librarySection, itemRatingKey or itemType in query parameters"
		Err.HelpText = "Ensure the librarySection, itemRatingKey and itemType are provided in query parameters."
		Err.Details = fmt.Sprintf("Query Params: librarySection=%s, itemRatingKey=%s, itemType=%s", librarySection, itemRatingKey, itemType)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	var updatedSet modals.PosterSet

	if itemType == "show" {
		updatedSet, Err = FetchShowSetByID(librarySection, itemRatingKey, setID)
		if Err.Message != "" {
			utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
			return
		}
		if updatedSet.ID == "" {
			Err.Function = utils.GetFunctionName()

			Err.Message = "No show set found for the provided ID"
			Err.HelpText = "Ensure the set ID is correct and the show set exists in the Mediux database."
			Err.Details = fmt.Sprintf("Set ID: %s, Library Section: %s, Item Rating Key: %s", setID, librarySection, itemRatingKey)
			utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
			return
		}
	} else if itemType == "movie" {
		updatedSet, Err = FetchMovieSetByID(librarySection, itemRatingKey, setID)
		if Err.Message != "" {
			utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
			return
		}
		if updatedSet.ID == "" {
			Err.Function = utils.GetFunctionName()

			Err.Message = "No movie set found for the provided ID"
			Err.HelpText = "Ensure the set ID is correct and the movie set exists in the Mediux database."
			Err.Details = fmt.Sprintf("Set ID: %s, Library Section: %s, Item Rating Key: %s", setID, librarySection, itemRatingKey)
			utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
			return
		}
	} else if itemType == "collection" {
		updatedSet, Err = FetchCollectionSetByID(librarySection, itemRatingKey, setID)
		if Err.Message != "" {
			utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
			return
		}
		if updatedSet.ID == "" {
			Err.Function = utils.GetFunctionName()

			Err.Message = "No collection set found for the provided ID"
			Err.HelpText = "Ensure the set ID is correct and the collection set exists in the Mediux database."
			Err.Details = fmt.Sprintf("Set ID: %s, Library Section: %s, Item Rating Key: %s", setID, librarySection, itemRatingKey)
			utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
			return
		}
	} else {
		Err.Function = utils.GetFunctionName()

		Err.Message = "Invalid item type provided"
		Err.HelpText = "Ensure the item type is either 'movie', 'show' or 'collection'."
		Err.Details = fmt.Sprintf("Provided item type: %s", itemType)
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    updatedSet,
	})
}

func FetchShowSetByID(librarySection, itemRatingKey, setID string) (modals.PosterSet, logging.StandardError) {

	Err := logging.NewStandardError()
	Err.Function = utils.GetFunctionName()
	requestBody := generateShowSetByIDRequestBody(setID)

	// If cache is empty, return false
	if cache.LibraryCacheStore.IsEmpty() {

		Err.Message = "Backend cache is empty"
		Err.HelpText = "Try refreshing the cache from the Home Page"
		Err.Details = "This typically happens when the backend cache is not initialized or has been cleared. Example on application restart."
		return modals.PosterSet{}, Err
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
		Err.Function = utils.GetFunctionName()

		Err.Message = "Failed to make request to Mediux API"
		Err.HelpText = "Ensure the Mediux API is reachable and the token is valid."
		Err.Details = fmt.Sprintf("Error: %s, Response: %s", err.Error(), response.Body())
		return modals.PosterSet{}, Err
	}

	// Parse the response body into a MediuxShowSetResponse struct
	var responseBody modals.MediuxShowSetResponse
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		Err.Function = utils.GetFunctionName()

		Err.Message = "Failed to parse Mediux API response"
		Err.HelpText = "Ensure the Mediux API is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s, Response: %s", err.Error(), response.Body())
		return modals.PosterSet{}, Err
	}

	showSet := responseBody.Data.ShowSetID

	logging.LOG.Trace(fmt.Sprintf("Processing show set: %s", showSet.SetTitle))
	logging.LOG.Trace(fmt.Sprintf("Date Created: %s", showSet.DateCreated))
	logging.LOG.Trace(fmt.Sprintf("Date Updated: %s", showSet.DateUpdated))

	// Process the response and return the poster sets
	posterSets := processShowResponse(librarySection, itemRatingKey, showSet.Show)
	posterSet := posterSets[0]

	posterSet.ID = showSet.ID
	posterSet.Title = showSet.SetTitle
	posterSet.User.Name = showSet.UserCreated.Username
	posterSet.Type = "show"
	posterSet.DateCreated = showSet.DateCreated
	posterSet.DateUpdated = showSet.DateUpdated

	return posterSet, logging.StandardError{}
}

func FetchMovieSetByID(librarySection, itemRatingKey, setID string) (modals.PosterSet, logging.StandardError) {
	Err := logging.NewStandardError()
	Err.Function = utils.GetFunctionName()
	requestBody := generateMovieSetByIDRequestBody(setID)

	// If cache is empty, return false
	if cache.LibraryCacheStore.IsEmpty() {

		Err.Message = "Backend cache is empty"
		Err.HelpText = "Try refreshing the cache from the Home Page"
		Err.Details = "This typically happens when the backend cache is not initialized or has been cleared. Example on application restart."
		return modals.PosterSet{}, Err
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
		Err.Function = utils.GetFunctionName()

		Err.Message = "Failed to make request to Mediux API"
		Err.HelpText = "Ensure the Mediux API is reachable and the token is valid."
		Err.Details = fmt.Sprintf("Error: %s, Response: %s", err.Error(), response.Body())
		return modals.PosterSet{}, Err
	}

	// Parse the response body into a MediuxMovieSetResponse struct
	var responseBody modals.MediuxMovieSetResponse
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		Err.Function = utils.GetFunctionName()

		Err.Message = "Failed to parse Mediux API response"
		Err.HelpText = "Ensure the Mediux API is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s, Response: %s", err.Error(), response.Body())
		return modals.PosterSet{}, Err
	}

	movieSet := responseBody.Data.MovieSetID
	logging.LOG.Trace(fmt.Sprintf("Processing movie set: %s", movieSet.SetTitle))
	logging.LOG.Trace(fmt.Sprintf("Date Created: %s", movieSet.DateCreated))
	logging.LOG.Trace(fmt.Sprintf("Date Updated: %s", movieSet.DateUpdated))

	// Process the response and return the poster sets
	posterSets := processMovieSetPostersAndBackdrops(librarySection, itemRatingKey, movieSet.Movie)
	if len(posterSets) == 0 {
		Err.Function = utils.GetFunctionName()

		Err.Message = "No poster sets found for the provided movie set ID"
		Err.HelpText = "Ensure the movie set ID is correct and the movie set exists in the Mediux database."
		Err.Details = fmt.Sprintf("Movie Set ID: %s, Library Section: %s, Item Rating Key: %s", setID, librarySection, itemRatingKey)
		return modals.PosterSet{}, Err
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

func FetchCollectionSetByID(librarySection, itemRatingKey, setID string) (modals.PosterSet, logging.StandardError) {
	Err := logging.NewStandardError()
	Err.Function = utils.GetFunctionName()

	// If cache is empty, return false
	if cache.LibraryCacheStore.IsEmpty() {

		Err.Message = "Backend cache is empty"
		Err.HelpText = "Try refreshing the cache from the Home Page"
		Err.Details = "This typically happens when the backend cache is not initialized or has been cleared. Example on application restart."
		return modals.PosterSet{}, Err
	}

	tmdbID, exists := cache.LibraryCacheStore.GetTMDBIDFromMediaItemRatingKey(librarySection, itemRatingKey)
	if !exists {

		Err.Message = "TMDB ID not found in cache"
		Err.HelpText = "Ensure the itemRatingKey is valid and the TMDB ID exists in the cache."
		Err.Details = fmt.Sprintf("Item Rating Key: %s, Library Section: %s", itemRatingKey, librarySection)
		return modals.PosterSet{}, Err
	}

	requestBody := generateCollectionSetByIDRequestBody(setID, tmdbID)
	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.Global.Mediux.Token)).
		SetBody(requestBody).
		Post("https://staged.mediux.io/graphql")
	if err != nil {
		Err.Function = utils.GetFunctionName()

		Err.Message = "Failed to make request to Mediux API"
		Err.HelpText = "Ensure the Mediux API is reachable and the token is valid."
		Err.Details = fmt.Sprintf("Error: %s, Response: %s", err.Error(), response.Body())
		return modals.PosterSet{}, Err
	}

	// Parse the response body into a MediuxMovieSetResponse struct
	var responseBody modals.MediuxCollectionSetResponse
	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		Err.Function = utils.GetFunctionName()

		Err.Message = "Failed to parse Mediux API response"
		Err.HelpText = "Ensure the Mediux API is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s, Response: %s", err.Error(), response.Body())
		return modals.PosterSet{}, Err
	}

	collectionSet := responseBody.Data.CollectionSetID
	logging.LOG.Trace(fmt.Sprintf("Processing collection set: %s", collectionSet.SetTitle))
	logging.LOG.Trace(fmt.Sprintf("Date Created: %s", collectionSet.DateCreated))
	logging.LOG.Trace(fmt.Sprintf("Date Updated: %s", collectionSet.DateUpdated))
	if collectionSet.ID == "" {
		Err.Function = utils.GetFunctionName()
		Err.Message = "No collection set found for the provided ID"
		Err.HelpText = "Ensure the collection set ID is correct and the collection set exists in the Mediux database."
		Err.Details = fmt.Sprintf("Collection Set ID: %s, Library Section: %s, Item Rating Key: %s", setID, librarySection, itemRatingKey)
		return modals.PosterSet{}, Err
	}

	// Process the response and return the poster sets
	posterSets := processMovieCollection(itemRatingKey, librarySection, tmdbID, collectionSet.Collection)
	if len(posterSets) == 0 {
		Err.Function = utils.GetFunctionName()

		Err.Message = "No poster sets found for the provided collection set ID"
		Err.HelpText = "Ensure the collection set ID is correct and the collection set exists in the Mediux database."
		Err.Details = fmt.Sprintf("Collection Set ID: %s, Library Section: %s, Item Rating Key: %s", setID, librarySection, itemRatingKey)
		return modals.PosterSet{}, logging.StandardError{}
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
