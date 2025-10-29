package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"strings"
)

func GetAllLibrariesOptions(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get All Libraries Options", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the media server information from the request
	var mediaServerInfo api.Config_MediaServer
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &mediaServerInfo, "mediaServerInfo")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, nil, Err)
		return
	}

	if strings.HasPrefix(mediaServerInfo.Token, "***") {
		mediaServerInfo.Token = api.Global_Config.MediaServer.Token
	}

	allSections, Err := api.CallFetchLibrarySectionOptions(ctx, mediaServerInfo)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, nil, Err)
		return
	}

	if len(allSections) == 0 {
		logAction.SetError("No library sections found",
			"Ensure that the Media Server has library sections configured",
			map[string]any{
				"MediaServerInfo": mediaServerInfo,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	} else {
		logAction.AppendResult("library_sections_count", len(allSections))
		logAction.AppendResult("library_sections", strings.Join(allSections, ", "))
	}

	api.Util_Response_SendJSON(w, ld, allSections)
}
