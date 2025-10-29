package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"strconv"
)

func (s *SonarrApp) HandleTags(ctx context.Context, app Config_SonarrRadarrApp, item MediaItem) logging.LogErrorInfo {
	return SR_HandleTags(ctx, app, item)
}

func (r *RadarrApp) HandleTags(ctx context.Context, app Config_SonarrRadarrApp, item MediaItem) logging.LogErrorInfo {
	return SR_HandleTags(ctx, app, item)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_CallHandleTags(ctx context.Context, item MediaItem) logging.LogErrorInfo {
	if len(Global_Config.SonarrRadarr.Applications) == 0 {
		return logging.LogErrorInfo{}
	}

	ctx, ld := logging.CreateLoggingContext(ctx, "Sonarr/Radarr - Handle Tags for Media Item")
	logAction := ld.AddAction(fmt.Sprintf("Handling Tags for %s", item.Title), logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer ld.Log()
	defer logAction.Complete()

	if item.Type != "movie" && item.Type != "show" {
		logAction.SetError("Unsupported Media Item Type",
			"Only media items of type 'movie' or 'show' are supported for tag handling.",
			map[string]any{
				"title":   item.Title,
				"type":    item.Type,
				"tmdb_id": item.TMDB_ID,
				"library": item.LibraryTitle,
			})
		return *logAction.Error
	}

	// If the media item doesn't have a TMDB ID or Library Title return
	if item.TMDB_ID == "" || item.LibraryTitle == "" {
		logAction.SetError("Missing Required Media Item Information",
			"Media item must have both TMDB ID and Library Title to handle tags.",
			map[string]any{
				"title":         item.Title,
				"type":          item.Type,
				"tmdb_id":       item.TMDB_ID,
				"library_title": item.LibraryTitle,
			})
		return *logAction.Error
	}

	for _, app := range Global_Config.SonarrRadarr.Applications {
		// Make sure all of the required information is set
		Err := SR_MakeSureAllInfoIsSet(ctx, app)
		if Err.Message != "" {
			continue
		}

		// If the type of app doesn't match the type of item, skip
		if (app.Type == "Sonarr" && item.Type != "show") || (app.Type == "Radarr" && item.Type != "movie") {
			continue
		}

		actionApp := logAction.AddSubAction(fmt.Sprintf("Handling Tags for %s (%s)", app.Type, app.Library), logging.LevelInfo)

		// If the library title doesn't match, skip
		if app.Library != item.LibraryTitle {
			actionApp.AppendWarning("message", fmt.Sprintf("Library title '%s' does not match application library '%s', skipping", item.LibraryTitle, app.Library))
			continue
		}

		interfaceSR, Err := NewSonarrRadarrInterface(ctx, app)
		if Err.Message != "" {
			actionApp.SetError("Failed to Get Interface",
				fmt.Sprintf("Skipping this %s instance because the interface could not be retrieved", app.Type),
				map[string]any{
					"error": Err.Message,
				})
			continue
		}
		Err = interfaceSR.HandleTags(ctx, app, item)
		if Err.Message != "" {
			actionApp.SetError("Failed to Handle Tags",
				fmt.Sprintf("An error occurred while handling tags for this %s instance", app.Type),
				map[string]any{
					"error": Err.Message,
				})
			continue
		}

	}
	return logging.LogErrorInfo{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_HandleTags(ctx context.Context, app Config_SonarrRadarrApp, item MediaItem) logging.LogErrorInfo {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Handling tags for %s in %s (%s)", item.Title, app.Type, app.Library), logging.LevelInfo)
	defer logAction.Complete()

	// Get the full information about the Media Item from Sonarr/Radarr
	tmdbIDInt, _ := strconv.Atoi(item.TMDB_ID)
	srItem, Err := SR_CallGetItemInfoFromTMDBID(ctx, app, tmdbIDInt)
	if Err.Message != "" {
		return Err
	}

	// Get all of the tags from Sonarr/Radarr
	allAvailableTags, Err := SR_CallGetAllTags(ctx, app)
	if Err.Message != "" {
		return Err
	}

	// Get all of the applications configured for labels and tags
	for _, labelApp := range Global_Config.LabelsAndTags.Applications {
		if labelApp.Application != "Sonarr" && labelApp.Application != "Radarr" {
			continue
		}

		// Only proceed if the application is enabled
		if !labelApp.Enabled {
			continue
		}

		// Check to see there is at least one tag to add or remove
		if len(labelApp.Add) == 0 && len(labelApp.Remove) == 0 {
			continue
		}

		// Build a map for quick lookup
		availableTagMap := make(map[string]SonarrRadarrTag)
		for _, tag := range allAvailableTags {
			availableTagMap[tag.Label] = tag
		}

		// Add any missing tags from app.Add
		tagsToAdd := make([]string, 0)
		for _, tagToAdd := range labelApp.Add {
			if _, exists := availableTagMap[tagToAdd]; !exists {
				tagsToAdd = append(tagsToAdd, tagToAdd)
			}
		}

		// If there are tags to add, add them
		if len(tagsToAdd) > 0 {
			newTags, Err := SR_CallAddNewTags(ctx, app, tagsToAdd)
			if Err.Message != "" {
				continue
			}
			// Add new tags to availableTagMap
			for _, newTag := range newTags {
				availableTagMap[newTag.Label] = newTag
			}
		}

		// Build finalTags: start with current tags from srItem, add all tags from app.Add
		var currentTagIDs []int64
		switch app.Type {
		case "Sonarr":
			srSonarrItem, ok := srItem.(SR_SonarrItem)
			if !ok {
				logAction.SetError("Type Assertion Failed", "Failed to assert type to SR_SonarrItem", nil)
				return *logAction.Error
			}
			currentTagIDs = srSonarrItem.Tags
		case "Radarr":
			srRadarrItem, ok := srItem.(SR_RadarrItem)
			if !ok {
				logAction.SetError("Type Assertion Failed", "Failed to assert type to SR_RadarrItem", nil)
				return *logAction.Error
			}
			currentTagIDs = srRadarrItem.Tags
		}

		// Build a set of tag IDs to avoid duplicates
		finalTagIDSet := make(map[int64]struct{})
		for _, id := range currentTagIDs {
			finalTagIDSet[id] = struct{}{}
		}

		// Add all tags from app.Add
		for _, tagLabel := range labelApp.Add {
			if tag, exists := availableTagMap[tagLabel]; exists {
				finalTagIDSet[int64(tag.ID)] = struct{}{}
			}
		}

		// Remove tags in app.Remove
		for _, tagLabel := range labelApp.Remove {
			if tag, exists := availableTagMap[tagLabel]; exists {
				delete(finalTagIDSet, int64(tag.ID))
			}
		}

		// Convert set to slice
		finalTagIDs := make([]int64, 0, len(finalTagIDSet))
		for id := range finalTagIDSet {
			finalTagIDs = append(finalTagIDs, id)
		}

		// Get tag labels for finalTagIDs
		finalTagLabels := make([]string, 0, len(finalTagIDs))
		for _, id := range finalTagIDs {
			for label, tag := range availableTagMap {
				if int64(tag.ID) == id {
					finalTagLabels = append(finalTagLabels, label)
					break
				}
			}
		}

		// Assign final tags to srItem
		switch app.Type {
		case "Sonarr":
			srSonarrItem, _ := srItem.(SR_SonarrItem)
			srSonarrItem.Tags = finalTagIDs
			srItem = srSonarrItem
		case "Radarr":
			srRadarrItem, _ := srItem.(SR_RadarrItem)
			srRadarrItem.Tags = finalTagIDs
			srItem = srRadarrItem
		}

		// Update the item in Sonarr/Radarr with the new tags
		Err = SR_CallUpdateItemInfo(ctx, app, srItem)
		if Err.Message != "" {
			return Err
		}

		logAction.AppendResult("final_tags", finalTagLabels)
	}

	return logging.LogErrorInfo{}
}
