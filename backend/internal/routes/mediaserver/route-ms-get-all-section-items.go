package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"time"
)

func GetAllSectionItems(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Trace(r.URL.Path)
	startTime := time.Now()
	Err := logging.NewStandardError()

	// Get the following information from the URL
	// Section ID
	// Section Title
	// Section Type
	// Item Start Index
	sectionID := r.URL.Query().Get("sectionID")
	sectionTitle := r.URL.Query().Get("sectionTitle")
	sectionType := r.URL.Query().Get("sectionType")
	sectionStartIndex := r.URL.Query().Get("sectionStartIndex")

	// Validate the section ID, title, type, and start index
	if sectionID == "" || sectionTitle == "" || sectionType == "" || sectionStartIndex == "" {
		Err.Message = "Missing required query parameters"
		Err.HelpText = "Ensure the URL contains sectionID, sectionTitle, sectionType, and sectionStartIndex query parameters."
		Err.Details = map[string]any{
			"sectionID":         sectionID,
			"sectionTitle":      sectionTitle,
			"sectionType":       sectionType,
			"sectionStartIndex": sectionStartIndex,
		}
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	// Fetch the section items from the media server
	section, Err := api.CallFetchLibrarySectionItems(sectionID, sectionTitle, sectionType, sectionStartIndex)
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data:    section,
	})
}
