import { MediaItem } from "@/types/mediaItem";

export interface MediuxUserAllSetsResponse {
	show_sets: MediuxUserShowSet[];
	movie_sets: MediuxUserMovieSet[];
	collection_sets: MediuxUserCollectionSet[];
	boxsets: MediuxUserBoxset[];
}

export interface MediuxUserShowSet {
	id: string;
	user_created: MediuxUserCreated;
	set_title: string;
	date_created: string; // ISO string or Date
	date_updated: string;
	show_id: MediuxUserShow;
	show_poster: MediuxUserImage[];
	show_backdrop: MediuxUserImage[];
	season_posters: MediuxUserSeasonPoster[];
	titlecards: MediuxUserTitlecard[];
	//MediaItem: MediaItem;
}

export interface MediuxUserMovieSet {
	id: string;
	user_created: MediuxUserCreated;
	set_title: string;
	date_created: string;
	date_updated: string;
	movie_id: MediuxUserMovie;
	movie_poster: MediuxUserImage[];
	movie_backdrop: MediuxUserImage[];
	//MediaItem: MediaItem;
}

export interface MediuxUserCollectionSet {
	id: string;
	user_created: MediuxUserCreated;
	set_title: string;
	date_created: string;
	date_updated: string;
	movie_posters: MediuxUserCollectionMovie[];
	movie_backdrops: MediuxUserCollectionMovie[];
}

export interface MediuxUserBoxset {
	id: string;
	user_created: MediuxUserCreated;
	boxset_title: string;
	date_created: string;
	date_updated: string;
	movie_sets: MediuxUserMovieSet[];
	show_sets: MediuxUserShowSet[];
	collection_sets: MediuxUserCollectionSet[];
}

// Reusable subtypes

export interface MediuxUserCreated {
	username: string;
}

export interface MediuxUserShow {
	id: string;
	date_updated: string;
	status: string;
	title: string;
	tagline: string;
	first_air_date: string;
	tvdb_id: string;
	imdb_id: string;
	trakt_id: string;
	slug: string;
	MediaItem: MediaItem;
}

export interface MediuxUserMovie {
	id: string;
	date_updated: string;
	status: string;
	title: string;
	tagline: string;
	release_date: string;
	tvdb_id: string;
	imdb_id: string;
	trakt_id: string;
	slug: string;
	MediaItem: MediaItem;
	//LibrarySection: string;
}

export interface MediuxUserImage {
	id: string;
	modified_on: string;
	filesize: string;
}

export interface MediuxUserSeasonPoster {
	id: string;
	modified_on: string;
	filesize: string;
	season: {
		season_number: number;
	};
}

export interface MediuxUserTitlecard {
	id: string;
	modified_on: string;
	filesize: string;
	episode: {
		episode_title: string;
		episode_number: number;
		season_id: {
			season_number: number;
		};
	};
}

export interface MediuxUserCollectionMovie {
	id: string;
	modified_on: string;
	filesize: string;
	movie: MediuxUserMovie;
}
