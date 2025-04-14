package database

import "poster-setter/internal/logging"

func CheckIfAlreadyInDatabase(ratingKey string) (bool, logging.ErrorLog) {
	query := `
SELECT COUNT(*) FROM auto_downloader WHERE id = ?`
	var count int
	err := db.QueryRow(query, ratingKey).Scan(&count)
	if err != nil {
		return false, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to query database",
		}}
	}
	if count > 0 {
		return true, logging.ErrorLog{}
	}
	return false, logging.ErrorLog{}
}
