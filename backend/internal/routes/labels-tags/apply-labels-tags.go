package routes_labelstags

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
)

func ApplyLabelsTags(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Apply Labels/Tags to Items", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Parse the request body to get the DBMediaItemWithPosterSets item
	var saveItem api.DBMediaItemWithPosterSets
	Err := api.DecodeRequestBodyJSON(ctx, r.Body, &saveItem, "ApplyLabelsTagsRequestBody")
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate the JSON structure
	validateAction := logAction.AddSubAction("Validate Save Item", logging.LevelDebug)
	// Make sure it contains a TMDB_ID and LibraryTitle
	if saveItem.TMDB_ID == "" || saveItem.LibraryTitle == "" {
		validateAction.SetError("Missing Required Fields", "TMDB_ID or LibraryTitle is empty",
			map[string]any{
				"body": saveItem,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate that there is a MediaItem
	if saveItem.MediaItem.TMDB_ID == "" {
		validateAction.SetError("Missing Media Item Field", "MediaItem.TMDB_ID is empty",
			map[string]any{
				"body": saveItem,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate that there is at least one PosterSet
	if len(saveItem.PosterSets) == 0 {
		validateAction.SetError("Missing Poster Set", "At least one PosterSet is required",
			map[string]any{
				"body": saveItem,
			})
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	// Validate each PosterSetDetail
	for _, ps := range saveItem.PosterSets {
		if ps.PosterSetID == "" {
			validateAction.SetError("Missing PosterSetDetail Field", "PosterSetDetail.PosterSetID is empty",
				map[string]any{
					"body": saveItem,
				})
			api.Util_Response_SendJSON(w, ld, nil)
			return
		}
	}
	validateAction.Complete()

	var selectedTypes []string
	if len(saveItem.PosterSets) > 1 {
		// If there are multiple poster sets, we append all unique selected types
		typeSet := make(map[string]struct{})
		for _, ps := range saveItem.PosterSets {
			for _, t := range ps.SelectedTypes {
				typeSet[t] = struct{}{}
			}
		}
		for t := range typeSet {
			selectedTypes = append(selectedTypes, t)
		}
	} else if len(saveItem.PosterSets) == 1 {
		// If only one poster set, use its selected types
		selectedTypes = saveItem.PosterSets[0].SelectedTypes
	}

	api.Plex_HandleLabels(saveItem.MediaItem, selectedTypes)
	Err = api.SR_CallHandleTags(ctx, saveItem.MediaItem, selectedTypes)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, saveItem)
}
