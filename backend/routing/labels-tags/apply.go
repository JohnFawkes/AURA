package routes_labels_tags

import (
	"aura/logging"
	"aura/mediaserver"
	"aura/models"
	sonarr_radarr "aura/sonarr-radarr"
	"aura/utils/httpx"
	"net/http"
)

func ApplyLabelsTags(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Apply Labels/Tags", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse and validate request body
	var req struct {
		MediaItem     models.MediaItem     `json:"media_item"`
		SelectedTypes models.SelectedTypes `json:"selected_types"`
	}
	Err := httpx.DecodeRequestBodyToJSON(ctx, r.Body, &req, "Apply Labels/Tags - Decode Request Body")
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Make sure that Media Item has all of the required fields
	if req.MediaItem.Title == "" || req.MediaItem.LibraryTitle == "" || req.MediaItem.TMDB_ID == "" || req.MediaItem.RatingKey == "" {
		logAction.SetError("Invalid Media Item structure",
			"Ensure that the request body contains a valid Media Item with all required fields",
			map[string]any{
				"media_item": req.MediaItem,
			},
		)
		httpx.SendResponse(w, ld, nil)
		return
	}

	mediaserver.AddLabelToMediaItem(ctx, req.MediaItem, req.SelectedTypes)
	Err = sonarr_radarr.HandleTags(ctx, req.MediaItem, req.SelectedTypes)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, "ok")
}
