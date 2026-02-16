package routes_search

import (
	"aura/cache"
	"aura/database"
	"aura/logging"
	"aura/mediux"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
	"strconv"
	"strings"
)

func Search(w http.ResponseWriter, r *http.Request) {
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
	searchCollectionItems := r.URL.Query().Get("search_collection_items") == "true"
	searchMediuxUsers := r.URL.Query().Get("search_mediux_users") == "true"
	searchSavedSets := r.URL.Query().Get("search_saved_sets") == "true"

	if !searchMediaItems && !searchCollectionItems && !searchMediuxUsers && !searchSavedSets {
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

	logging.Dev().Timestamp().
		Str("query_original", searchQuery).
		Str("query_cleaned", cleanedQuery).
		Str("library_filter", libraryFilter).
		Str("year_filter", yearFilter).
		Str("id_filter", idFilter).
		Bool("search_media_items", searchMediaItems).
		Bool("search_collection_items", searchCollectionItems).
		Bool("search_mediux_users", searchMediuxUsers).
		Bool("search_saved_sets", searchSavedSets).
		Msg("Processing search query")

	var Err logging.LogErrorInfo

	var response struct {
		SearchQuery                   string                  `json:"search_query"`
		MediaItems                    []models.MediaItem      `json:"media_items"`
		MediaItemsLastFullUpdate      int64                   `json:"media_items_last_full_update"`
		CollectionItems               []models.CollectionItem `json:"collection_items"`
		CollectionItemsLastFullUpdate int64                   `json:"collection_items_last_full_update"`
		MediuxUsernames               []models.MediuxUserInfo `json:"mediux_usernames"`
		MediuxUsernamesLastFullUpdate int64                   `json:"mediux_usernames_last_full_update"`
		SavedSets                     []models.DBSavedItem    `json:"saved_sets"`
	}
	response.MediaItemsLastFullUpdate = cache.LibraryStore.LastFullUpdate
	response.MediuxUsernamesLastFullUpdate = cache.MediuxUsers.LastFullUpdate
	response.CollectionItemsLastFullUpdate = cache.CollectionsStore.LastFullUpdate

	if searchMediaItems {
		maxPerSection := 5
		maxTotal := 10
		allSections := cache.LibraryStore.GetAllSectionsSortedByTitle()
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
					response.MediaItems = append(response.MediaItems, item)
					count++
					if count >= maxPerSection || len(response.MediaItems) >= maxTotal {
						break
					}
				}
			}
			if len(response.MediaItems) >= maxTotal {
				break
			}
		}
	}

	if searchCollectionItems {
		maxPerSection := 5
		maxTotal := 10
		allCollections := cache.CollectionsStore.GetAllCollections()
		for _, collection := range allCollections {
			count := 0
			if hasLibraryFilter && !strings.EqualFold(collection.LibraryTitle, libraryFilter) {
				continue
			}
			queryWords := strings.Fields(strings.ToLower(cleanedQuery))
			titleLower := strings.ToLower(collection.Title)
			allWordsMatch := true
			for _, word := range queryWords {
				if !strings.Contains(titleLower, word) {
					allWordsMatch = false
					break
				}
			}
			if allWordsMatch {
				response.CollectionItems = append(response.CollectionItems, collection)
				count++
				if count >= maxPerSection || len(response.CollectionItems) >= maxTotal {
					break
				}
			}
			if len(response.CollectionItems) >= maxTotal {
				break
			}
		}
	}

	if searchMediuxUsers {
		// Get MediUX Usernames and filter based on query
		if cleanedQuery != "" {
			response.MediuxUsernames, Err = mediux.SearchUsersByUsername(ctx, cleanedQuery)
			if Err.Message != "" {
				ld.Status = logging.StatusWarn
				httpx.SendResponse(w, ld, map[string]any{
					"search_query":     searchQuery,
					"media_items":      response.MediaItems,
					"mediux_usernames": response.MediuxUsernames,
					"error":            Err})
				return
			}
		}
	}

	if searchSavedSets {
		// Get Saved Sets and filter based on query
		out, Err := database.GetAllSavedSets(ctx, models.DBFilter{
			ItemTMDB_ID:      idFilter,
			ItemTitle:        cleanedQuery,
			ItemLibraryTitle: libraryFilter,
			ItemYear:         yearFilterInt,
		})
		if Err.Message != "" {
			ld.Status = logging.StatusWarn
			httpx.SendResponse(w, ld, response)
			return
		}
		response.SavedSets = out.Items
	}

	logging.LOGGER.Trace().Timestamp().
		Int("found_media_items", len(response.MediaItems)).
		Int("found_mediux_usernames", len(response.MediuxUsernames)).
		Int("found_saved_sets", len(response.SavedSets)).
		Msg("Search completed")

	httpx.SendResponse(w, ld, response)
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
