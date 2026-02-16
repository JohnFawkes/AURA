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

	// If title ends with "(YYYY)", pull that year and strip it from the title.
	if m := titleYearSuffixRe.FindStringSubmatch(title); len(m) == 2 {
		if y, err := strconv.Atoi(m[1]); err == nil {
			year = y
			title = strings.TrimSpace(titleYearSuffixRe.ReplaceAllString(title, ""))
		}
	}

	return fmt.Sprintf("'%s' (%d) | %s [TMDB: %s | Key: %s]",
		title, year, item.LibraryTitle, item.TMDB_ID, item.RatingKey,
	)
}

func CollectionItemInfo(item models.CollectionItem) string {
	return fmt.Sprintf("'%s' | %s [Key: %s]",
		item.Title, item.LibraryTitle, item.RatingKey,
	)
}
