package database

import (
	"aura/logging"
	"context"
	"strings"
)

func (s *SQliteDB) IgnoreMediaItem(ctx context.Context, tmdbID, libraryTitle, mode string) (Err logging.LogErrorInfo) {
	Err = logging.LogErrorInfo{}

	if s == nil || s.conn == nil {
		return logging.LogErrorInfo{Message: "Database connection is nil"}
	}

	tmdbID = strings.TrimSpace(tmdbID)
	libraryTitle = strings.TrimSpace(libraryTitle)
	mode = strings.ToLower(strings.TrimSpace(mode))

	if tmdbID == "" || libraryTitle == "" {
		return logging.LogErrorInfo{
			Message: "tmdb_id and library_title are required",
			Detail:  map[string]any{"tmdb_id": tmdbID, "library_title": libraryTitle},
		}
	}

	if mode != "always" && mode != "temp" {
		return logging.LogErrorInfo{
			Message: "Invalid ignore mode",
			Detail:  map[string]any{"mode": mode},
		}
	}

	// Determine insert vs update for logging
	var existed int
	_ = s.conn.QueryRowContext(ctx, `
        SELECT 1
        FROM IgnoredItems
        WHERE tmdb_id = ? AND library_title = ?
        LIMIT 1;
    `, tmdbID, libraryTitle).Scan(&existed)
	op := "INSERT"
	if existed == 1 {
		op = "UPDATE"
	}

	_, err := s.conn.ExecContext(ctx, `
        INSERT INTO IgnoredItems (tmdb_id, library_title, mode)
        VALUES (?, ?, ?)
        ON CONFLICT(tmdb_id, library_title) DO UPDATE SET
            mode = excluded.mode;
    `, tmdbID, libraryTitle, mode)
	if err != nil {
		return logging.LogErrorInfo{
			Message: "Failed to ignore media item",
			Detail:  map[string]any{"error": err.Error(), "tmdb_id": tmdbID, "library_title": libraryTitle, "mode": mode},
		}
	}

	logging.LOGGER.Debug().Timestamp().
		Str("op", op).
		Str("table", "IgnoredItems").
		Str("tmdb_id", tmdbID).
		Str("library_title", libraryTitle).
		Str("mode", mode).
		Msg("Ignored media item")

	return Err
}
