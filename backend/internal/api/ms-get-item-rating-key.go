package api

func Plex_GetItemRatingKey(plexItem MediaItem, file PosterFile) string {
	if plexItem.Type == "movie" || file.Type == "poster" || file.Type == "backdrop" {
		return plexItem.RatingKey
	}
	if file.Type == "seasonPoster" || file.Type == "specialSeasonPoster" {
		return Plex_GetSeasonRatingKey(plexItem, file)
	}
	if file.Type == "titlecard" {
		return Plex_GetEpisodeRatingKey(plexItem, file)
	}
	return ""
}

func Plex_GetSeasonRatingKey(plexItem MediaItem, file PosterFile) string {
	seasonNumberFromSet := file.Season.Number

	// Find the matching season Number in the Plex.Series.Seasons array where SeasonNumber == seasonNumberFromSet
	for _, season := range plexItem.Series.Seasons {
		if season.SeasonNumber == seasonNumberFromSet {
			return season.RatingKey
		}
	}
	return ""
}

func Plex_GetEpisodeRatingKey(plexItem MediaItem, file PosterFile) string {
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

func Plex_GetEpisodePathFromPlex(plexItem MediaItem, file PosterFile) string {
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

func EJ_GetItemRatingKey(embyJellyItem MediaItem, file PosterFile) string {
	if embyJellyItem.Type == "movie" || file.Type == "poster" || file.Type == "backdrop" {
		return embyJellyItem.RatingKey
	}
	if file.Type == "seasonPoster" || file.Type == "specialSeasonPoster" {
		return EJ_GetSeasonRatingKey(embyJellyItem, file)
	}
	if file.Type == "titlecard" {
		return EJ_GetEpisodeRatingKey(embyJellyItem, file)
	}
	return ""
}

func EJ_GetSeasonRatingKey(embyJellyItem MediaItem, file PosterFile) string {
	seasonNumberFromSet := file.Season.Number

	// Find the matching season Number in the EmbyJelly.Series.Seasons array where SeasonNumber == seasonNumberFromSet
	for _, season := range embyJellyItem.Series.Seasons {
		if season.SeasonNumber == seasonNumberFromSet {
			return season.RatingKey
		}
	}
	return ""
}

func EJ_GetEpisodeRatingKey(embyJellyItem MediaItem, file PosterFile) string {
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
