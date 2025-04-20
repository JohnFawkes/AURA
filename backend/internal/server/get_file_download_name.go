package mediaserver

import (
	"fmt"
	"poster-setter/internal/modals"
	"poster-setter/internal/utils"
)

func getFileDownloadName(file modals.PosterFile) string {
	if file.Type == "poster" {
		return "Poster"
	} else if file.Type == "backdrop" {
		return "Backdrop"
	} else if file.Type == "seasonPoster" {
		return fmt.Sprintf("Season %s Poster", utils.Get2DigitNumber(int64(file.Season.Number)))
	} else if file.Type == "titlecard" {
		return fmt.Sprintf("S%sE%s - %s Titlecard", utils.Get2DigitNumber(int64(file.Episode.SeasonNumber)), utils.Get2DigitNumber(int64(file.Episode.EpisodeNumber)), file.Episode.Title)
	}
	return file.Type
}
