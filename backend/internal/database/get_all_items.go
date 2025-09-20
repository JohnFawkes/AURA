package database

import (
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func GetAllItemsWithFilter(w http.ResponseWriter, r *http.Request) {
	logging.LOG.Debug(r.URL.Path)
	startTime := time.Now()

	params := r.URL.Query()
	mediaItemID := params.Get("mediaItemID")
	mediaItemLibraryTitles := params["mediaItemLibraryTitles"]
	if len(mediaItemLibraryTitles) == 0 {
		mediaItemLibraryTitles = params["mediaItemLibraryTitles[]"]
	}
	if len(mediaItemLibraryTitles) == 0 {
		if mltStr := params.Get("mediaItemLibraryTitles"); mltStr != "" {
			mediaItemLibraryTitles = strings.Split(mltStr, ",")
		}
	}
	cleanedQuery := params.Get("cleanedQuery")
	mediaItemYear := 0
	if yearStr := params.Get("mediaItemYear"); yearStr != "" {
		fmt.Sscanf(yearStr, "%d", &mediaItemYear)
	}
	autoDownloadOnly := false
	if adoStr := params.Get("autoDownloadOnly"); adoStr == "true" {
		autoDownloadOnly = true
	}
	userNames := params["userNames"]
	if len(userNames) == 0 {
		userNames = params["userNames[]"]
	}
	if len(userNames) == 0 {
		if unStr := params.Get("userNames"); unStr != "" {
			userNames = strings.Split(unStr, ",")
		}
	}
	itemsPerPage := 20
	if ippStr := params.Get("itemsPerPage"); ippStr != "" {
		fmt.Sscanf(ippStr, "%d", &itemsPerPage)
	}
	pageNumber := 1
	if pnStr := params.Get("pageNumber"); pnStr != "" {
		fmt.Sscanf(pnStr, "%d", &pageNumber)
	}
	sortOption := "dateUpdated"
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

	items, totalItems, uniqueUsers, Err := GetAllItemsFromDatabaseWithFilter(
		mediaItemID,
		cleanedQuery,
		mediaItemLibraryTitles,
		mediaItemYear,
		autoDownloadOnly,
		userNames,
		itemsPerPage,
		pageNumber,
		sortOption,
		sortOrder,
	)
	if Err.Message != "" {
		utils.SendErrorResponse(w, utils.ElapsedTime(startTime), Err)
		return
	}

	logging.LOG.Debug(fmt.Sprintf("Retrieved %d items (Total Unique Items: %d, Unique Users: %d)", len(items), totalItems, len(uniqueUsers)))

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data: map[string]any{
			"items":        items,
			"total_items":  totalItems,
			"unique_users": uniqueUsers,
		},
	})
}

func GetAllItemsFromDatabaseWithFilter(
	mediaItemID string,
	cleanedQuery string,
	mediaItemLibraryTitles []string,
	mediaItemYear int,
	autoDownloadOnly bool,
	userNames []string,
	itemsPerPage, pageNumber int,
	sortOption string, sortOrder string,
) ([]modals.DBMediaItemWithPosterSets, int, []string, logging.StandardError) {
	Err := logging.NewStandardError()

	// Helper to build WHERE clauses and args, optionally skipping user filter
	buildWhere := func(skipUserFilter bool) ([]string, []any) {
		var clauses []string
		var args []any
		if mediaItemID != "" {
			clauses = append(clauses, "media_item_id = ?")
			args = append(args, mediaItemID)
			return clauses, args
		}
		if autoDownloadOnly {
			clauses = append(clauses, "auto_download = 1")
		}
		if len(mediaItemLibraryTitles) > 0 {
			var titleClauses []string
			for _, title := range mediaItemLibraryTitles {
				titleClauses = append(titleClauses, "json_extract(media_item, '$.LibraryTitle') = ?")
				args = append(args, title)
			}
			clauses = append(clauses, "("+strings.Join(titleClauses, " OR ")+")")
		}
		if mediaItemYear != 0 {
			clauses = append(clauses, "json_extract(media_item, '$.Year') = ?")
			args = append(args, mediaItemYear)
		}
		if cleanedQuery != "" {
			clauses = append(clauses, "LOWER(json_extract(media_item, '$.Title')) LIKE ?")
			args = append(args, "%"+strings.ToLower(cleanedQuery)+"%")
		}
		if !skipUserFilter && len(userNames) > 0 {
			var userClauses []string
			for _, name := range userNames {
				userClauses = append(userClauses, "json_extract(poster_set, '$.User.Name') = ?")
				args = append(args, name)
			}
			clauses = append(clauses, "("+strings.Join(userClauses, " OR ")+")")
		}
		return clauses, args
	}

	// --- COUNT QUERY (for total unique items) ---
	whereClauses, filterArgs := buildWhere(false)
	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}
	countQuery := `
SELECT COUNT(DISTINCT media_item_id)
FROM SavedItems
` + whereSQL

	var totalItems int
	if err := db.QueryRow(countQuery, filterArgs...).Scan(&totalItems); err != nil {
		Err.Message = "Failed to count unique media items"
		Err.HelpText = "Check error details for more information."
		Err.Details = "Error: " + err.Error() + ", Query: " + countQuery
		return nil, 0, nil, Err
	}

	// --- UNIQUE USERS QUERY (same filters, but skip user filter) ---
	uniqueUsersWhereClauses, uniqueUsersArgs := buildWhere(true)
	uniqueUsersWhereSQL := ""
	if len(uniqueUsersWhereClauses) > 0 {
		uniqueUsersWhereSQL = "WHERE " + strings.Join(uniqueUsersWhereClauses, " AND ")
	}
	uniqueUsersQuery := `
SELECT DISTINCT json_extract(poster_set, '$.User.Name')
FROM SavedItems
` + uniqueUsersWhereSQL

	rows, err := db.Query(uniqueUsersQuery, uniqueUsersArgs...)
	if err != nil {
		Err.Message = "Failed to get unique user names"
		Err.HelpText = "Check error details for more information."
		Err.Details = "Error: " + err.Error() + ", Query: " + uniqueUsersQuery
		return nil, 0, nil, Err
	}
	defer rows.Close()
	var uniqueUsers []string
	for rows.Next() {
		var user string
		if err := rows.Scan(&user); err == nil && user != "" {
			uniqueUsers = append(uniqueUsers, user)
		}
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY media_item_id"
	switch sortOption {
	case "title":
		orderBy = "ORDER BY LOWER(json_extract(media_item, '$.Title'))"
	case "dateUpdated":
		orderBy = "ORDER BY last_update"
	case "year":
		orderBy = "ORDER BY json_extract(media_item, '$.Year')"
	case "library":
		orderBy = "ORDER BY LOWER(json_extract(media_item, '$.LibraryTitle'))"
	}
	if sortOrder == "desc" {
		orderBy += " DESC"
	} else {
		orderBy += " ASC"
	}

	// --- MAIN QUERY (with LIMIT/OFFSET) ---
	limit := itemsPerPage
	offset := (pageNumber - 1) * itemsPerPage
	mainArgs := append(filterArgs, limit, offset)
	mainQuery := `
SELECT media_item_id, media_item, poster_set_id, poster_set, selected_types, auto_download, last_update
FROM SavedItems
` + whereSQL + `
` + orderBy + `
LIMIT ? OFFSET ?`

	logging.LOG.Debug(fmt.Sprintf("Executing query: %s with args: %v", mainQuery, mainArgs))

	rows2, err := db.Query(mainQuery, mainArgs...)
	if err != nil {
		Err.Message = "Failed to query filtered items from database"
		Err.HelpText = "Check error details for more information."
		Err.Details = "Error: " + err.Error() + ", Query: " + mainQuery
		return nil, 0, nil, Err
	}
	defer rows2.Close()

	result, groupErr := groupRowsIntoMediaItems(rows2, mainQuery)
	if groupErr.Message != "" {
		return nil, 0, nil, groupErr
	}

	return result, totalItems, uniqueUsers, logging.StandardError{}
}

