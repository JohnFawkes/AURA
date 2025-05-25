import { LibrarySection } from "./mediaItem";

export interface PosterSet {
	ID: string;
	Title: string;
	Type: string;
	User: {
		Name: string;
	};
	DateCreated: string;
	DateUpdated: string;
	Poster?: PosterFile;
	OtherPosters?: PosterFile[];
	Backdrop?: PosterFile;
	OtherBackdrops?: PosterFile[];
	SeasonPosters?: PosterFile[];
	TitleCards?: PosterFile[];
	Status: string;
}

export interface PosterFile {
	ID: string;
	Type: string;
	Modified: string;
	FileSize: number;
	Movie?: PosterFileMovie;
	Season?: PosterFileSeason;
	Episode?: PosterFileEpisode;
}

export interface PosterFileMovie {
	ID: string;
	Title: string;
	Status: string;
	Tagline: string;
	Slug: string;
	DateUpdated: string;
	TVbdID: string;
	ImdbID: string;
	TraktID: string;
	ReleaseDate: string;
	RatingKey?: string;
	LibrarySection: LibrarySection;
}

export interface PosterFileSeason {
	Number: number;
}

export interface PosterFileEpisode {
	Title: string;
	EpisodeNumber: number;
	SeasonNumber: number;
}
