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
		Msg("Processing search query")

	var mediaItems []api.MediaItem
	var mediux_usernames []api.MediuxUserInfo

	allSections := api.Global_Cache_LibraryStore.GetAllSectionsSortedByTitle()
searchLoop:
	for _, section := range allSections {
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
			if strings.Contains(strings.ToLower(item.Title), strings.ToLower(cleanedQuery)) {
				mediaItems = append(mediaItems, item)
			}
			if len(mediaItems) >= 10 {
				break searchLoop
			}
		}
	}

	// Get MediUX Usernames and filter based on query
	var Err logging.LogErrorInfo
	if cleanedQuery != "" {
		mediux_usernames, Err = api.Mediux_GetAllUsers(ctx, cleanedQuery)
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

	logging.LOGGER.Debug().Timestamp().
		Int("found_media_items", len(mediaItems)).
		Int("found_mediux_usernames", len(mediux_usernames)).
		Msg("Search completed")

	api.Util_Response_SendJSON(w, ld, map[string]any{
		"search_query":     searchQuery,
		"media_items":      mediaItems,
		"mediux_usernames": mediux_usernames,
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
