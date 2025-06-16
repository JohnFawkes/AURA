import { log } from "@/lib/logger";

import {
	MediuxUserBoxset,
	MediuxUserCollectionMovie,
	MediuxUserCollectionSet,
	MediuxUserMovieSet,
	MediuxUserShowSet,
} from "@/types/mediuxUserAllSets";
import { PosterFileMovie, PosterFileShow, PosterSet } from "@/types/posterSets";

// Function - Convert Boxset Show Set to Poster Set
export const BoxsetShowToPosterSet = (showSet: MediuxUserShowSet) => {
	const showDetails: PosterFileShow = {
		ID: showSet.show_id.id,
		Title: showSet.show_id.title,
		MediaItem: showSet.show_id.MediaItem,
	};

	const posterSet: PosterSet = {
		ID: showSet.id,
		Title: showSet.set_title,
		Type: "show",
		User: {
			Name: showSet?.user_created?.username || "",
		},
		DateCreated: showSet.date_created,
		DateUpdated: showSet.date_updated,
		Poster:
			showSet.show_poster && showSet.show_poster.length > 0
				? {
						ID: showSet.show_poster[0].id,
						Type: "poster",
						Modified: showSet.show_poster[0].modified_on,
						FileSize: Number(showSet.show_poster[0].filesize),
						Show: showDetails,
					}
				: undefined,
		Backdrop:
			showSet.show_backdrop && showSet.show_backdrop.length > 0
				? {
						ID: showSet.show_backdrop[0].id,
						Type: "backdrop",
						Modified: showSet.show_backdrop[0].modified_on,
						FileSize: Number(showSet.show_backdrop[0].filesize),
						Show: showDetails,
					}
				: undefined,
		SeasonPosters: showSet.season_posters.map((poster) => ({
			ID: poster.id,
			Type: "seasonPoster",
			Modified: poster.modified_on,
			FileSize: Number(poster.filesize),
			Show: showDetails,
			Season: {
				Number: poster.season.season_number,
			},
		})),
		TitleCards: showSet.titlecards.map((titlecard) => ({
			ID: titlecard.id,
			Type: "titlecard",
			Modified: titlecard.modified_on,
			FileSize: Number(titlecard.filesize),
			Show: showDetails,
			Episode: {
				Title: titlecard.episode.episode_title,
				EpisodeNumber: titlecard.episode.episode_number,
				SeasonNumber: titlecard.episode.season_id.season_number,
			},
		})),
		Status: showSet.show_id.status,
	};
	return [posterSet];
};

// Function - Convert Boxset Movie Set to Poster Set
export const BoxsetMovieToPosterSet = (movieSet: MediuxUserMovieSet) => {
	const poster = movieSet.movie_poster[0] || undefined;
	const backdrop = movieSet.movie_backdrop[0] || undefined;

	if (!poster && !backdrop) {
		log("BoxsetMovieToPosterSet - No poster or backdrop found for movie set", movieSet);
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
					Movie: movie_details,
				}
			: undefined,
		Backdrop: backdrop
			? {
					ID: backdrop.id,
					Type: "backdrop",
					Modified: backdrop.modified_on,
					FileSize: Number(backdrop.filesize) || 0,
					Movie: movie_details,
				}
			: undefined,

		Status: movieSet.movie_id.status ?? "none",
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
					Movie: convertCollectionMovieToPosterFileMovie(firstBackdrop.movie),
				}
			: undefined,
		// Add remaining posters to OtherPosters
		OtherPosters: collectionSet.movie_posters.slice(1).map((poster) => ({
			ID: poster.id,
			Type: "poster",
			Modified: poster.modified_on,
			FileSize: Number(poster.filesize),
			Movie: convertCollectionMovieToPosterFileMovie(poster.movie),
		})),
		// Add remaining backdrops to OtherBackdrops
		OtherBackdrops: collectionSet.movie_backdrops.slice(1).map((backdrop) => ({
			ID: backdrop.id,
			Type: "backdrop",
			Modified: backdrop.modified_on,
			FileSize: Number(backdrop.filesize),
			Movie: convertCollectionMovieToPosterFileMovie(backdrop.movie),
		})),
		Status: firstPoster.movie.status ?? "none",
	};

	return [posterSet];
};

// Function - Convert Boxset to Poster Set
export const BoxsetToPosterSet = (boxset: MediuxUserBoxset) => {
	const posterSets: PosterSet[] = [];

	// Convert Show Sets
	if (boxset.show_sets && boxset.show_sets.length > 0) {
		boxset.show_sets.forEach((showSet) => {
			const posterSet = BoxsetShowToPosterSet(showSet);
			posterSet.every((set) => {
				set.User.Name = boxset.user_created.username;
			});
			posterSets.push(...posterSet);
		});
	}

	// Convert Movie Sets
	if (boxset.movie_sets && boxset.movie_sets.length > 0) {
		boxset.movie_sets.forEach((movieSet) => {
			const posterSet = BoxsetMovieToPosterSet(movieSet);
			posterSet.every((set) => {
				set.User.Name = boxset.user_created.username;
			});
			posterSets.push(...posterSet);
		});
	}

	// Convert Collection Sets
	if (boxset.collection_sets && boxset.collection_sets.length > 0) {
		boxset.collection_sets.forEach((collectionSet) => {
			const posterSet = BoxsetCollectionToPosterSet(collectionSet);
			posterSet.every((set) => {
				set.User.Name = boxset.user_created.username;
			});
			posterSets.push(...posterSet);
		});
	}

	log("BoxsetToPosterSet - Converted Boxset to Poster Sets", {
		boxset: boxset.boxset_title,
		posterSets: posterSets,
	});
	return posterSets;
};
