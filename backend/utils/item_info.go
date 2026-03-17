package utils

import (
	"aura/models"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var titleYearSuffixRe = regexp.MustCompile(`\s*\((\d{4})\)\s*$`)

func MediaItemInfo(item models.MediaItem) string {
	title := item.Title
	year := item.Year
	libraryTitle := item.LibraryTitle
	tmdbID := item.TMDB_ID
	ratingKey := item.RatingKey

	// Null/zero checks
	if title == "" {
		title = "<no title>"
	}
	if year == 0 {
		year = -1
	}
	if libraryTitle == "" {
		libraryTitle = "<no library>"
	}
	if tmdbID == "" {
		tmdbID = "<no tmdb>"
	}
	if ratingKey == "" {
		ratingKey = "<no key>"
	}

	// If title ends with "(YYYY)", pull that year and strip it from the title.
	if m := titleYearSuffixRe.FindStringSubmatch(title); len(m) == 2 {
		if y, err := strconv.Atoi(m[1]); err == nil {
			year = y
			title = strings.TrimSpace(titleYearSuffixRe.ReplaceAllString(title, ""))
		}
	}

	return fmt.Sprintf("'%s' (%d) | %s [TMDB: %s | Key: %s]",
		title, year, libraryTitle, tmdbID, ratingKey,
	)
}

func CollectionItemInfo(item models.CollectionItem) string {
	title := item.Title
	libraryTitle := item.LibraryTitle
	ratingKey := item.RatingKey

	if title == "" {
		title = "<no title>"
	}
	if libraryTitle == "" {
		libraryTitle = "<no library>"
	}
	if ratingKey == "" {
		ratingKey = "<no key>"
	}

	return fmt.Sprintf("'%s' | %s [Key: %s]",
		title, libraryTitle, ratingKey,
	)
}
