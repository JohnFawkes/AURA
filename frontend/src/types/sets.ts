// Set-related types
import type { User } from "./user";
import type { Language, AssetFile } from "./files";
import type { Movie, Show, Collection } from "./content";

// Forward declarations to avoid circular dependencies
export interface ShowSet {
  id: number;
  user_created: User;
  date_created: string | null;
  user_updated: User;
  date_updated: string | null;
  set_title: string | null;
  description: string | null;
  show_id: Show;
  boxset_id: Boxset | null;
  language: Language;
  files: AssetFile[];

  showPoster?: AssetFile[];
  showBackdrop?: AssetFile[];
  seasonPosters?: AssetFile[];
  titlecards?: AssetFile[];
}

export interface MovieSet {
  id: number;
  user_created: User;
  date_created: string | null;
  user_updated: User;
  date_updated: string | null;
  set_title: string | null;
  description: string | null;
  movie_id: Movie;
  boxset_id: Boxset | null;
  language: Language;
  files: AssetFile[];

  moviePoster?: AssetFile[];
  movieBackdrop?: AssetFile[];
  movieLogo?: AssetFile[];
  movieAlbumArt?: AssetFile[];
}

export interface CollectionSet {
  id: number;
  user_created: User;
  date_created: string | null;
  user_updated: User;
  date_updated: string | null;
  set_title: string | null;
  description: string | null;
  collection_id: Collection;
  boxset_id: Boxset | null;
  language: Language;
  files: AssetFile[];

  collectionPoster?: AssetFile[];
  collectionBackdrop?: AssetFile[];
  moviePosters?: AssetFile[];
  movieBackdrops?: AssetFile[];
  movieLogos?: AssetFile[];
  movieAlbumArt?: AssetFile[];
}

// Define Boxset structure
export interface Boxset {
  id: number;
  user_created: User;
  date_created: string | null;
  user_updated: User;
  date_updated: string | null;
  boxset_title: string | null;
  description: string | null;
  poster: AssetFile | null;
  show_sets: ShowSet[];
  collection_sets: CollectionSet[];
  movie_sets: MovieSet[];
}
