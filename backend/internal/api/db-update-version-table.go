package api

import (
	"aura/internal/logging"
	"database/sql"
	"fmt"
)

func DB_UpdateVersionTable(db *sql.DB, newVersion int) error {
	res, err := db.Exec(`UPDATE VERSION SET version = ?;`, newVersion)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		_, err = db.Exec(`INSERT INTO VERSION (version) VALUES (?);`, newVersion)
		if err != nil {
			return err
		}
	}
	logging.LOG.Info(fmt.Sprintf("Database version updated to %d", newVersion))
	return nil
}
