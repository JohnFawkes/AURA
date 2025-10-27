package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// Route_Mediux_GetAllUserSets handles the API request to fetch all user-defined sets from Mediux for a specific user.
//
// Takes in the username as a URL parameter.
//
/* Returns a JSON response with the following structure:
{
	ShowSets       []MediuxUserShowSet       `json:"show_sets"`
	MovieSets      []MediuxUserMovieSet      `json:"movie_sets"`
	CollectionSets []MediuxUserCollectionSet `json:"collection_sets"`
	Boxsets        []MediuxUserBoxset        `json:"boxsets"`
}
*/
func GetAllUserSets(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)
	Err := logging.NewStandardError()

	// Get the username from the URL
	username := chi.URLParam(r, "username")
	if username == "" {
		Err.Message = "Missing username in URL"
		Err.HelpText = "Ensure the username is provided in the URL path."
		Err.Details = map[string]any{
			"url":      r.URL.Path,
			"method":   r.Method,
			"username": username,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	allSetsResponse, Err := api.Mediux_FetchAllUserSets(username)
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    allSetsResponse.Data,
	})
}
