package routes_db

import (
	"aura/database"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Delete Item From Database", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the query parameters
	tmdbID := r.URL.Query().Get("tmdb_id")
	libraryTitle := r.URL.Query().Get("library_title")

	// Validate the parameters
	if tmdbID == "" || libraryTitle == "" {
		ld.AddAction("Invalid parameters for deleting item from database", logging.LevelError)
		httpx.SendResponse(w, ld, nil)
		return
	}

	// Delete the item
	Err := database.DeleteAllPosterSetsForMediaItem(ctx, tmdbID, libraryTitle)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, "ok")
}
