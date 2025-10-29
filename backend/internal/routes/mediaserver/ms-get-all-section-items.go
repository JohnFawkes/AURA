package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetAllSectionItems(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get All Section Items", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	actionGetQueryParams := logAction.AddSubAction("Get all query params", logging.LevelTrace)
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
		actionGetQueryParams.SetError("Missing Query Parameters", "One or more required query parameters are missing",
			map[string]any{
				"sectionID":         sectionID,
				"sectionTitle":      sectionTitle,
				"sectionType":       sectionType,
				"sectionStartIndex": sectionStartIndex,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}
	actionGetQueryParams.Complete()

	// Fetch the section items from the media server
	section, Err := api.CallFetchLibrarySectionItems(ctx, sectionID, sectionTitle, sectionType, sectionStartIndex)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, nil, Err)
		return
	}

	api.Util_Response_SendJSON(w, ld, section)
}
