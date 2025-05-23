package emby_jellyfin

import "aura/internal/modals"

func getItemRatingKey(embyJellyItem modals.MediaItem, file modals.PosterFile) string {
	if embyJellyItem.Type == "movie" || file.Type == "poster" || file.Type == "backdrop" {
		return embyJellyItem.RatingKey
	}
	if file.Type == "seasonPoster" {
		return getSeasonRatingKey(embyJellyItem, file)
	}
	if file.Type == "titlecard" {
		return getEpisodeRatingKey(embyJellyItem, file)
	}
	return ""
}

func getSeasonRatingKey(embyJellyItem modals.MediaItem, file modals.PosterFile) string {
	seasonNumberFromSet := file.Season.Number

	// Find the matching season Number in the EmbyJelly.Series.Seasons array where SeasonNumber == seasonNumberFromSet
	for _, season := range embyJellyItem.Series.Seasons {
		if season.SeasonNumber == seasonNumberFromSet {
			return season.RatingKey
		}
	}
	return ""
}

func getEpisodeRatingKey(embyJellyItem modals.MediaItem, file modals.PosterFile) string {
	seasonNumberFromSet := file.Episode.SeasonNumber
	episodeNumberFromSet := file.Episode.EpisodeNumber

	for _, season := range embyJellyItem.Series.Seasons {
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
