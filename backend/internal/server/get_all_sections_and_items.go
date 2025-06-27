package mediaserver

import (
	"aura/internal/cache"
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	mediaserver_shared "aura/internal/server/shared"
	"strconv"
)

func GetAllSectionsAndItems() {

	// Get all sections from the media server
	allSections, Err := CallFetchLibrarySectionInfo()
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return
	}

	var mediaServer mediaserver_shared.MediaServer
	switch config.Global.MediaServer.Type {
	case "Plex":
		mediaServer = &mediaserver_shared.PlexServer{}
	case "Emby", "Jellyfin":
		mediaServer = &mediaserver_shared.EmbyJellyServer{}
	default:
		return
	}

	// Iterate through each section and fetch all its items starting at 0
	for _, section := range allSections {
		sectionStartIndex := "0"
		var allItems []modals.MediaItem

		for {
			items, totalSize, Err := mediaServer.FetchLibrarySectionItems(section, sectionStartIndex)
			if Err.Message != "" {
				logging.LOG.Error(Err.Message)
				break
			}

			if len(items) == 0 {
				break
			}

			allItems = append(allItems, items...)

			// Increment the start index for the next batch of items by 500
			sectionStartIndexInt, err := strconv.Atoi(sectionStartIndex)
			if err != nil {
				logging.LOG.Error("Failed to convert sectionStartIndex to int: " + err.Error())
				break
			}
			sectionStartIndex = strconv.Itoa(sectionStartIndexInt + 500)

			// If the total size is less than the next start index, we have fetched all items
			if totalSize < sectionStartIndexInt+500 {
				break
			}

		}

		section.MediaItems = allItems
		cache.LibraryCacheStore.Update(&section)
	}

}
