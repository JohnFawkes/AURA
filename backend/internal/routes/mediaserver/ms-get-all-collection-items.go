package routes_ms

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func GetCollectionItems(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Collection Items", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get all sections from the media server
	allSections, logErr := api.CallFetchLibrarySectionInfo(ctx)
	if logErr.Message != "" {
		api.Util_Response_SendJSON(w, nil, logErr)
		return
	}

	// Loop through all libraries in the Sections
	var collectionItems []api.CollectionItem
	for _, section := range allSections {
		if section.Type != "movie" {
			continue
		}

		sectionCollectionItems, Err := api.CallFetchMovieCollectionItems(ctx, section)
		if Err.Message != "" {
			api.Util_Response_SendJSON(w, nil, Err)
			return
		}
		collectionItems = append(collectionItems, sectionCollectionItems...)
	}

	api.Util_Response_SendJSON(w, ld, collectionItems)
}
