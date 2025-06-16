package database

import (
	"aura/internal/logging"
)

func CheckIfMediaItemExistsInDatabase(ratingKey string) (bool, logging.StandardError) {
	query := `
SELECT COUNT(*) FROM SavedItems WHERE media_item_id = ?`
	var count int
	err := db.QueryRow(query, ratingKey).Scan(&count)
	if err != nil {
		Err := logging.NewStandardError()
		Err.Message = "Failed to query database for media item"
		Err.HelpText = "Ensure the database connection is established and the query is correct."
		Err.Details = "Query: " + query + ", RatingKey: " + ratingKey
		return false, Err
	}
	if count > 0 {
		return true, logging.StandardError{}
	}
	return false, logging.StandardError{}
}

func CheckIfItemExistsInDatabase(ratingKey string, setID string) (bool, logging.StandardError) {
	query := `
SELECT COUNT(*) FROM SavedItems WHERE media_item_id = ? AND poster_set_id = ?`
	var count int
	err := db.QueryRow(query, ratingKey, setID).Scan(&count)
	if err != nil {
		Err := logging.NewStandardError()
		Err.Message = "Failed to query database for item"
		Err.HelpText = "Ensure the database connection is established and the query is correct."
		Err.Details = "Query: " + query + ", RatingKey: " + ratingKey + ", SetID: " + setID
		return false, Err
	}
	if count > 0 {
		return true, logging.StandardError{}
	}
	return false, logging.StandardError{}
}
