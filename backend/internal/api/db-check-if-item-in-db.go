package api

import (
	logging "aura/internal/logging"
)

func DB_CheckIfMediaItemExists(TMDB_ID, libraryTitle string) (bool, []PosterSetSummary, logging.StandardError) {
	if db == nil {
		return false, nil, logging.StandardError{}
	}
	var results []PosterSetSummary
	query := `
SELECT s.PosterSetID, p.PosterSetUser, p.SelectedTypes
FROM SavedItems s
JOIN PosterSets p ON s.PosterSetID = p.PosterSetID AND s.TMDB_ID = p.TMDB_ID AND s.LibraryTitle = p.LibraryTitle
WHERE s.TMDB_ID = ? AND s.LibraryTitle = ?`
	rows, err := db.Query(query, TMDB_ID, libraryTitle)
	if err != nil {
		Err := logging.NewStandardError()
		Err.Message = "Failed to query database for media item"
		Err.HelpText = "Ensure the database connection is established and the query is correct."
		Err.Details = map[string]any{
			"error":        err.Error(),
			"query":        query,
			"TMDB_ID":      TMDB_ID,
			"libraryTitle": libraryTitle,
		}
		return false, nil, Err
	}
	defer rows.Close()
	for rows.Next() {
		var summary PosterSetSummary
		var selectedTypes string
		if err := rows.Scan(&summary.PosterSetID, &summary.PosterSetUser, &selectedTypes); err != nil {
			Err := logging.NewStandardError()
			Err.Message = "Failed to scan database row"
			Err.HelpText = "Check the database schema and ensure the query matches the expected columns."
			Err.Details = map[string]any{
				"error": err.Error(),
				"row":   summary,
			}
			return false, nil, Err
		}
		summary.SelectedTypes = []string{}
		if selectedTypes != "" {
			summary.SelectedTypes = append(summary.SelectedTypes, selectedTypes)
		}
		results = append(results, summary)
	}

	return len(results) > 0, results, logging.StandardError{}
}
