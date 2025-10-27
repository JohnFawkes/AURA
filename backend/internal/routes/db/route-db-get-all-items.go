package routes_db

import (
	"aura/internal/api"
	"aura/internal/logging"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func GetAllItems(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()

	params := r.URL.Query()
	logging.LOG.Trace(fmt.Sprintf("Query Params: %v", params))

	// Query Param - TMDB ID (Search Term)
	searchTMDBID := params.Get("searchTMDBID")

	// Query Param - Library (Search Term)
	searchLibrary := params.Get("searchLibrary")
	if searchLibrary != "" {
		searchLibrary = strings.TrimSpace(searchLibrary)
	}

	// Query Param - Year (Search Term)
	searchYear := 0
	if yearStr := params.Get("searchYear"); yearStr != "" {
		fmt.Sscanf(yearStr, "%d", &searchYear)
	}

	// Query Param - Title (Search Term)
	searchTitle := params.Get("searchTitle")

	// Query Param - Library Sections (Filter)
	librarySections := params["librarySections"]
	if len(librarySections) == 0 {
		librarySections = params["librarySections[]"]
	}
	if len(librarySections) == 0 {
		if lsStr := params.Get("librarySections"); lsStr != "" {
			librarySections = strings.Split(lsStr, ",")
		}
	}

	// Query Param - Filtered Types (Filter)
	filteredTypes := params["filteredTypes"]
	if len(filteredTypes) == 0 {
		filteredTypes = params["filteredTypes[]"]
	}
	if len(filteredTypes) == 0 {
		if ftStr := params.Get("filteredTypes"); ftStr != "" {
			filteredTypes = strings.Split(ftStr, ",")
		}
	}

	// Query Param - AutoDownload Only (Filter)
	filterAutodownload := params.Get("filterAutodownload")
	if filterAutodownload == "" {
		filterAutodownload = "all"
	}
	if filterAutodownload != "all" && filterAutodownload != "on" && filterAutodownload != "off" && filterAutodownload != "none" {
		logging.LOG.Warn(fmt.Sprintf("Invalid filterAutodownload value: %s, defaulting to 'all'", filterAutodownload))
		filterAutodownload = "all"
	}

	// Query Param - Multi-Set Only (Filter)
	multisetOnly := false
	if msStr := params.Get("multisetOnly"); msStr == "true" {
		multisetOnly = true
	}

	// Query Param - Filtered Usernames (Filter)
	filteredUsernames := params["filteredUsernames"]
	if len(filteredUsernames) == 0 {
		filteredUsernames = params["filteredUsernames[]"]
	}
	if len(filteredUsernames) == 0 {
		if fuStr := params.Get("filteredUsernames"); fuStr != "" {
			filteredUsernames = strings.Split(fuStr, ",")
		}
	}

	// Query Param - Pagination & Sorting
	// Items Per Page (Default: 20, -1 for no pagination)
	// Page Number (Default: 1)
	// Sort Option (Default: dateDownloaded)
	// Sort Order (Default: desc)
	itemsPerPage := 20
	if ippStr := params.Get("itemsPerPage"); ippStr != "" {
		fmt.Sscanf(ippStr, "%d", &itemsPerPage)
	}
	pageNumber := 1
	if pnStr := params.Get("pageNumber"); pnStr != "" {
		fmt.Sscanf(pnStr, "%d", &pageNumber)
	}
	sortOption := "dateDownloaded"
	if soStr := params.Get("sortOption"); soStr != "" {
		sortOption = soStr
	}
	sortOrder := "desc"
	if soStr := params.Get("sortOrder"); soStr != "" {
		if soStr == "asc" || soStr == "desc" {
			sortOrder = soStr
		} else {
			logging.LOG.Warn(fmt.Sprintf("Invalid sortOrder value: %s, defaulting to 'desc'", soStr))
		}
	}

	items, totalItems, uniqueUsers, Err := api.DB_GetAllItemsWithFilter(
		searchTMDBID,
		searchLibrary,
		searchYear,
		searchTitle,
		librarySections,
		filteredTypes,
		filterAutodownload,
		multisetOnly,
		filteredUsernames,
		itemsPerPage,
		pageNumber,
		sortOption,
		sortOrder,
	)
	if Err.Message != "" {
		api.Util_Response_SendJsonError(w, api.Util_ElapsedTime(startTime), Err)
		return
	}

	logging.LOG.Debug(fmt.Sprintf("Retrieved %d items (Total Unique Items: %d, Unique Users: %d)", len(items), totalItems, len(uniqueUsers)))

	api.Util_Response_SendJson(w, http.StatusOK, api.JSONResponse{
		Status:  "success",
		Elapsed: api.Util_ElapsedTime(startTime),
		Data: map[string]any{
			"items":        items,
			"total_items":  totalItems,
			"unique_users": uniqueUsers,
		},
	})
}
