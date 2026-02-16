package mediux

import (
	"aura/logging"
	"aura/utils/httpx"
	"context"
	_ "embed"
)

//go:embed gen_tmdb_id_by_tvdb_id_movie.graphql
var queryMovieTMDBIDByTVDBID string

//go:embed gen_tmdb_id_by_tvdb_id_show.graphql
var queryShowTMDBIDByTVDBID string

func SearchTMDBIDByTVDBID(ctx context.Context, tvdbID string, itemType string) (tmdbID string, found bool, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Search TMDB ID By TVDB ID", logging.LevelTrace)
	defer logAction.Complete()

	tmdbID = ""
	found = false
	Err = logging.LogErrorInfo{}

	var query string
	switch itemType {
	case "movie":
		query = queryMovieTMDBIDByTVDBID
	case "show":
		query = queryShowTMDBIDByTVDBID
	default:
		logAction.SetError("Invalid item type for TMDB ID search", "Item type must be either 'movie' or 'show'",
			map[string]any{
				"tvdb_id":   tvdbID,
				"item_type": itemType,
			})
		return tmdbID, found, Err
	}

	// Send the GraphQL request
	resp, Err := makeGraphQLRequest(ctx, MediuxGraphQLQueryBody{
		Query:     query,
		Variables: map[string]any{"tvdb_id": tvdbID},
		QueryName: "searchTMDBIDByTVDBID",
	})
	if Err.Message != "" {
		return tmdbID, found, Err
	}

	// Decode the response
	switch itemType {
	case "movie":
		var movieResponse struct {
			Data struct {
				Movies []struct {
					ID     string `json:"id"`
					TVDBID string `json:"tvdb_id"`
				} `json:"movies"`
			} `json:"data"`
			Errors []ErrorResponse `json:"errors,omitempty"`
		}
		Err = httpx.DecodeResponseToJSON(ctx, resp.Body(), &movieResponse, "MediUX Movie TMDB ID By TVDB ID Response")
		if Err.Message != "" {
			return tmdbID, found, Err
		}

		// If any errors returned from MediUX, log and return
		if len(movieResponse.Errors) > 0 {
			logAction.SetError("Errors returned from MediUX API", "Review the errors for more details",
				map[string]any{
					"tvdb_id":   tvdbID,
					"item_type": itemType,
					"errors":    movieResponse.Errors,
				})
			return tmdbID, found, *logAction.Error
		}

		// Check if we found a matching movie
		if len(movieResponse.Data.Movies) > 0 {
			tmdbID = movieResponse.Data.Movies[0].ID
			found = true
		}
	case "show":
		var showResponse struct {
			Data struct {
				Shows []struct {
					ID     string `json:"id"`
					TVDBID string `json:"tvdb_id"`
				} `json:"shows"`
			} `json:"data"`
			Errors []ErrorResponse `json:"errors,omitempty"`
		}
		Err = httpx.DecodeResponseToJSON(ctx, resp.Body(), &showResponse, "MediUX Show TMDB ID By TVDB ID Response")
		if Err.Message != "" {
			return tmdbID, found, Err
		}

		// If any errors returned from MediUX, log and return
		if len(showResponse.Errors) > 0 {
			logAction.SetError("Errors returned from MediUX API", "Review the errors for more details",
				map[string]any{
					"tvdb_id":   tvdbID,
					"item_type": itemType,
					"errors":    showResponse.Errors,
				})
			return tmdbID, found, *logAction.Error
		}

		// Check if we found a matching show
		if len(showResponse.Data.Shows) > 0 {
			tmdbID = showResponse.Data.Shows[0].ID
			found = true
		}
	}

	return tmdbID, found, Err
}
