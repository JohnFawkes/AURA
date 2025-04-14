export interface LibrarySection {
	ID: string;
	Type: string; // "movie" or "show"
	Title: string;
	MediaItems: MediaItem[];
}

export interface MediaItem {
	RatingKey: string;
	LibraryTitle: string;
	Type: string; // "movie" or "show"
	Title: string;
	Year: number;
	Thumb: string;
	AudienceRating: number;
	UserRating: number;
	ContentRating: string;
	Summary: string;
	UpdatedAt: number;
	Guids: Guid[];
	Movie?: Movie;
	Series?: Series;
}

export interface Guid {
	Provider: string;
	ID: string;
}

export interface Movie {
	File: File;
}

export interface Series {
	Seasons: Season[];
	SeasonCount: number;
	EpisodeCount: number;
}

export interface Season {
	RatingKey: string;
	SeasonNumber: number;
	Title: string;
	Episodes: Episode[];
}

export interface Episode {
	RatingKey: string;
	Title: string;
	SeasonNumber: number;
	EpisodeNumber: number;
	File: File;
}

export interface File {
	Path: string;
	Size: number;
	Duration: number;
}
