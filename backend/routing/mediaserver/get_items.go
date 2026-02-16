package routes_ms

import (
	"aura/cache"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func GetAllItems(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get Media Server Items", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	var mediaItemsFilters struct {
		LibraryTitles []string `json:"library_titles"`
		InDatabase    string   `json:"in_database"`
		IgnoredMode   string   `json:"ignored_mode"`
		ItemPerPage   int      `json:"items_per_page"`
		PageNumber    int      `json:"page_number"`
		SortOption    string   `json:"sort_option"`
		SortOrder     string   `json:"sort_order"`
	}

	getQueryParamsAction := logAction.AddSubAction("Get Query Parameters", logging.LevelDebug)
	// Query Param - Library Titles (Filter)
	libraryTitlesStr := r.URL.Query().Get("library_titles")
	if libraryTitlesStr != "" {
		mediaItemsFilters.LibraryTitles = strings.Split(libraryTitlesStr, ",")
	}

	// Query Param - In Database (Filter)
	mediaItemsFilters.InDatabase = r.URL.Query().Get("in_database")
	if mediaItemsFilters.InDatabase == "" {
		mediaItemsFilters.InDatabase = "all"
	}

	// Query Param - Ignored Mode (Filter)
	mediaItemsFilters.IgnoredMode = r.URL.Query().Get("ignored_mode")
	if mediaItemsFilters.IgnoredMode == "" {
		mediaItemsFilters.IgnoredMode = "none"
	}

	// Query Param - Pagination & Sorting
	// Items Per Page (Default: 20)
	// Page Number (Default: 1)
	// Sort Option (Default: dateUpdated)
	// Sort Order (Default: desc)
	itemsPerPageStr := r.URL.Query().Get("items_per_page")
	mediaItemsFilters.ItemPerPage = 20
	if itemsPerPageStr != "" {
		if val, err := strconv.Atoi(itemsPerPageStr); err == nil {
			mediaItemsFilters.ItemPerPage = val
		}
	}

	pageNumberStr := r.URL.Query().Get("page_number")
	mediaItemsFilters.PageNumber = 1
	if pageNumberStr != "" {
		if val, err := strconv.Atoi(pageNumberStr); err == nil {
			mediaItemsFilters.PageNumber = val
		}
	}

	mediaItemsFilters.SortOption = normalizeSortOption(r.URL.Query().Get("sort_option"))
	mediaItemsFilters.SortOrder = normalizeSortOrder(r.URL.Query().Get("sort_order"))
	getQueryParamsAction.Complete()

	logging.LOGGER.Debug().Timestamp().
		Strs("library_titles", mediaItemsFilters.LibraryTitles).
		Str("in_database", mediaItemsFilters.InDatabase).
		Str("ignored_mode", mediaItemsFilters.IgnoredMode).
		Int("items_per_page", mediaItemsFilters.ItemPerPage).
		Int("page_number", mediaItemsFilters.PageNumber).
		Str("sort_option", mediaItemsFilters.SortOption).
		Str("sort_order", mediaItemsFilters.SortOrder).
		Msg("Media Server Items Query Parameters")

	// Get all the media items from the cache
	mediaItems := cache.LibraryStore.GetAllMediaItems()

	var filteredItems []models.MediaItem

	// Apply filters
	for _, item := range mediaItems {
		addItem := true

		// Filter by Library Titles
		if len(mediaItemsFilters.LibraryTitles) > 0 {
			found := false
			for _, libTitle := range mediaItemsFilters.LibraryTitles {
				if strings.EqualFold(item.LibraryTitle, libTitle) {
					found = true
					break
				}
			}
			if !found {
				addItem = false
				logging.Dev().Timestamp().
					Str("item_title", item.Title).
					Str("library_title", item.LibraryTitle).
					Strs("filter_library_titles", mediaItemsFilters.LibraryTitles).
					Msg("Excluding item due to library title filter")
			}
		}

		// Filter by In Database
		if mediaItemsFilters.InDatabase != "all" {
			if mediaItemsFilters.InDatabase == "inDB" && len(item.DBSavedSets) == 0 {
				addItem = false
				logging.Dev().Timestamp().
					Str("item_title", item.Title).
					Str("library_title", item.LibraryTitle).
					Msg("Excluding item due to in_database filter (inDB)")
			} else if mediaItemsFilters.InDatabase == "notInDB" && len(item.DBSavedSets) > 0 {
				addItem = false
				logging.Dev().Timestamp().
					Str("item_title", item.Title).
					Str("library_title", item.LibraryTitle).
					Msg("Excluding item due to in_database filter (notInDB)")
			}
		}

		// Filter by Ignored Mode
		if mediaItemsFilters.IgnoredMode != "none" {
			if !item.IgnoredInDB {
				addItem = false
				logging.Dev().Timestamp().
					Str("item_title", item.Title).
					Str("library_title", item.LibraryTitle).
					Msg("Excluding item due to ignored_mode filter (not ignored)")
			} else if mediaItemsFilters.IgnoredMode == "always" && !strings.EqualFold(item.IgnoredMode, "always") {
				addItem = false
				logging.Dev().Timestamp().
					Str("item_title", item.Title).
					Str("library_title", item.LibraryTitle).
					Msg("Excluding item due to ignored_mode filter (always)")
			} else if mediaItemsFilters.IgnoredMode == "temp" && !strings.EqualFold(item.IgnoredMode, "temp") {
				addItem = false
				logging.Dev().Timestamp().
					Str("item_title", item.Title).
					Str("library_title", item.LibraryTitle).
					Msg("Excluding item due to ignored_mode filter (temp)")
			}
		}

		if addItem {
			filteredItems = append(filteredItems, item)
		}
	}

	// --- Sorting ---
	sort.SliceStable(filteredItems, func(i, j int) bool {
		a := filteredItems[i]
		b := filteredItems[j]

		less := false

		switch mediaItemsFilters.SortOption {
		case "dateAdded":
			if a.AddedAt != b.AddedAt {
				less = a.AddedAt < b.AddedAt
			} else {
				less = strings.ToLower(a.Title) < strings.ToLower(b.Title)
			}
		case "dateRelease":
			if a.ReleasedAt != b.ReleasedAt {
				less = a.ReleasedAt < b.ReleasedAt
			} else {
				less = strings.ToLower(a.Title) < strings.ToLower(b.Title)
			}
		case "title":
			at := strings.ToLower(a.Title)
			bt := strings.ToLower(b.Title)
			if at != bt {
				less = at < bt
			} else {
				less = a.RatingKey < b.RatingKey
			}
		case "dateUpdated":
			fallthrough
		default:
			if a.UpdatedAt != b.UpdatedAt {
				less = a.UpdatedAt < b.UpdatedAt
			} else {
				less = strings.ToLower(a.Title) < strings.ToLower(b.Title)
			}
		}

		if mediaItemsFilters.SortOrder == "asc" {
			return less
		}
		return !less
	})

	// --- Pagination ---
	totalItems := len(filteredItems)
	totalPages := 0
	if mediaItemsFilters.ItemPerPage > 0 {
		totalPages = (totalItems + mediaItemsFilters.ItemPerPage - 1) / mediaItemsFilters.ItemPerPage
	}

	start := (mediaItemsFilters.PageNumber - 1) * mediaItemsFilters.ItemPerPage
	if start < 0 {
		start = 0
	}
	if start > totalItems {
		start = totalItems
	}
	end := start + mediaItemsFilters.ItemPerPage
	if end > totalItems {
		end = totalItems
	}

	paged := []models.MediaItem{}
	if start < end {
		paged = filteredItems[start:end]
	}

	// --- Response (items + metadata) ---
	resp := map[string]any{
		"items":       paged,
		"total_items": totalItems,
		"total_pages": totalPages,
	}

	httpx.SendResponse(w, ld, resp)
}

func normalizeSortOption(v string) string {
	v = strings.TrimSpace(v)
	switch {
	case strings.EqualFold(v, "dateUpdated"), strings.EqualFold(v, "date_updated"), strings.EqualFold(v, "updated"):
		return "dateUpdated"
	case strings.EqualFold(v, "dateAdded"), strings.EqualFold(v, "date_added"), strings.EqualFold(v, "added"):
		return "dateAdded"
	case strings.EqualFold(v, "dateRelease"), strings.EqualFold(v, "date_release"), strings.EqualFold(v, "released"), strings.EqualFold(v, "releaseDate"):
		return "dateRelease"
	case strings.EqualFold(v, "title"), strings.EqualFold(v, "name"):
		return "title"
	default:
		return "dateUpdated"
	}
}

func normalizeSortOrder(v string) string {
	v = strings.TrimSpace(v)
	if strings.EqualFold(v, "asc") {
		return "asc"
	}
	return "desc"
}
