package api

import (
	logging "aura/internal/logging"
	"context"
)

func DB_CheckIfMediaItemExists(ctx context.Context, TMDB_ID, libraryTitle string) (bool, []PosterSetSummary, logging.LogErrorInfo) {
	if db == nil {
		return false, nil, logging.LogErrorInfo{}
	}

	var results []PosterSetSummary
	query := `
SELECT s.PosterSetID, p.PosterSetUser, p.SelectedTypes
FROM SavedItems s
JOIN PosterSets p ON s.PosterSetID = p.PosterSetID AND s.TMDB_ID = p.TMDB_ID AND s.LibraryTitle = p.LibraryTitle
WHERE s.TMDB_ID = ? AND s.LibraryTitle = ?`
	rows, err := db.Query(query, TMDB_ID, libraryTitle)
	if err != nil {
		_, logAction := logging.AddSubActionToContext(ctx, "Checking if media item exists in database", logging.LevelError)
		defer logAction.Complete()
		logAction.SetError("Failed to query database for media item", err.Error(), map[string]any{
			"error":        err.Error(),
			"query":        query,
			"TMDB_ID":      TMDB_ID,
			"libraryTitle": libraryTitle,
		})
		return false, nil, *logAction.Error
	}
	defer rows.Close()
	for rows.Next() {
		var summary PosterSetSummary
		var selectedTypes string
		if err := rows.Scan(&summary.PosterSetID, &summary.PosterSetUser, &selectedTypes); err != nil {
			_, logAction := logging.AddSubActionToContext(ctx, "Checking if media item exists in database", logging.LevelError)
			defer logAction.Complete()
			logAction.SetError("Failed to scan database row", err.Error(), map[string]any{
				"error": err.Error(),
				"row":   summary,
			})
			return false, nil, *logAction.Error
		}
		summary.SelectedTypes = []string{}
		if selectedTypes != "" {
			summary.SelectedTypes = append(summary.SelectedTypes, selectedTypes)
		}
		results = append(results, summary)
	}

	return len(results) > 0, results, logging.LogErrorInfo{}
}
