package api

import (
	"aura/internal/logging"
	"fmt"
	"strconv"
)

func (s *SonarrApp) HandleTags(app Config_SonarrRadarrApp, item MediaItem) logging.StandardError {
	return SR_HandleTags(app, item)
}

func (r *RadarrApp) HandleTags(app Config_SonarrRadarrApp, item MediaItem) logging.StandardError {
	return SR_HandleTags(app, item)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_CallHandleTags(item MediaItem) logging.StandardError {
	if len(Global_Config.SonarrRadarr.Applications) == 0 {
		return logging.NewStandardError() // No applications configured, nothing to do
	}
	for _, app := range Global_Config.SonarrRadarr.Applications {
		// If the media item type is not "movie" or "show", return
		if item.Type != "movie" && item.Type != "show" {
			logging.LOG.Warn(fmt.Sprintf("Item type '%s' is not 'movie' or 'show', skipping", item.Type))
			continue
		}

		// If the media item doesn't have a TMDB ID or Library Title return
		if item.TMDB_ID == "" || item.LibraryTitle == "" {
			logging.LOG.Warn("Item is missing TMDB ID or Library Title, skipping tag handling")
			continue
		}

		// Make sure all of the required information is set
		Err := SR_MakeSureAllInfoIsSet(app)
		if Err.Message != "" {
			continue
		}

		// If the type of app doesn't match the type of item, skip
		if (app.Type == "Sonarr" && item.Type != "show") || (app.Type == "Radarr" && item.Type != "movie") {
			continue
		}

		// If the library title doesn't match, skip
		if app.Library != item.LibraryTitle {
			logging.LOG.Warn(fmt.Sprintf("%s (%s) does not match item library title '%s', skipping", app.Type, app.Library, item.LibraryTitle))
			continue
		}

		interfaceSR, Err := SR_GetSonarrRadarrInterface(app)
		if Err.Message != "" {
			logging.LOG.Warn(Err.Message)
			continue
		}

		Err = interfaceSR.HandleTags(app, item)
		if Err.Message != "" {
			logging.LOG.Warn(Err.Message)
		}
	}
	return logging.StandardError{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func SR_HandleTags(app Config_SonarrRadarrApp, item MediaItem) logging.StandardError {
	Err := logging.NewStandardError()

	// Get the full item information from Sonarr/Radarr
	tmdbIDInt, _ := strconv.Atoi(item.TMDB_ID)
	srItem, Err := SR_CallGetItemInfoFromTMDBID(app, tmdbIDInt)
	if Err.Message != "" {
		return Err
	}

	// Get all of the tags from Sonarr/Radarr
	allAvailableTags, Err := SR_CallGetAllTags(app)
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

		// Actually add missing tags to Sonarr/Radarr
		if len(tagsToAdd) > 0 {
			newTags, Err := SR_CallAddNewTags(app, tagsToAdd)
			if Err.Message != "" {
				return Err
			}
			// Add new tags to availableTagMap
			for _, tag := range newTags {
				availableTagMap[tag.Label] = tag
			}
		}

		// Build finalTags: start with current tags from srItem, add all tags from app.Add
		var currentTagIDs []int64
		switch app.Type {
		case "Sonarr":
			srSonarrItem, ok := srItem.(SR_SonarrItem)
			if !ok {
				Err.Message = "srItem is not of type SR_SonarrItem"
				return Err
			}
			currentTagIDs = srSonarrItem.Tags
		case "Radarr":
			srRadarrItem, ok := srItem.(SR_RadarrItem)
			if !ok {
				Err.Message = "srItem is not of type SR_RadarrItem"
				return Err
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
		Err = SR_CallUpdateItemInfo(app, srItem)
		if Err.Message != "" {
			return Err
		}
		logging.LOG.Info(fmt.Sprintf("Updated %s (%s) item '%s' with tags: %v", app.Type, app.Library, item.Title, finalTagLabels))
	}

	return Err
}
