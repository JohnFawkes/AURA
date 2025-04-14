package plex

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

func getSeasonRatingKey(plexItem modals.MediaItem, file modals.PosterFile) string {
	seasonNumberFromSet := file.Season.Number

	// Find the matching season Number in the Plex.Series.Seasons array where SeasonNumber == seasonNumberFromSet
	for _, season := range plexItem.Series.Seasons {
		if season.SeasonNumber == seasonNumberFromSet {
			return season.RatingKey
		}
	}
	return ""
}

func getEpisodeRatingKey(plexItem modals.MediaItem, file modals.PosterFile) string {
	seasonNumberFromSet := file.Episode.SeasonNumber
	episodeNumberFromSet := file.Episode.EpisodeNumber

	for _, season := range plexItem.Series.Seasons {
		if season.SeasonNumber == seasonNumberFromSet {
			for _, episode := range season.Episodes {
				if episodeNumberFromSet == episode.EpisodeNumber && seasonNumberFromSet == episode.SeasonNumber {
					return episode.RatingKey
				}
			}
		}
	}
	return ""
}

func getEpisodePathFromPlex(plexItem modals.MediaItem, file modals.PosterFile) string {
	seasonNumberFromSet := file.Episode.SeasonNumber
	episodeNumberFromSet := file.Episode.EpisodeNumber

	for _, season := range plexItem.Series.Seasons {
		if season.SeasonNumber == seasonNumberFromSet {
			for _, episode := range season.Episodes {
				if episodeNumberFromSet == episode.EpisodeNumber && seasonNumberFromSet == episode.SeasonNumber {
					return episode.File.Path
				}
			}
		}
	}
	return ""
}
