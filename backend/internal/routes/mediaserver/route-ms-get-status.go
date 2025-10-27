package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"time"
)

/*
Route_MS_GetStatus

Checks the status of the configured media server and returns its version number.

Takes no parameters.

Method: GET

Endpoint: /health/status/mediaserver

Example Response:

	{
	    Status: "success",
	    Elapsed: "15ms",
	    Data: { Media Server Version Number }
	}

In case of an error, an error response is returned with details about the failure.
*/
func GetStatus(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	mediaServer, Err := api.GetMediaServerInterface(api.Config_MediaServer{})
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	mediaServerConfig := api.Global_Config.MediaServer

	status, Err := mediaServer.GetMediaServerStatus(mediaServerConfig)
	if Err.Message != "" {
		logging.LOG.Warn(Err.Message)
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    status,
	})
}
