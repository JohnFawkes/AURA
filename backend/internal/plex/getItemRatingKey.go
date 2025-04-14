package plex

import "poster-setter/internal/modals"

func getItemRatingKey(plex modals.MediaItem, file modals.PosterFile) string {
	if plex.Type == "movie" || file.Type == "poster" || file.Type == "backdrop" {
		return plex.RatingKey
	}
	if file.Type == "seasonPoster" {
		return getSeasonRatingKey(plex, file)
	}
	if file.Type == "titlecard" {
		return getEpisodeRatingKey(plex, file)
	}
	return ""
}
