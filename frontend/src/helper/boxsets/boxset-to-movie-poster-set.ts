import { log } from "@/lib/logger";

import { PosterFileMovie, PosterSet } from "@/types/media-and-posters/poster-sets";
import { MediuxUserMovieSet } from "@/types/mediux/mediux-sets";

// Function - Convert Boxset Movie Set to Poster Set
export const BoxsetMovieToPosterSet = (movieSet: MediuxUserMovieSet) => {
	const poster = movieSet.movie_poster[0] || undefined;
	const backdrop = movieSet.movie_backdrop[0] || undefined;

	if (!poster && !backdrop) {
		log("WARN", "Boxset", "BoxsetMovieToPosterSet", "No poster or backdrop found for movie set", {
			movieSet,
		});
		return [];
	}

	const movie_details: PosterFileMovie = {
		ID: movieSet.movie_id.id,
		Title: movieSet.movie_id.title,
		Status: movieSet.movie_id.status,
		Tagline: movieSet.movie_id.tagline,
		Slug: movieSet.movie_id.slug,
		DateUpdated: movieSet.movie_id.date_updated,
		TVbdID: movieSet.movie_id.tvdb_id,
		ImdbID: movieSet.movie_id.imdb_id,
		TraktID: movieSet.movie_id.trakt_id,
		ReleaseDate: movieSet.movie_id.release_date,
		MediaItem: movieSet.movie_id.MediaItem,
	};

	const posterSet: PosterSet = {
		ID: movieSet.id,
		Title: movieSet.set_title,
		Type: "movie",
		User: {
			Name: movieSet.user_created.username,
		},
		DateCreated: movieSet.date_created,
		DateUpdated: movieSet.date_updated,
		Poster: poster
			? {
					ID: poster.id,
					Type: "poster",
					Modified: poster.modified_on,
					FileSize: Number(poster.filesize) || 0,
					Src: poster.src,
					Blurhash: poster.blurhash,
					Movie: movie_details,
				}
			: undefined,
		Backdrop: backdrop
			? {
					ID: backdrop.id,
					Type: "backdrop",
					Modified: backdrop.modified_on,
					FileSize: Number(backdrop.filesize) || 0,
					Src: backdrop.src,
					Blurhash: backdrop.blurhash,
					Movie: movie_details,
				}
			: undefined,

		Status: movieSet.movie_id.status ?? "none",
	};

	return [posterSet];
};
