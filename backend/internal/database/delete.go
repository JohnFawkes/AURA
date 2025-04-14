package database

import "poster-setter/internal/logging"

func DeleteFromDatabase(ratingKey string) logging.ErrorLog {
	query := `
DELETE FROM auto_downloader WHERE id = ?`
	_, err := db.Exec(query, ratingKey)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to delete data from database",
		}}
	}
	return logging.ErrorLog{}
}
