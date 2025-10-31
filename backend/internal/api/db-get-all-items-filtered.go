package api

import (
	"aura/internal/logging"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func DB_GetAllItemsWithFilter(
	ctx context.Context,
	searchTMDBID string,
	searchLibrary string,
	searchYear int,
	searchTitle string,
	librarySections []string,
	filteredTypes []string,
	filterAutodownload string,
	multisetOnly bool,
	filteredUsernames []string,
	itemsPerPage int,
	pageNumber int,
	sortOption string,
	sortOrder string,
) ([]DBMediaItemWithPosterSets, int, []string, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Getting all DB items with filters", logging.LevelDebug)
	defer logAction.Complete()

	var returnResult []DBMediaItemWithPosterSets
	var totalItems int
	var uniqueUsers []string

	// --- Build the Where Clauses ---
	whereClauses := []string{}
	queryArgs := []any{}

	// TMDB_ID filter
	tmdbClauses := buildWhereClause_SearchTMDB_ID(searchTMDBID, searchLibrary, &queryArgs)
	whereClauses = append(whereClauses, tmdbClauses...)

	if len(tmdbClauses) == 0 {
		libraryClauses := buildWhereClause_SearchLibrary(searchLibrary, &queryArgs)
		whereClauses = append(whereClauses, libraryClauses...)

		yearClauses := buildWhereClause_SearchYear(searchYear, &queryArgs)
		whereClauses = append(whereClauses, yearClauses...)

		titleClauses := buildWhereClause_SearchTitle(searchTitle, &queryArgs)
		whereClauses = append(whereClauses, titleClauses...)

		librarySectionClauses := buildWhereClause_LibrarySections(librarySections, &queryArgs)
		whereClauses = append(whereClauses, librarySectionClauses...)

		typeClauses := buildWhereClause_FilteredTypes(filteredTypes, &queryArgs)
		whereClauses = append(whereClauses, typeClauses...)

		autoDownloadClauses := buildWhereClause_FilterAutoDownload(filterAutodownload)
		whereClauses = append(whereClauses, autoDownloadClauses...)

		multiSetClauses := buildWhereClause_MultiSetOnly(multisetOnly)
		whereClauses = append(whereClauses, multiSetClauses...)

		userClauses := buildWhereClause_FilteredUserNames(filteredUsernames, &queryArgs)
		whereClauses = append(whereClauses, userClauses...)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "\nWHERE\t" + strings.Join(whereClauses, "\nAND\t")
	}

	// --- RUN: Unique Users Query ---
	uniqueUsersQuery := buildUniqueUsersQuery()
	uniqueUsers, Err := runUniqueUsersQuery(uniqueUsersQuery, queryArgs)
	if Err.Message != "" {
		return returnResult, totalItems, uniqueUsers, Err
	}

	// --- Build the Order By Clause ---
	orderByClause := buildOrderByClause(sortOption, sortOrder)

	// --- First Query: Get unique keys for pagination ---
	pageKeysQuery := fmt.Sprintf(`
SELECT s.TMDB_ID, s.LibraryTitle
FROM SavedItems s
JOIN MediaItems m ON s.TMDB_ID = m.TMDB_ID AND s.LibraryTitle = m.LibraryTitle
JOIN PosterSets p ON s.PosterSetID = p.PosterSetID AND s.TMDB_ID = p.TMDB_ID AND s.LibraryTitle = p.LibraryTitle
JOIN (
    SELECT TMDB_ID, LibraryTitle, MAX(LastDownloaded) AS MaxLastDownloaded
    FROM PosterSets
    GROUP BY TMDB_ID, LibraryTitle
) p_max ON p.TMDB_ID = p_max.TMDB_ID AND p.LibraryTitle = p_max.LibraryTitle AND p.LastDownloaded = p_max.MaxLastDownloaded
%s
GROUP BY s.TMDB_ID, s.LibraryTitle
%s
LIMIT ? OFFSET ?
`, whereSQL, orderByClause)

	pageKeysArgs := append([]any{}, queryArgs...)
	pageKeysArgs = append(pageKeysArgs, itemsPerPage, (pageNumber-1)*itemsPerPage)

	pagesQueryAction := logAction.AddSubAction("Getting Paginated Item Keys From DB", logging.LevelDebug)
	pagesQueryAction.AppendResult("query", pageKeysQuery)
	pagesQueryAction.AppendResult("query_args", pageKeysArgs)
	rows, err := db.Query(pageKeysQuery, pageKeysArgs...)
	if err != nil {
		pagesQueryAction.SetError("Failed to query paginated item keys from database",
			"Ensure the database is accessible and the query is correct.", map[string]any{
				"error": err.Error(),
				"query": pageKeysQuery,
				"args":  pageKeysArgs,
			})
		return returnResult, totalItems, uniqueUsers, *pagesQueryAction.Error
	}
	defer rows.Close()
	pagesQueryAction.Complete()

	// Collect keys for this page
	pageKeys := make([][2]string, 0, itemsPerPage)
	for rows.Next() {
		var tmdbID, libraryTitle string
		if err := rows.Scan(&tmdbID, &libraryTitle); err != nil {
			pagesQueryAction.SetError("Failed to scan row for paginated item keys", "Ensure the database is accessible and the query is correct.", map[string]any{
				"error": err.Error(),
				"query": pageKeysQuery,
				"args":  pageKeysArgs,
			})
			return nil, totalItems, uniqueUsers, *pagesQueryAction.Error
		}
		pageKeys = append(pageKeys, [2]string{tmdbID, libraryTitle})
	}
	if len(pageKeys) == 0 {
		return []DBMediaItemWithPosterSets{}, 0, uniqueUsers, logging.LogErrorInfo{}
	}

	// --- Second Query: Get all poster sets for those keys ---
	keyClauses := make([]string, 0, len(pageKeys))
	keyArgs := []any{}
	for _, key := range pageKeys {
		keyClauses = append(keyClauses, "(s.TMDB_ID = ? AND s.LibraryTitle = ?)")
		keyArgs = append(keyArgs, key[0], key[1])
	}
	keysWhereSQL := "WHERE " + strings.Join(keyClauses, " OR ")

	mainQuery := fmt.Sprintf(`
SELECT 
    s.TMDB_ID, s.LibraryTitle, s.PosterSetID,
    m.Title, m.ReleasedAt, m.Full_JSON,
    p.PosterSetUser, p.PosterSet_JSON, p.SelectedTypes, p.AutoDownload, p.LastDownloaded
FROM SavedItems s
JOIN MediaItems m ON s.TMDB_ID = m.TMDB_ID AND s.LibraryTitle = m.LibraryTitle
JOIN PosterSets p ON s.PosterSetID = p.PosterSetID AND s.TMDB_ID = p.TMDB_ID AND s.LibraryTitle = p.LibraryTitle
%s
%s
`, keysWhereSQL, orderByClause)

	mainQueryAction := logAction.AddSubAction("Getting All Item Details From DB", logging.LevelDebug)
	mainQueryAction.AppendResult("query", mainQuery)
	mainQueryAction.AppendResult("query_args", keyArgs)
	mainRows, err := db.Query(mainQuery, keyArgs...)
	if err != nil {
		mainQueryAction.SetError("Failed to query item details from database",
			"Ensure the database is accessible and the query is correct.", map[string]any{
				"error": err.Error(),
				"query": mainQuery,
				"args":  keyArgs,
			})
		return returnResult, totalItems, uniqueUsers, *mainQueryAction.Error
	}
	defer mainRows.Close()
	mainQueryAction.Complete()

	// --- Group Rows by Media Item ---
	returnResult, Err = DB_GroupRowsByMediaItem(ctx, mainRows, mainQuery)
	if Err.Message != "" {
		return returnResult, totalItems, uniqueUsers, Err
	}

	// --- Get total count of unique items ---
	countQuery := fmt.Sprintf(`
SELECT COUNT(DISTINCT s.TMDB_ID || '|' || s.LibraryTitle) as TotalItems
FROM SavedItems s
JOIN MediaItems m ON s.TMDB_ID = m.TMDB_ID AND s.LibraryTitle = m.LibraryTitle
JOIN PosterSets p ON s.PosterSetID = p.PosterSetID AND s.TMDB_ID = p.TMDB_ID AND s.LibraryTitle = p.LibraryTitle
%s
`, whereSQL)
	countQueryAction := logAction.AddSubAction("Getting Total Count of Unique Items From DB", logging.LevelDebug)
	countQueryAction.AppendResult("query", countQuery)
	countQueryAction.AppendResult("query_args", queryArgs)
	countRows, err := db.Query(countQuery, queryArgs...)
	if err != nil {
		countQueryAction.SetError("Failed to query total item count from database",
			"Ensure the database is accessible and the query is correct.", map[string]any{
				"error": err.Error(),
				"query": countQuery,
				"args":  queryArgs,
			})
		return returnResult, totalItems, uniqueUsers, *countQueryAction.Error
	}
	defer countRows.Close()
	countQueryAction.Complete()

	if countRows.Next() {
		if err := countRows.Scan(&totalItems); err != nil {
			totalItems = 0
		}
	}

	return returnResult, totalItems, uniqueUsers, Err
}

func buildOrderByClause(sortOption, sortOrder string) string {
	orderByClause := "ORDER BY m.TMDB_ID" // Default
	switch sortOption {
	case "title":
		orderByClause = "ORDER BY LOWER(m.Title)"
	case "year":
		orderByClause = `ORDER BY CASE WHEN m.ReleasedAt = '' OR m.ReleasedAt < '1' THEN m.Year ELSE m.ReleasedAt END`
	case "library":
		orderByClause = "ORDER BY LOWER(s.LibraryTitle)"
	case "dateDownloaded":
		orderByClause = "ORDER BY p.LastDownloaded"
	}
	if sortOrder == "asc" {
		orderByClause += " ASC"
	} else {
		orderByClause += " DESC"
	}
	return orderByClause
}

func buildWhereClause_SearchTMDB_ID(searchTMDBID, searchLibrary string, queryArgs *[]any) []string {
	tmdbClauses := []string{}
	if searchTMDBID != "" {
		tmdbClauses = append(tmdbClauses, "s.TMDB_ID = ?")
		*queryArgs = append(*queryArgs, searchTMDBID)
		libraryClauses := buildWhereClause_SearchLibrary(searchLibrary, queryArgs)
		tmdbClauses = append(tmdbClauses, libraryClauses...)
	}
	return tmdbClauses
}

func buildWhereClause_SearchLibrary(searchLibrary string, queryArgs *[]any) []string {
	libraryClauses := []string{}
	if searchLibrary != "" {
		libraryClauses = append(libraryClauses, "LOWER(s.LibraryTitle) LIKE ?")
		*queryArgs = append(*queryArgs, "%"+searchLibrary+"%")
	}
	return libraryClauses
}

func buildWhereClause_SearchYear(searchYear int, queryArgs *[]any) []string {
	yearClauses := []string{}
	if searchYear > 0 {
		yearClauses = append(yearClauses, "m.Year = ?")
		*queryArgs = append(*queryArgs, searchYear)
	}
	return yearClauses
}

func buildWhereClause_SearchTitle(searchTitle string, queryArgs *[]any) []string {
	titleClauses := []string{}
	if searchTitle != "" {
		titleClauses = append(titleClauses, "LOWER(m.Title) LIKE ?")
		*queryArgs = append(*queryArgs, "%"+searchTitle+"%")
	}
	return titleClauses
}

func buildWhereClause_LibrarySections(librarySections []string, queryArgs *[]any) []string {
	libraryClauses := []string{}
	if len(librarySections) > 0 {
		placeholders := strings.Repeat("?,", len(librarySections))
		placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
		libraryClauses = append(libraryClauses, fmt.Sprintf("s.LibraryTitle IN (%s)", placeholders))
		for _, lib := range librarySections {
			*queryArgs = append(*queryArgs, lib)
		}
	}
	return libraryClauses
}

func buildWhereClause_FilteredTypes(filteredTypes []string, queryArgs *[]any) []string {
	typeClauses := []string{}
	if len(filteredTypes) > 0 {
		hasNone := false
		var realTypes []string
		for _, t := range filteredTypes {
			if t == "None" || t == "none" {
				hasNone = true
			} else {
				realTypes = append(realTypes, t)
			}
		}
		if hasNone {
			typeClauses = append(typeClauses, "(p.SelectedTypes IS NULL OR p.SelectedTypes = '')")
		}
		for _, t := range realTypes {
			typeClauses = append(typeClauses, "(',' || p.SelectedTypes || ',' LIKE ?)")
			*queryArgs = append(*queryArgs, "%,"+t+",%")
		}
		if len(typeClauses) > 1 {
			combined := "(" + strings.Join(typeClauses, " OR ") + ")"
			typeClauses = []string{combined}
		}
	}
	return typeClauses
}

func buildWhereClause_FilterAutoDownload(filterAutoDownload string) []string {
	autoDownloadClauses := []string{}
	switch filterAutoDownload {
	case "on":
		autoDownloadClauses = append(autoDownloadClauses, "p.AutoDownload = 1")
	case "off":
		autoDownloadClauses = append(autoDownloadClauses, "p.AutoDownload = 0")
	case "none", "all":
		// No filter
	}
	return autoDownloadClauses
}

func buildWhereClause_FilteredUserNames(userNames []string, queryArgs *[]any) []string {
	userClauses := []string{}
	if len(userNames) > 0 {
		hasNoUser := false
		var realUsers []string
		for _, u := range userNames {
			if u == "|||no-user|||" {
				hasNoUser = true
			} else {
				realUsers = append(realUsers, u)
			}
		}
		if hasNoUser {
			userClauses = append(userClauses, "(p.PosterSetUser IS NULL OR p.PosterSetUser = '')")
		}
		for _, u := range realUsers {
			userClauses = append(userClauses, "(',' || p.PosterSetUser || ',' LIKE ?)")
			*queryArgs = append(*queryArgs, "%,"+u+",%")
		}
		if len(userClauses) > 1 {
			combined := "(" + strings.Join(userClauses, " OR ") + ")"
			userClauses = []string{combined}
		}
	}
	return userClauses
}

func buildWhereClause_MultiSetOnly(multisetOnly bool) []string {
	multiSetClauses := []string{}
	if multisetOnly {
		multiSetClauses = append(multiSetClauses, `
            EXISTS (
                SELECT 1
                FROM SavedItems si
                WHERE si.TMDB_ID = s.TMDB_ID
                  AND si.LibraryTitle = s.LibraryTitle
                GROUP BY si.TMDB_ID, si.LibraryTitle
                HAVING COUNT(*) > 1
            )
        `)
	}
	return multiSetClauses
}

func buildUniqueUsersQuery() string {
	uniqueUsersQuery := `
	SELECT DISTINCT LOWER(p.PosterSetUser) as UserName
FROM SavedItems s
JOIN PosterSets p ON s.PosterSetID = p.PosterSetID
ORDER BY UserName ASC
`
	return uniqueUsersQuery
}

func runUniqueUsersQuery(uniqueUsersQuery string, queryArgs []any) ([]string, logging.LogErrorInfo) {
	var uniqueUsers []string
	hasNoUser := false

	rows, err := db.Query(uniqueUsersQuery, queryArgs...)
	if err != nil {
		return uniqueUsers, logging.LogErrorInfo{
			Message: "Failed to query unique users from database",
			Help:    "Ensure the database is accessible and the query is correct.",
			Detail: map[string]any{
				"error": err.Error(),
				"query": uniqueUsersQuery,
				"args":  queryArgs,
			},
		}
	}
	defer rows.Close()

	for rows.Next() {
		var user sql.NullString
		if err := rows.Scan(&user); err == nil {
			if !user.Valid || strings.TrimSpace(user.String) == "" {
				hasNoUser = true
			} else {
				uniqueUsers = append(uniqueUsers, user.String)
			}
		}
	}
	if hasNoUser {
		uniqueUsers = append(uniqueUsers, "|||no-user|||")
	}
	return uniqueUsers, logging.LogErrorInfo{}
}

func DB_GroupRowsByMediaItem(ctx context.Context, rows *sql.Rows, query string) ([]DBMediaItemWithPosterSets, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Grouping rows by media item", logging.LevelTrace)
	defer logAction.Complete()

	resultOrder := []string{}
	resultMap := make(map[string]*DBMediaItemWithPosterSets)

	rowCount := 0
	for rows.Next() {
		var (
			tmdbID         string
			libraryTitle   string
			posterSetID    string
			title          string
			year           int
			mediaItemJSON  string
			posterSetUser  string
			posterSetJSON  string
			selectedTypes  sql.NullString
			autoDownload   bool
			lastDownloaded sql.NullString
		)

		if err := rows.Scan(
			&tmdbID,
			&libraryTitle,
			&posterSetID,
			&title,
			&year,
			&mediaItemJSON,
			&posterSetUser,
			&posterSetJSON,
			&selectedTypes,
			&autoDownload,
			&lastDownloaded,
		); err != nil {
			return nil, logging.LogErrorInfo{
				Message: "Failed to scan row from database",
				Help:    "Ensure the database is accessible and the query is correct.",
				Detail: map[string]any{
					"error": err.Error(),
					"query": query,
				},
			}
		}

		// Unmarshal media item if not already in map
		mediaItemGroup, exists := resultMap[tmdbID+"|"+libraryTitle]
		if !exists {
			var mediaItem MediaItem
			if Err := UnmarshalMediaItem(mediaItemJSON, &mediaItem); Err.Message != "" {
				return nil, Err
			}
			resultOrder = append(resultOrder, tmdbID+"|"+libraryTitle)
			mediaItemGroup = &DBMediaItemWithPosterSets{
				TMDB_ID:       tmdbID,
				LibraryTitle:  libraryTitle,
				MediaItem:     mediaItem,
				MediaItemJSON: mediaItemJSON,
				PosterSets:    make([]DBPosterSetDetail, 0, 4),
			}
			resultMap[tmdbID+"|"+libraryTitle] = mediaItemGroup
		}

		// Unmarshal poster set
		var posterSet PosterSet
		if Err := UnmarshalPosterSet(posterSetJSON, &posterSet); Err.Message != "" {
			return nil, Err
		}

		// Parse selected types
		var selectedTypesSlice []string
		if selectedTypes.Valid && selectedTypes.String != "" {
			selectedTypesSlice = strings.Split(selectedTypes.String, ",")
		}

		// Append poster set detail
		mediaItemGroup.PosterSets = append(mediaItemGroup.PosterSets, DBPosterSetDetail{
			PosterSetID:    posterSetID,
			PosterSet:      posterSet,
			PosterSetJSON:  posterSetJSON,
			LastDownloaded: lastDownloaded.String,
			SelectedTypes:  selectedTypesSlice,
			AutoDownload:   autoDownload,
		})

		rowCount++
	}

	if rowCount == 0 {
		return []DBMediaItemWithPosterSets{}, logging.LogErrorInfo{}
	} else {
		logAction.AppendResult("rows_processed", rowCount)
		logAction.AppendResult("unique_items", len(resultOrder))
		logAction.AppendResult("multiset_items", len(resultOrder)-len(resultMap))
	}
	// Build result slice in order
	result := make([]DBMediaItemWithPosterSets, 0, len(resultOrder))
	for _, id := range resultOrder {
		result = append(result, *resultMap[id])
	}

	return result, logging.LogErrorInfo{}

}

func DB_GetAllItems() ([]DBMediaItemWithPosterSets, logging.LogErrorInfo) {
	ctx, ld := logging.CreateLoggingContext(context.Background(), "AutoDownload - Get All Items")
	logAction := ld.AddAction("Fetching All Items from DB", logging.LevelDebug)
	ctx = logging.WithCurrentAction(ctx, logAction)
	defer ld.Log()

	const itemsPerPage = 500 // Use a reasonable batch size
	var allItems []DBMediaItemWithPosterSets

	// First, get the total number of items
	_, totalItems, _, Err := DB_GetAllItemsWithFilter(
		ctx,
		"",               // searchTMDBID
		"",               // searchLibrary
		0,                // searchYear
		"",               // searchTitle
		[]string{},       // librarySections
		[]string{},       // filteredTypes
		"all",            // filterAutoDownload
		false,            // multisetOnly
		[]string{},       // filteredUsernames
		1,                // itemsPerPage (just to get count)
		1,                // pageNumber
		"dateDownloaded", // sortOption
		"desc",           // sortOrder
	)
	if Err.Message != "" {
		return nil, Err
	}

	numPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	for page := 1; page <= numPages; page++ {
		items, _, _, pageErr := DB_GetAllItemsWithFilter(
			ctx,
			"",               // searchTMDBID
			"",               // searchLibrary
			0,                // searchYear
			"",               // searchTitle
			[]string{},       // librarySections
			[]string{},       // filteredTypes
			"all",            // filterAutoDownload
			false,            // multisetOnly
			[]string{},       // filteredUsernames
			itemsPerPage,     // itemsPerPage
			page,             // pageNumber
			"dateDownloaded", // sortOption
			"desc",           // sortOrder
		)
		if pageErr.Message != "" {
			return nil, pageErr
		}
		allItems = append(allItems, items...)
	}

	return allItems, logging.LogErrorInfo{}
}
