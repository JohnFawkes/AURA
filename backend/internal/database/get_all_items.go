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
	logging.LOG.Trace(fmt.Sprintf("Query Params: %v", params))
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
	filteredTypes := params["filteredTypes"]
	if len(filteredTypes) == 0 {
		filteredTypes = params["filteredTypes[]"]
	}
	if len(filteredTypes) == 0 {
		if ftStr := params.Get("filteredTypes"); ftStr != "" {
			filteredTypes = strings.Split(ftStr, ",")
		}
	}
	filterMultiSetOnly := false
	if msStr := params.Get("filterMultiSetOnly"); msStr == "true" {
		filterMultiSetOnly = true
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
		filteredTypes,
		filterMultiSetOnly,
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
	filteredTypes []string,
	filterMultiSetOnly bool,
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
		if len(filteredTypes) > 0 {
			var typeClauses []string
			hasNone := false
			var realTypes []string
			for _, t := range filteredTypes {
				if t == "none" {
					hasNone = true
				} else {
					realTypes = append(realTypes, t)
				}
			}
			if hasNone {
				// Match rows where selected_types is NULL or empty
				typeClauses = append(typeClauses, "(selected_types IS NULL OR TRIM(selected_types) = '')")
			}
			for _, t := range realTypes {
				typeClauses = append(typeClauses, "(',' || selected_types || ',') LIKE ?")
				args = append(args, "%,"+t+",%")
			}
			if len(typeClauses) > 0 {
				clauses = append(clauses, "("+strings.Join(typeClauses, " OR ")+")")
			}
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

	var totalItems int

	// --- MULTI-SET FILTER LOGIC ---
	if filterMultiSetOnly {
		// 1. Get media_item_ids with more than one poster set (with pagination)
		countQuery := `
SELECT COUNT(*) FROM (
    SELECT media_item_id
    FROM SavedItems
    ` + whereSQL + `
    GROUP BY media_item_id
    HAVING COUNT(*) > 1
)
`
		if err := db.QueryRow(countQuery, filterArgs...).Scan(&totalItems); err != nil {
			Err.Message = "Failed to count multi-set media items"
			Err.HelpText = "Check error details for more information."
			Err.Details = "Error: " + err.Error() + ", Query: " + countQuery
			return nil, 0, nil, Err
		}

		limit := itemsPerPage
		offset := (pageNumber - 1) * itemsPerPage

		// 2. Get paginated media_item_ids
		idQuery := `
SELECT media_item_id
FROM SavedItems
` + whereSQL + `
GROUP BY media_item_id
HAVING COUNT(*) > 1
` + fmt.Sprintf("ORDER BY media_item_id LIMIT %d OFFSET %d", limit, offset)

		idRows, err := db.Query(idQuery, filterArgs...)
		if err != nil {
			Err.Message = "Failed to query multi-set media_item_ids"
			Err.HelpText = "Check error details for more information."
			Err.Details = "Error: " + err.Error() + ", Query: " + idQuery
			return nil, 0, nil, Err
		}
		defer idRows.Close()

		var ids []string
		for idRows.Next() {
			var id string
			if err := idRows.Scan(&id); err == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil, totalItems, nil, logging.StandardError{}
		}

		// 3. Fetch all poster sets for those media_item_ids
		placeholders := strings.Repeat("?,", len(ids))
		placeholders = placeholders[:len(placeholders)-1]
		mainQuery := fmt.Sprintf(`
SELECT media_item_id, media_item, poster_set_id, poster_set, selected_types, auto_download, last_update
FROM SavedItems
WHERE media_item_id IN (%s)
`, placeholders)

		mainArgs := make([]any, len(ids))
		for i, id := range ids {
			mainArgs[i] = id
		}

		rows2, err := db.Query(mainQuery, mainArgs...)
		if err != nil {
			Err.Message = "Failed to query multi-set items"
			Err.HelpText = "Check error details for more information."
			Err.Details = "Error: " + err.Error() + ", Query: " + mainQuery
			return nil, 0, nil, Err
		}
		defer rows2.Close()

		result, groupErr := groupRowsIntoMediaItems(rows2, mainQuery)
		if groupErr.Message != "" {
			return nil, 0, nil, groupErr
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

		userRows, err := db.Query(uniqueUsersQuery, uniqueUsersArgs...)
		if err != nil {
			Err.Message = "Failed to get unique user names"
			Err.HelpText = "Check error details for more information."
			Err.Details = "Error: " + err.Error() + ", Query: " + uniqueUsersQuery
			return nil, 0, nil, Err
		}
		defer userRows.Close()
		var uniqueUsers []string
		for userRows.Next() {
			var user string
			if err := userRows.Scan(&user); err == nil && user != "" {
				uniqueUsers = append(uniqueUsers, user)
			}
		}

		return result, totalItems, uniqueUsers, logging.StandardError{}
	}

	// --- NORMAL (non-multiset) LOGIC ---
	countQuery := `
SELECT COUNT(DISTINCT media_item_id)
FROM SavedItems
` + whereSQL

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

func groupRowsIntoMediaItems(rows *sql.Rows, query string) ([]modals.DBMediaItemWithPosterSets, logging.StandardError) {
	Err := logging.NewStandardError()
	resultMap := make(map[string]*modals.DBMediaItemWithPosterSets)

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

		// Unmarshal media item if not already in map
		mediaItemGroup, exists := resultMap[savedItem.MediaItemID]
		if !exists {
			var mediaItem modals.MediaItem
			if Err = UnmarshalMediaItem(savedItem.MediaItemJSON, &mediaItem); Err.Message != "" {
				return nil, Err
			}
			mediaItemGroup = &modals.DBMediaItemWithPosterSets{
				MediaItemID:   savedItem.MediaItemID,
				MediaItem:     mediaItem,
				MediaItemJSON: savedItem.MediaItemJSON,
				PosterSets:    make([]modals.DBPosterSetDetail, 0, 4),
			}
			resultMap[savedItem.MediaItemID] = mediaItemGroup
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

		mediaItemGroup.PosterSets = append(mediaItemGroup.PosterSets, psDetail)
	}

	// Convert map to slice
	result := make([]modals.DBMediaItemWithPosterSets, 0, len(resultMap))
	for _, v := range resultMap {
		result = append(result, *v)
	}

	return result, logging.StandardError{}
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
