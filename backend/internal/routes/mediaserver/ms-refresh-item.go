package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func RefreshMediaItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Refresh Media Item", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	var req struct {
		RatingKey    string `json:"ratingKey"`
		Title        string `json:"title"`
		LibraryTitle string `json:"libraryTitle"`
	}
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &req, "refreshMediaItemRequest")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, nil, Err)
		return
	}

	Err = api.CallRefreshMediaItem(ctx, req.RatingKey)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, nil, Err)
		return
	}

	api.Util_Response_SendJSON(w, ld, map[string]any{
		"refreshed": true,
	})
}
