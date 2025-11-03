import { PosterSet } from "@/types/media-and-posters/poster-sets";
import { MediuxUserCollectionMovie, MediuxUserCollectionSet } from "@/types/mediux/mediux-sets";

export const BoxsetCollectionToPosterSet = (collectionSet: MediuxUserCollectionSet) => {
	// Get first poster and backdrop for primary display
	const firstPoster = collectionSet.movie_posters[0];
	const firstBackdrop = collectionSet.movie_backdrops[0];

	// Create single poster set
	const posterSet: PosterSet = {
		ID: collectionSet.id,
		Title: collectionSet.set_title,
		Type: "collection",
		User: {
			Name: collectionSet.user_created.username,
		},
		DateCreated: collectionSet.date_created,
		DateUpdated: collectionSet.date_updated,
		// Set primary poster
		Poster: firstPoster
			? {
					ID: firstPoster.id,
					Type: "poster",
					Modified: firstPoster.modified_on,
					FileSize: Number(firstPoster.filesize),
					Src: firstPoster.src,
					Blurhash: firstPoster.blurhash,
					Movie: convertCollectionMovieToPosterFileMovie(firstPoster.movie),
				}
			: undefined,
		// Set primary backdrop
		Backdrop: firstBackdrop
			? {
					ID: firstBackdrop.id,
					Type: "backdrop",
					Modified: firstBackdrop.modified_on,
					FileSize: Number(firstBackdrop.filesize),
					Src: firstBackdrop.src,
					Blurhash: firstBackdrop.blurhash,
					Movie: convertCollectionMovieToPosterFileMovie(firstBackdrop.movie),
				}
			: undefined,
		// Add remaining posters to OtherPosters
		OtherPosters: collectionSet.movie_posters.slice(1).map((poster) => ({
			ID: poster.id,
			Type: "poster",
			Modified: poster.modified_on,
			FileSize: Number(poster.filesize),
			Src: poster.src,
			Blurhash: poster.blurhash,
			Movie: convertCollectionMovieToPosterFileMovie(poster.movie),
		})),
		// Add remaining backdrops to OtherBackdrops
		OtherBackdrops: collectionSet.movie_backdrops.slice(1).map((backdrop) => ({
			ID: backdrop.id,
			Type: "backdrop",
			Modified: backdrop.modified_on,
			FileSize: Number(backdrop.filesize),
			Src: backdrop.src,
			Blurhash: backdrop.blurhash,
			Movie: convertCollectionMovieToPosterFileMovie(backdrop.movie),
		})),
		Status: firstPoster.movie.status ?? "none",
	};

	return [posterSet];
};

const convertCollectionMovieToPosterFileMovie = (movie: MediuxUserCollectionMovie["movie"]) => ({
	ID: movie.id,
	Title: movie.title,
	Status: movie.status,
	Tagline: movie.tagline,
	Slug: movie.slug,
	DateUpdated: movie.date_updated,
	TVbdID: movie.tvdb_id,
	ImdbID: movie.imdb_id,
	TraktID: movie.trakt_id,
	ReleaseDate: movie.release_date,
	MediaItem: movie.MediaItem,
});
