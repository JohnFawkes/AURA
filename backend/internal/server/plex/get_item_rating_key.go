package plex

import "aura/internal/modals"

func getItemRatingKey(plexItem modals.MediaItem, file modals.PosterFile) string {
	if plexItem.Type == "movie" || file.Type == "poster" || file.Type == "backdrop" {
		return plexItem.RatingKey
	}
	if file.Type == "seasonPoster" {
		return getSeasonRatingKey(plexItem, file)
	}
	if file.Type == "titlecard" {
		return getEpisodeRatingKey(plexItem, file)
	}
	return ""
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
