package database

import (
	"aura/internal/logging"
)

func CheckIfMediaItemExistsInDatabase(ratingKey string) (bool, logging.ErrorLog) {
	query := `
SELECT COUNT(*) FROM SavedItems WHERE media_item_id = ?`
	var count int
	err := db.QueryRow(query, ratingKey).Scan(&count)
	if err != nil {
		return false, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to query database for media item",
		}}
	}
	if count > 0 {
		return true, logging.ErrorLog{}
	}
	return false, logging.ErrorLog{}
}

func CheckIfItemExistsInDatabase(ratingKey string, setID string) (bool, logging.ErrorLog) {
	query := `
SELECT COUNT(*) FROM SavedItems WHERE media_item_id = ? AND poster_set_id = ?`
	var count int
	err := db.QueryRow(query, ratingKey, setID).Scan(&count)
	if err != nil {
		return false, logging.ErrorLog{Err: err, Log: logging.Log{
			Message: "Failed to query database for media item and set ID",
		}}
	}
	if count > 0 {
		return true, logging.ErrorLog{}
	}
	return false, logging.ErrorLog{}
}
