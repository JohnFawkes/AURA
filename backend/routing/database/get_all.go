package routes_db

import (
	"aura/database"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
	"strconv"
	"strings"
)

func GetAllItems(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get All Database Items", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	dbFilter := models.DBFilter{}

	getQueryParamsAction := logAction.AddSubAction("Get Query Parameters", logging.LevelDebug)
	// Query Param - TMDB ID (Search Term)
	dbFilter.ItemTMDB_ID = r.URL.Query().Get("item_tmdb_id")

	// Query Param - Library (Search Term)
	dbFilter.ItemLibraryTitle = r.URL.Query().Get("item_library_title")

	// Query Param - Year (Search Term)
	itemYearStr := r.URL.Query().Get("item_year")
	itemYear := 0
	if itemYearStr != "" {
		if val, err := strconv.Atoi(itemYearStr); err == nil {
			itemYear = val
		}
	}
	dbFilter.ItemYear = itemYear

	// Query Param - Title (Search Term)
	dbFilter.ItemTitle = r.URL.Query().Get("item_title")

	// Query Param - Library Sections (Filter)
	librarySectionsStr := r.URL.Query().Get("library_titles")
	var filteredLibraryTitles []string
	if librarySectionsStr != "" {
		filteredLibraryTitles = strings.Split(librarySectionsStr, ",")
	}
	dbFilter.LibraryTitles = filteredLibraryTitles

	// Query Param - Filtered Types (Filter)
	imageTypesStr := r.URL.Query().Get("image_types")
	var filteredTypes []string
	if imageTypesStr != "" {
		filteredTypes = strings.Split(imageTypesStr, ",")
	}
	dbFilter.ImageTypes = filteredTypes

	// Query Param - AutoDownload Only (Filter)
	filterAutodownload := r.URL.Query().Get("autodownload")
	if filterAutodownload == "" {
		filterAutodownload = "all"
	}
	if filterAutodownload != "all" && filterAutodownload != "on" && filterAutodownload != "off" && filterAutodownload != "none" {
		getQueryParamsAction.AppendWarning("message", "Invalid filterAutodownload value, defaulting to 'all'")
		filterAutodownload = "all"
	}
	dbFilter.Autodownload = filterAutodownload

	// Query Param - Multi-Set Only (Filter)
	multisetOnly := false
	multisetOnlyStr := r.URL.Query().Get("multiset_only")
	if strings.ToLower(multisetOnlyStr) == "true" {
		multisetOnly = true
	}
	dbFilter.MultiSetOnly = multisetOnly

	// Query Param - Filtered Usernames (Filter)
	filteredUsernamesStr := r.URL.Query().Get("usernames")
	var filteredUsernames []string
	if filteredUsernamesStr != "" {
		filteredUsernames = strings.Split(filteredUsernamesStr, ",")
	} else if fuStr := r.URL.Query().Get("filtered_usernames"); fuStr != "" {
		filteredUsernames = strings.Split(fuStr, ",")
	}
	dbFilter.Usernames = filteredUsernames

	// Query Param - Pagination & Sorting
	// Items Per Page (Default: 20, -1 for no pagination)
	// Page Number (Default: 1)
	// Sort Option (Default: date_downloaded)
	// Sort Order (Default: desc)
	itemsPerPage := 20
	pageNumber := 1
	sortOption := "date_downloaded"
	sortOrder := "desc"
	ippStr := r.URL.Query().Get("items_per_page")
	if ippStr != "" {
		if val, err := strconv.Atoi(ippStr); err == nil {
			itemsPerPage = val
		}
	}
	pnStr := r.URL.Query().Get("page_number")
	if pnStr != "" {
		if val, err := strconv.Atoi(pnStr); err == nil {
			pageNumber = val
		}
	}
	sortOption = r.URL.Query().Get("sort_option")
	sortOrder = r.URL.Query().Get("sort_order")
	getQueryParamsAction.Complete()
	dbFilter.ItemsPerPage = itemsPerPage
	dbFilter.PageNumber = pageNumber
	dbFilter.SortOption = sortOption
	dbFilter.SortOrder = sortOrder

	// Get the items from the database
	out, Err := database.GetAllSavedSets(ctx, dbFilter)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}
	totalItems := out.Total

	// // Get Saved Sets Count
	// totalSets, ErrCount := database.GetCountSavedSets(ctx)
	// if ErrCount.Message != "" {
	// 	httpx.SendResponse(w, ld, nil)
	// 	return
	// }

	// Get a list of unique users
	uniqueUsers, Err := database.GetAllUniqueUsers(ctx)
	if Err.Message != "" {
		httpx.SendResponse(w, ld, nil)
		return
	}

	httpx.SendResponse(w, ld, map[string]any{
		"items":        out.Items,
		"total_items":  totalItems,
		"unique_users": uniqueUsers,
		//"total_sets":   totalSets,
	})
}
