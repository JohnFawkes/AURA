package api

import "fmt"

func MediaServer_GetFileDownloadName(file PosterFile) string {
	switch file.Type {
	case "poster":
		return "Poster"
	case "backdrop":
		return "Backdrop"
	case "seasonPoster", "specialSeasonPoster":
		return fmt.Sprintf("Season %s Poster", Util_Format_Get2DigitNumber(int64(file.Season.Number)))
	case "titlecard":
		return fmt.Sprintf("S%sE%s - %s Titlecard", Util_Format_Get2DigitNumber(int64(file.Episode.SeasonNumber)), Util_Format_Get2DigitNumber(int64(file.Episode.EpisodeNumber)), file.Episode.Title)
	}
	return file.Type
}
