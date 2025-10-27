package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"fmt"
	"net/http"
	"time"
)

func GetAllSections(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	allSections, Err := api.CallFetchLibrarySectionInfo()
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	if len(allSections) == 0 {
		Err.Message = "No sections found"
		Err.HelpText = fmt.Sprintf("Ensure that the media server has sections configured for %s.", api.Global_Config.MediaServer.Type)
		Err.Details = fmt.Sprintf("No sections found in %s for the configured libraries", api.Global_Config.MediaServer.Type)
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Respond with a success message
	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    allSections,
	})
}
