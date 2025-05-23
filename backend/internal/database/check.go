package database

import (
	"aura/internal/logging"
)

func CheckIfMediaItemAlreadyInDatabase(ratingKey string) (bool, logging.ErrorLog) {
	query := `
SELECT COUNT(*) FROM Media_Item WHERE id = ?`
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

func CheckIfPosterSetAlreadyInDatabase(setID string) (bool, logging.ErrorLog) {
	query := `
SELECT COUNT(*) FROM Poster_Sets WHERE id = ?`
	var count int
	err := db.QueryRow(query, setID).Scan(&count)
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
