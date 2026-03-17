package utils

import (
	"aura/models"
	"fmt"
)

func GetFileDownloadName(itemTitle string, imageFile models.ImageFile) string {
	switch imageFile.Type {
	case "poster":
		return "Poster"
	case "backdrop":
		return "Backdrop"
	case "season_poster", "special_season_poster":
		if imageFile.SeasonNumber != nil {
			if *imageFile.SeasonNumber == 0 {
				return "Special Season Poster"
			} else {
				return fmt.Sprintf("Season %s Poster", FormatIntAsTwoDigitString(*imageFile.SeasonNumber))
			}
		} else {
			return "Season Poster"
		}
	case "titlecard":
		if imageFile.SeasonNumber != nil && imageFile.EpisodeNumber != nil {
			return fmt.Sprintf("S%sE%s Titlecard", FormatIntAsTwoDigitString(*imageFile.SeasonNumber), FormatIntAsTwoDigitString(*imageFile.EpisodeNumber))
		} else if imageFile.SeasonNumber != nil {
			return fmt.Sprintf("S%s Titlecard", FormatIntAsTwoDigitString(*imageFile.SeasonNumber))
		} else if imageFile.EpisodeNumber != nil {
			return fmt.Sprintf("E%s Titlecard", FormatIntAsTwoDigitString(*imageFile.EpisodeNumber))
		} else {
			if imageFile.Title != "" {
				return fmt.Sprintf("%s Titlecard", imageFile.Title)
			}
			return "Titlecard"
		}
	default:
		return fmt.Sprintf("%s - Image", itemTitle)
	}
}
