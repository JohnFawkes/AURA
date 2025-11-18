package routes_search

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"strconv"
	"strings"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Search Handler", logging.LevelDebug)
	ctx = logging.WithCurrentAction(ctx, logAction)

	// Get the Search Query Parameter
	searchQuery := r.URL.Query().Get("query")
	if searchQuery == "" {
		return
	}

	// Get Filters from Query
	searchMediaItems := r.URL.Query().Get("search_media_items") == "true"
	searchMediuxUsers := r.URL.Query().Get("search_mediux_users") == "true"
	searchSavedSets := r.URL.Query().Get("search_saved_sets") == "true"

	if !searchMediaItems && !searchMediuxUsers && !searchSavedSets {
		return
	}

	yearFilter, hasYearFilter := extractYearFromQuery(searchQuery)
	libraryFilter, hasLibraryFilter := extractLibraryFromQuery(searchQuery)
	idFilter, hasIDFilter := extractIDFromQuery(searchQuery)
	cleanedQuery := removeOtherFiltersFromQuery(searchQuery)

	var yearFilterInt int
	if hasYearFilter {
		var err error
		yearFilterInt, err = strconv.Atoi(yearFilter)
		if err != nil {
			logging.LOGGER.Warn().Timestamp().Msgf("Invalid year filter: %s", yearFilter)
			hasYearFilter = false
		}
	}

	logging.LOGGER.Debug().Timestamp().
		Str("query_original", searchQuery).
		Str("query_cleaned", cleanedQuery).
		Str("library_filter", libraryFilter).
		Str("year_filter", yearFilter).
		Str("id_filter", idFilter).
		Bool("searchMediaItems", searchMediaItems).
		Bool("searchMediuxUsers", searchMediuxUsers).
		Bool("searchSavedSets", searchSavedSets).
		Msg("Processing search query")

	var mediux_usernames []api.MediuxUserInfo
	var mediaItems []api.MediaItem
	var savedSets []api.DBMediaItemWithPosterSets
	var Err logging.LogErrorInfo

	if searchMediaItems {
		maxPerSection := 5
		maxTotal := 10
		allSections := api.Global_Cache_LibraryStore.GetAllSectionsSortedByTitle()
		for _, section := range allSections {
			count := 0
			for _, item := range section.MediaItems {
				if hasLibraryFilter && !strings.EqualFold(item.LibraryTitle, libraryFilter) {
					continue
				}
				if hasYearFilter && item.Year != yearFilterInt {
					continue
				}
				if hasIDFilter && (item.TMDB_ID != idFilter && item.RatingKey != idFilter) {
					continue
				}
				queryWords := strings.Fields(strings.ToLower(cleanedQuery))
				titleLower := strings.ToLower(item.Title)
				allWordsMatch := true
				for _, word := range queryWords {
					if !strings.Contains(titleLower, word) {
						allWordsMatch = false
						break
					}
				}
				if allWordsMatch {
					mediaItems = append(mediaItems, item)
					count++
					if count >= maxPerSection || len(mediaItems) >= maxTotal {
						break
					}
				}
			}
			if len(mediaItems) >= maxTotal {
				break
			}
		}
	}

	if searchMediuxUsers {
		// Get MediUX Usernames and filter based on query
		if cleanedQuery != "" {
			mediux_usernames, Err = api.Mediux_SearchUsers(ctx, cleanedQuery)
			if Err.Message != "" {
				ld.Status = logging.StatusWarn
				api.Util_Response_SendJSON(w, ld, map[string]any{
					"search_query":     searchQuery,
					"media_items":      mediaItems,
					"mediux_usernames": mediux_usernames,
					"error":            Err})
				return
			}
		}
	}

	if searchSavedSets {
		dbItems, _, _, Err := api.DB_GetAllItemsWithFilter(
			ctx,
			idFilter,
			libraryFilter,
			yearFilterInt,
			cleanedQuery,
			[]string{},       // librarySections
			[]string{},       // filteredTypes
			"all",            // filterAutoDownload
			false,            // multisetOnly
			[]string{},       // filteredUsernames
			50,               // itemsPerPage (just to get count)
			1,                // pageNumber
			"dateDownloaded", // sortOption
			"desc",           // sortOrder
			"",               // posterSetID
		)
		if Err.Message != "" {
			ld.Status = logging.StatusWarn
			api.Util_Response_SendJSON(w, ld, map[string]any{
				"search_query":     searchQuery,
				"media_items":      mediaItems,
				"mediux_usernames": mediux_usernames,
				"error":            Err})
			return
		}

		// Only return 3 saved sets
		maxSavedSets := 3
		savedSets = dbItems
		if len(dbItems) > maxSavedSets {
			savedSets = dbItems[:maxSavedSets]
		}
	}

	logging.LOGGER.Debug().Timestamp().
		Int("found_media_items", len(mediaItems)).
		Int("found_mediux_usernames", len(mediux_usernames)).
		Int("found_saved_sets", len(savedSets)).
		Msg("Search completed")

	api.Util_Response_SendJSON(w, ld, map[string]any{
		"search_query":     searchQuery,
		"media_items":      mediaItems,
		"mediux_usernames": mediux_usernames,
		"saved_sets":       savedSets,
		"error":            nil,
	})

}

func extractYearFromQuery(query string) (string, bool) {
	prefix := "y:"
	suffix := ":"
	startIdx := strings.Index(strings.ToLower(query), prefix)
	if startIdx == -1 {
		return "", false
	}
	startIdx += len(prefix)
	endIdx := strings.Index(query[startIdx:], suffix)
	if endIdx == -1 {
		return "", false
	}
	year := query[startIdx : startIdx+endIdx]
	return year, true
}

func extractLibraryFromQuery(query string) (string, bool) {
	prefix := "l:"
	suffix := ":"
	startIdx := strings.Index(strings.ToLower(query), prefix)
	if startIdx == -1 {
		return "", false
	}
	startIdx += len(prefix)
	endIdx := strings.Index(query[startIdx:], suffix)
	if endIdx == -1 {
		return "", false
	}
	library := query[startIdx : startIdx+endIdx]
	return library, true
}

func extractIDFromQuery(query string) (string, bool) {
	prefix := "id:"
	suffix := ":"
	startIdx := strings.Index(strings.ToLower(query), prefix)
	if startIdx == -1 {
		return "", false
	}
	startIdx += len(prefix)
	endIdx := strings.Index(query[startIdx:], suffix)
	if endIdx == -1 {
		return "", false
	}
	id := query[startIdx : startIdx+endIdx]
	return id, true
}

func removeOtherFiltersFromQuery(query string) string {
	filters := []string{"y:", "l:", "id:"}
	cleanedQuery := query

	for _, filter := range filters {
		for {
			startIdx := strings.Index(strings.ToLower(cleanedQuery), filter)
			if startIdx == -1 {
				break
			}
			endIdx := strings.Index(cleanedQuery[startIdx+len(filter):], ":")
			if endIdx == -1 {
				break
			}
			cleanedQuery = strings.TrimSpace(cleanedQuery[:startIdx] + cleanedQuery[startIdx+len(filter)+endIdx+1:])
		}
	}

	return strings.TrimSpace(cleanedQuery)
}
