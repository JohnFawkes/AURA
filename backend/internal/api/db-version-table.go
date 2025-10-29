package api

import (
	"aura/internal/logging"
	"context"
	"database/sql"
)

func DB_UpdateVersionTable(ctx context.Context, db *sql.DB, newVersion int) logging.LogErrorInfo {
	res, err := db.Exec(`UPDATE VERSION SET version = ?;`, newVersion)
	if err != nil {
		_, logAction := logging.AddSubActionToContext(ctx, "Updating VERSION Table", logging.LevelError)
		defer logAction.Complete()
		logAction.SetError(
			"Failed to update VERSION Table", err.Error(), nil)
		return logging.LogErrorInfo{Message: "Failed to update VERSION table"}
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		_, err = db.Exec(`INSERT INTO VERSION (version) VALUES (?);`, newVersion)
		if err != nil {
			_, logAction := logging.AddSubActionToContext(ctx, "Inserting into VERSION Table", logging.LevelError)
			defer logAction.Complete()
			logAction.SetError(
				"Failed to insert into VERSION Table", err.Error(), nil)

			return logging.LogErrorInfo{Message: "Failed to insert into VERSION table"}
		}
	}

	return logging.LogErrorInfo{}
}
