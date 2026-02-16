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
		return fmt.Sprintf("S%sE%s Titlecard", FormatIntAsTwoDigitString(*imageFile.SeasonNumber), FormatIntAsTwoDigitString(*imageFile.EpisodeNumber))
	default:
		return fmt.Sprintf("%s - Image", itemTitle)
	}
}
