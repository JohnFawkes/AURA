package mediux

import (
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"context"
	_ "embed"
	"fmt"
)

//go:embed gen_item_movie.graphql
var queryMovieInfoByTMDB_ID string

//go:embed gen_item_show.graphql
var queryShowInfoByTMDB_ID string

type itemMovieByTMDB_ID_Response struct {
	Data   itemMovieByTMDB_ID_Data `json:"data"`
	Errors []ErrorResponse         `json:"errors,omitempty"`
}

type itemMovieByTMDB_ID_Data struct {
	Movie BaseMovieInfo `json:"movies_by_id"`
}

type itemShowByTMDB_ID_Response struct {
	Data   itemShowByTMDB_ID_Data `json:"data"`
	Errors []ErrorResponse        `json:"errors,omitempty"`
}

type itemShowByTMDB_ID_Data struct {
	Show BaseShowInfo `json:"shows_by_id"`
}

func GetBaseItemInfoByTMDB_ID(tmdbID string, itemType string) (itemInfo models.BaseMediuxItemInfo, Err logging.LogErrorInfo) {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Get Base Item Info By TMDB ID")
	defer ld.Log()
	logAction := ld.AddAction(fmt.Sprintf("Getting Base Item Info for TMDB ID: %s", tmdbID), logging.LevelInfo)
	defer logAction.Complete()
	ctx = logging.WithCurrentAction(ctx, logAction)

	itemInfo = models.BaseMediuxItemInfo{}

	switch itemType {
	case "movie":
		// Send the GraphQL request for movie
		resp, Err := makeGraphQLRequest(ctx, MediuxGraphQLQueryBody{
			Query:     queryMovieInfoByTMDB_ID,
			Variables: map[string]any{"tmdb_id": tmdbID},
			QueryName: "getMovieInfoByTMDB_ID",
		})
		if Err.Message != "" {
			return itemInfo, Err
		}

		// Decode the response
		var movieResponse itemMovieByTMDB_ID_Response
		Err = httpx.DecodeResponseToJSON(ctx, resp.Body(), &movieResponse, "MediUX Movie By TMDB ID Response")
		if Err.Message != "" {
			return itemInfo, Err
		}

		// If any errors returned from MediUX, log and return
		if len(movieResponse.Errors) > 0 {
			logAction.SetError("Errors returned from MediUX API", "Review the errors for more details",
				map[string]any{
					"tmdb_id":   tmdbID,
					"item_type": "movie",
					"errors":    movieResponse.Errors,
				})
			return itemInfo, Err
		}

		// Map the response to BaseMediuxItemInfo
		itemInfo = models.BaseMediuxItemInfo{
			TMDB_ID:           movieResponse.Data.Movie.ID,
			Type:              "movie",
			DateUpdated:       movieResponse.Data.Movie.DateUpdated,
			Status:            movieResponse.Data.Movie.Status,
			Title:             movieResponse.Data.Movie.Title,
			Tagline:           movieResponse.Data.Movie.Tagline,
			ReleaseDate:       movieResponse.Data.Movie.ReleaseDate,
			ImdbID:            movieResponse.Data.Movie.ImdbID,
			TraktID:           movieResponse.Data.Movie.TraktID,
			Slug:              movieResponse.Data.Movie.Slug,
			TMDB_PosterPath:   movieResponse.Data.Movie.PosterPath,
			TMDB_BackdropPath: movieResponse.Data.Movie.BackdropPath,
		}

	case "show":
		// Send the GraphQL request for show
		resp, Err := makeGraphQLRequest(ctx, MediuxGraphQLQueryBody{
			Query:     queryShowInfoByTMDB_ID,
			Variables: map[string]any{"tmdb_id": tmdbID},
			QueryName: "getShowInfoByTMDB_ID",
		})
		if Err.Message != "" {
			return itemInfo, Err
		}

		// Decode the response
		var showResponse itemShowByTMDB_ID_Response
		Err = httpx.DecodeResponseToJSON(ctx, resp.Body(), &showResponse, "MediUX Show By TMDB ID Response")
		if Err.Message != "" {
			return itemInfo, Err
		}

		// If any errors returned from MediUX, log and return
		if len(showResponse.Errors) > 0 {
			logAction.SetError("Errors returned from MediUX API", "Review the errors for more details",
				map[string]any{
					"tmdb_id":   tmdbID,
					"item_type": "show",
					"errors":    showResponse.Errors,
				})
			return itemInfo, Err
		}

		// Map the response to BaseMediuxItemInfo
		itemInfo = models.BaseMediuxItemInfo{
			TMDB_ID:           showResponse.Data.Show.ID,
			Type:              "show",
			DateUpdated:       showResponse.Data.Show.DateUpdated,
			Status:            showResponse.Data.Show.Status,
			Title:             showResponse.Data.Show.Title,
			Tagline:           showResponse.Data.Show.Tagline,
			TvdbID:            showResponse.Data.Show.TvdbID,
			ImdbID:            showResponse.Data.Show.ImdbID,
			TraktID:           showResponse.Data.Show.TraktID,
			Slug:              showResponse.Data.Show.Slug,
			TMDB_PosterPath:   showResponse.Data.Show.PosterPath,
			TMDB_BackdropPath: showResponse.Data.Show.BackdropPath,
		}

	default:
		logAction.SetError("Invalid item type provided", "Item type must be either 'movie' or 'show'",
			map[string]any{
				"tmdb_id":   tmdbID,
				"item_type": itemType,
			})
		return itemInfo, Err
	}

	return itemInfo, Err
}
