package database

import (
	"aura/logging"
	"context"
	"strings"
)

func (s *SQliteDB) StopIgnoringMediaItem(ctx context.Context, tmdbID, libraryTitle string) (Err logging.LogErrorInfo) {
	Err = logging.LogErrorInfo{}

	if s == nil || s.conn == nil {
		return logging.LogErrorInfo{Message: "Database connection is nil"}
	}

	tmdbID = strings.TrimSpace(tmdbID)
	libraryTitle = strings.TrimSpace(libraryTitle)

	res, err := s.conn.ExecContext(ctx, `
        DELETE FROM IgnoredItems
        WHERE tmdb_id = ? AND library_title = ?;
    `, tmdbID, libraryTitle)
	if err != nil {
		_, logAction := logging.AddSubActionToContext(ctx, "Stopping ignore for media item", logging.LevelError)
		defer logAction.Complete()
		logAction.SetError("Failed to delete ignore entry from database", err.Error(), map[string]any{
			"error":         err.Error(),
			"tmdb_id":       tmdbID,
			"library_title": libraryTitle,
		})
		return *logAction.Error
	}

	n, _ := res.RowsAffected()
	logging.LOGGER.Debug().Timestamp().
		Str("op", "DELETE").
		Str("table", "IgnoredItems").
		Int64("count", n).
		Str("tmdb_id", tmdbID).
		Str("library_title", libraryTitle).
		Msg("Stopped ignoring media item")

	return Err
}
