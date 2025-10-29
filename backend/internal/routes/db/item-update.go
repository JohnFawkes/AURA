package routes_db

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Update Item In Database", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse the request body to get the DBMediaItemWithPosterSets
	var saveItem api.DBMediaItemWithPosterSets
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &saveItem, "DBMediaItemWithPosterSets")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	for _, posterSet := range saveItem.PosterSets {
		if posterSet.ToDelete {
			Err := api.DB_Delete_PosterSet(ctx, posterSet.PosterSetID, saveItem.MediaItem.TMDB_ID, saveItem.MediaItem.LibraryTitle)
			if Err.Message != "" {
				api.Util_Response_SendJSON(w, ld, nil)
				return
			}
		} else {
			Err := api.DB_InsertAllInfoIntoTables(ctx, saveItem)
			if Err.Message != "" {
				api.Util_Response_SendJSON(w, ld, nil)
				return
			}
		}
	}

	api.Util_Response_SendJSON(w, ld, "success")
}
