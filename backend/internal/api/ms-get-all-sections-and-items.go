package api

import (
	"aura/internal/logging"
	"context"
	"strconv"
	"sync"
)

// Called in main on startup so that the cache has all sections and items
func GetAllSectionsAndItems() {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "Setting Up - Get All Sections and Items")
	defer ld.Log()

	action := ld.AddAction("Fetching All Sections", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, action)
	defer action.Complete()

	// Get all sections from the media server
	allSections, logErr := CallFetchLibrarySectionInfo(ctx)
	if logErr.Message != "" {
		return
	}

	// Create a wait group
	var wg sync.WaitGroup

	// Loop through all libraries in the Media Server config
	for _, section := range allSections {
		wg.Add(1)
		go func(section LibrarySection, ctx context.Context) {
			defer wg.Done()
			sectionStartIndex := "0"
			var allItems []MediaItem

			for {
				sectionInfo, Err := CallFetchLibrarySectionItems(ctx, section.ID, section.Title, section.Type, sectionStartIndex)
				if Err.Message != "" {
					break
				}

				if len(sectionInfo.MediaItems) == 0 {
					break
				}

				allItems = append(allItems, sectionInfo.MediaItems...)

				section.MediaItems = allItems
				Global_Cache_LibraryStore.UpdateSection(&section)

				sec, _ := Global_Cache_LibraryStore.GetSectionByTitle(section.Title)
				logging.LOGGER.Info().Timestamp().Int("items", len(sec.MediaItems)).Msgf("Fetched items for section '%s'", section.Title)

				sectionStartIndexInt, err := strconv.Atoi(sectionStartIndex)
				if err != nil {
					logging.LOGGER.Error().Timestamp().Msgf("Failed to convert sectionStartIndex to int: %s", err.Error())
					break
				}
				sectionStartIndex = strconv.Itoa(sectionStartIndexInt + 500)

				if sectionInfo.TotalSize < sectionStartIndexInt+500 {
					break
				}
			}
		}(section, ctx)
	}

	wg.Wait()
}
