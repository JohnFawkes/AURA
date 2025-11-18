package routes_db

import (
	"aura/internal/api"
	"aura/internal/logging"
	"net/http"
	"strconv"
	"strings"
)

func GetAllItems(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get All Items From Database", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	getQueryParamsAction := logAction.AddSubAction("Get Query Parameters", logging.LevelDebug)
	// Query Param - TMDB ID (Search Term)
	searchTMDBID := r.URL.Query().Get("searchTMDBID")

	// Query Param - Library (Search Term)
	searchLibrary := r.URL.Query().Get("searchLibrary")

	// Query Param - Year (Search Term)
	searchYearStr := r.URL.Query().Get("searchYear")
	searchYear := 0
	if searchYearStr != "" {
		if val, err := strconv.Atoi(searchYearStr); err == nil {
			searchYear = val
		}
	}

	// Query Param - Title (Search Term)
	searchTitle := r.URL.Query().Get("searchTitle")

	// Query Param - Library Sections (Filter)
	librarySectionsStr := r.URL.Query().Get("librarySections")
	var librarySections []string
	if librarySectionsStr != "" {
		librarySections = strings.Split(librarySectionsStr, ",")
	}

	// Query Param - Filtered Types (Filter)
	filteredTypesStr := r.URL.Query().Get("filteredTypes")
	var filteredTypes []string
	if filteredTypesStr != "" {
		filteredTypes = strings.Split(filteredTypesStr, ",")
	}

	// Query Param - AutoDownload Only (Filter)
	filterAutodownload := r.URL.Query().Get("filterAutodownload")
	if filterAutodownload == "" {
		filterAutodownload = "all"
	}
	if filterAutodownload != "all" && filterAutodownload != "on" && filterAutodownload != "off" && filterAutodownload != "none" {
		getQueryParamsAction.AppendWarning("message", "Invalid filterAutodownload value, defaulting to 'all'")
		filterAutodownload = "all"
	}

	// Query Param - Multi-Set Only (Filter)
	multisetOnly := false
	multisetOnlyStr := r.URL.Query().Get("multisetOnly")
	if strings.ToLower(multisetOnlyStr) == "true" {
		multisetOnly = true
	}

	// Query Param - Filtered Usernames (Filter)
	filteredUsernamesStr := r.URL.Query().Get("filteredUsernames")
	var filteredUsernames []string
	if filteredUsernamesStr != "" {
		filteredUsernames = strings.Split(filteredUsernamesStr, ",")
	} else if fuStr := r.URL.Query().Get("filteredUsernames"); fuStr != "" {
		filteredUsernames = strings.Split(fuStr, ",")
	}

	// Query Param - Pagination & Sorting
	// Items Per Page (Default: 20, -1 for no pagination)
	// Page Number (Default: 1)
	// Sort Option (Default: dateDownloaded)
	// Sort Order (Default: desc)
	itemsPerPage := 20
	pageNumber := 1
	sortOption := "dateDownloaded"
	sortOrder := "desc"
	ippStr := r.URL.Query().Get("itemsPerPage")
	if ippStr != "" {
		if val, err := strconv.Atoi(ippStr); err == nil {
			itemsPerPage = val
		}
	}
	pnStr := r.URL.Query().Get("pageNumber")
	if pnStr != "" {
		if val, err := strconv.Atoi(pnStr); err == nil {
			pageNumber = val
		}
	}
	sortOption = r.URL.Query().Get("sortOption")
	sortOrder = r.URL.Query().Get("sortOrder")
	getQueryParamsAction.Complete()

	// Get the items from the database
	items, totalItems, uniqueUsers, Err := api.DB_GetAllItemsWithFilter(
		ctx,
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
		"", // posterSetID
	)
	if Err.Message != "" {
		api.Util_Response_SendJSON(w, ld, nil)
		return
	}

	api.Util_Response_SendJSON(w, ld, map[string]any{
		"items":        items,
		"total_items":  totalItems,
		"unique_users": uniqueUsers,
	})
}
