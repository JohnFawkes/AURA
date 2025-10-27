package api

import (
	"aura/internal/logging"
	"strconv"
	"sync"
)

// Called in main on startup so that the cache has all sections and items
func GetAllSectionsAndItems() {

	// Get all sections from the media server
	allSections, Err := CallFetchLibrarySectionInfo()
	if Err.Message != "" {
		logging.LOG.Error(Err.Message)
		return
	}

	mediaServer, Err := GetMediaServerInterface(Config_MediaServer{})
	if Err.Message != "" {
		return
	}

	var wg sync.WaitGroup

	for _, section := range allSections {
		wg.Add(1)
		go func(section LibrarySection) {
			defer wg.Done()
			sectionStartIndex := "0"
			var allItems []MediaItem

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

				section.MediaItems = allItems
				Global_Cache_LibraryStore.UpdateSection(&section)

				sec, _ := Global_Cache_LibraryStore.GetSectionByTitle(section.Title)
				logging.LOG.Trace("Current cached items for section " + section.Title + ": " + strconv.Itoa(len(sec.MediaItems)))

				sectionStartIndexInt, err := strconv.Atoi(sectionStartIndex)
				if err != nil {
					logging.LOG.Error("Failed to convert sectionStartIndex to int: " + err.Error())
					break
				}
				sectionStartIndex = strconv.Itoa(sectionStartIndexInt + 500)

				if totalSize < sectionStartIndexInt+500 {
					break
				}
			}
		}(section)
	}

	wg.Wait()
}
