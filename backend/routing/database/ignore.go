package routes_db

import (
	"aura/database"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

func IgnoreItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Ignore Item In Database", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get query parameters
	tmdbID := r.URL.Query().Get("tmdb_id")
	libraryTitle := r.URL.Query().Get("library_title")
	mode := r.URL.Query().Get("mode") // e.g., "always", "temp"

	if tmdbID == "" || libraryTitle == "" || mode == "" {
		logAction.SetError("Missing required query parameters", "TMDB ID, Library Title, and Mode are required",
			map[string]any{
				"tmdb_id":       tmdbID,
				"library_title": libraryTitle,
				"mode":          mode,
			})
		httpx.SendResponse(w, ld, nil)
		return
	} else if mode != "always" && mode != "temp" {
		logAction.SetError("Invalid mode parameter", "Ignore mode must be 'always' or 'temp'", map[string]any{
			"mode": mode,
		})
		httpx.SendResponse(w, ld, nil)
		return
	}

	Err := database.IgnoreMediaItem(ctx, tmdbID, libraryTitle, mode)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	var response struct {
		Ignored      bool   `json:"ignored"`
		TmdbID       string `json:"tmdb_id"`
		LibraryTitle string `json:"library_title"`
		Mode         string `json:"mode"`
	}
	response.Ignored = true
	response.TmdbID = tmdbID
	response.LibraryTitle = libraryTitle
	response.Mode = mode

	httpx.SendResponse(w, ld, response)
}

func StopIgnoringItem(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Stop Ignoring Item In Database", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get query parameters
	tmdbID := r.URL.Query().Get("tmdb_id")
	libraryTitle := r.URL.Query().Get("library_title")

	if tmdbID == "" || libraryTitle == "" {
		logAction.SetError("Missing required query parameters", "TMDB ID and Library Title are required",
			map[string]any{
				"tmdb_id":       tmdbID,
				"library_title": libraryTitle,
			})
		httpx.SendResponse(w, ld, nil)
		return
	}

	Err := database.StopIgnoringMediaItem(ctx, tmdbID, libraryTitle)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	var response struct {
		Ignored      bool   `json:"ignored"`
		TmdbID       string `json:"tmdb_id"`
		LibraryTitle string `json:"library_title"`
	}
	response.Ignored = false
	response.TmdbID = tmdbID
	response.LibraryTitle = libraryTitle

	httpx.SendResponse(w, ld, response)
}