func GetAllItemsFromDatabase() ([]modals.DBMediaItemWithPosterSets, logging.StandardError) {
	Err := logging.NewStandardError()

	// Query all rows from SavedItems.
	query := `
SELECT media_item_id, media_item, poster_set_id, poster_set, selected_types, auto_download, last_update
FROM SavedItems
ORDER BY media_item_id`
	rows, err := db.Query(query)
	if err != nil {
		Err.Message = "Failed to query all items from database"
		Err.HelpText = "Ensure the database connection is established and the query is correct."
		Err.Details = "Query: " + query
		return nil, Err
	}
	defer rows.Close()

	result, groupErr := groupRowsIntoMediaItems(rows, query)
	if groupErr.Message != "" {
		return nil, groupErr
	}

	return result, logging.StandardError{}
}

func groupRowsIntoMediaItems(rows *sql.Rows, query string) ([]modals.DBMediaItemWithPosterSets, logging.StandardError) {
	Err := logging.NewStandardError()
	var (
		result    []modals.DBMediaItemWithPosterSets
		currentID string
		current   *modals.DBMediaItemWithPosterSets
	)

	for rows.Next() {
		var savedItem modals.DBSavedItem
		var selectedTypesStr string
		if err := rows.Scan(
			&savedItem.MediaItemID,
			&savedItem.MediaItemJSON,
			&savedItem.PosterSetID,
			&savedItem.PosterSetJSON,
			&selectedTypesStr,
			&savedItem.AutoDownload,
			&savedItem.LastDownloaded,
		); err != nil {
			Err.Message = "Failed to scan row from SavedItems"
			Err.HelpText = "Check schema/data types."
			Err.Details = "Error: " + err.Error() + ", Query: " + query
			return nil, Err
		}

		// Start a new group when media_item_id changes
		if savedItem.MediaItemID != currentID {
			if current != nil {
				result = append(result, *current)
			}
			currentID = savedItem.MediaItemID

			var mediaItem modals.MediaItem
			if Err = UnmarshalMediaItem(savedItem.MediaItemJSON, &mediaItem); Err.Message != "" {
				return nil, Err
			}

			current = &modals.DBMediaItemWithPosterSets{
				MediaItemID:   savedItem.MediaItemID,
				MediaItem:     mediaItem,
				MediaItemJSON: savedItem.MediaItemJSON,
				PosterSets:    make([]modals.DBPosterSetDetail, 0, 4),
			}
		}

		var posterSet modals.PosterSet
		if Err = UnmarshalPosterSet(savedItem.PosterSetJSON, &posterSet); Err.Message != "" {
			return nil, Err
		}

		if selectedTypesStr != "" {
			savedItem.SelectedTypes = strings.Split(selectedTypesStr, ",")
		} else {
			savedItem.SelectedTypes = nil
		}

		psDetail := modals.DBPosterSetDetail{
			PosterSetID:    savedItem.PosterSetID,
			PosterSet:      posterSet,
			PosterSetJSON:  savedItem.PosterSetJSON,
			LastDownloaded: savedItem.LastDownloaded,
			SelectedTypes:  savedItem.SelectedTypes,
			AutoDownload:   savedItem.AutoDownload,
		}
		current.PosterSets = append(current.PosterSets, psDetail)
	}

	if current != nil {
		result = append(result, *current)
	}

	return result, logging.StandardError{}
}
