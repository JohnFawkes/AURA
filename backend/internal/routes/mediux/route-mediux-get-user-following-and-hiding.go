package routes_mediux

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"time"
)

// Route_Mediux_GetUserFollowingAndHiding handles the API request to fetch the user's following and hiding preferences from Mediux.
//
// Takes no parameters.
//
/* Returns a JSON response with the following structure:
{
	Follows []MediuxUserFollowHideUserInfo `json:"Follows"`
	Hides   []MediuxUserFollowHideUserInfo `json:"Hides"`
}
*/
func GetUserFollowingAndHiding(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	// Fetch user following and hiding data from the Mediux API
	userFollowHide, Err := api.Mediux_FetchUserFollowingAndHiding()
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    userFollowHide,
	})
}
