// // Content types - Shows, Movies, Collections
// import type { AssetFile } from "./files";
// import type { CollectionSet, MovieSet, ShowSet } from "./sets";

// // Shows
// export interface Show {
// 	id: string; // varchar(255)
// 	date_created: string | null; // timestamp NULL
// 	date_updated: string | null; // timestamp NULL
// 	status: string | null; // varchar(255) NULL [NULL]
// 	title: string | null; // varchar(255) NULL [Unknown Show]
// 	tagline: string | null; // varchar(255) NULL [NULL]
// 	first_air_date: string | null; // date NULL
// 	backdrop: string | null; // varchar(255) NULL [NULL]
// 	tvdb_id: string | null; // varchar(255) NULL [NULL]
// 	imdb_id: string | null; // varchar(255) NULL [NULL]
// 	trakt_id: string | null; // varchar(255) NULL [NULL]
// 	slug: string | null; // varchar(255) NULL [NULL]

// 	// Relationships and additional fields
// 	seasons?: Season[];
// 	show_sets?: ShowSet[];

// 	showPosters?: AssetFile[];
// 	showBackdrops?: AssetFile[];
// 	showLogos?: AssetFile[]; // Alias of AssetFile type
// 	posters?: AssetFile[];
// 	backdrops?: AssetFile[];
// 	albums?: AssetFile[];
// 	logos?: AssetFile[];
// 	sets?: Array<{ id: string }>;
// }

// export interface Season {
// 	id: string; // varchar(255)
// 	date_created: string | null; // timestamp NULL
// 	date_updated: string | null; // timestamp NULL
// 	season_number: number; // int(11) NOT NULL
// 	season_name: string | null; // varchar(255) NULL [NULL]
// 	show_id: Pick<Show, "id" | "title"> | null; // TYPE OF Show
// 	episodes: Episode[]; // ARRAY TYPE OF Episode
// 	files: AssetFile[]; // ARRAY TYPE OF AssetFile

// 	// Additional fields for compatibility
// 	seasonPosters?: AssetFile[];
// 	seasonAlbums?: AssetFile[];
// }

// export interface Episode {
// 	id: string; // varchar(255)
// 	date_created: string | null; // timestamp NULL
// 	date_updated: string | null; // timestamp NULL
// 	episode_title: string | null; // varchar(255) NULL
// 	episode_number: number | null; // int(11) NULL
// 	show_id: Pick<Show, "id" | "title"> | null; // TYPE OF Show
// 	season_id: Pick<Season, "id" | "season_number"> | null; // TYPE OF Season
// 	air_date: string | null; // date NULL
// 	files: AssetFile[]; // ARRAY TYPE OF AssetFile

// 	// Additional fields for compatibility
// 	name?: string;
// }

// // Movies
// export interface Movie {
// 	id: string; // varchar(255)
// 	date_created: string | null; // timestamp NULL
// 	date_updated: string | null; // timestamp NULL
// 	status: string | null; // varchar(255) NULL
// 	title: string | null; // varchar(255) NULL
// 	tagline: string | null; // varchar(255) NULL
// 	release_date: string | null; // date NULL
// 	backdrop: string | null; // varchar(255) NULL
// 	movie_sets: MovieSet[] | null;
// 	// Relationships and additional fields
// 	collection_id?: Collection | null;
// 	sets?: Array<{ id: string }>;
// 	posters?: AssetFile[];
// 	backdrops?: AssetFile[];
// 	albums?: AssetFile[];
// 	logos?: AssetFile[];
// 	moviePosters?: AssetFile[];
// 	movieBackdrops?: AssetFile[];
// 	movieAlbums?: AssetFile[];
// 	movieLogos?: AssetFile[];
// }

// // Collections
// export interface Collection {
// 	id: string; // varchar(255)
// 	date_created: string | null; // timestamp NULL
// 	date_updated: string | null; // timestamp NULL
// 	collection_name: string | null; // varchar(255) NULL
// 	overview: string | null; // text NULL
// 	backdrop: string | null; // varchar(255) NULL

// 	// Relationships and additional fields
// 	belongs_to_collection?: boolean;
// 	movies?: Movie[];
// 	sets?: Array<{ id: string }>;
// 	collectionPosters?: AssetFile[];
// 	collectionBackdrops?: AssetFile[];
// 	collectionLogos?: AssetFile[];
// 	collection_sets?: CollectionSet[];
// 	posters?: AssetFile[];
// 	backdrops?: AssetFile[];
// 	logos?: AssetFile[];
// }

// // Content type constants
// export type ContentCarouselType = "show" | "movie" | "collection";
// export type SortDirection = "asc" | "desc";
// export type SortField = "release_date" | "date_created";
