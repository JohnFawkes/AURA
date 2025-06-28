export interface LibrarySection {
	ID: string;
	Type: string; // "movie" or "show"
	Title: string;
	TotalSize: number;
	MediaItems: MediaItem[];
}

export interface MediaItem {
	RatingKey: string;
	Type: "show" | "movie";
	LibraryTitle: string;
	Title: string;
	Year: number;
	ExistInDatabase: boolean;
	Thumb?: string;
	ContentRating?: string;
	Summary?: string;
	UpdatedAt?: number;
	AddedAt?: number;
	ReleasedAt?: number;
	Guids: Guid[];
	Movie?: MediaItemMovie;
	Series?: MediaItemSeries;
}

export interface Guid {
	Provider?: string;
	ID?: string;
	Rating?: string;
}

export interface MediaItemMovie {
	File: MediaItemFile;
}

export interface MediaItemSeries {
	Seasons: MediaItemSeason[];
	SeasonCount: number;
	EpisodeCount: number;
}

export interface MediaItemSeason {
	RatingKey: string;
	SeasonNumber: number;
	Title: string;
	Episodes: MediaItemEpisode[];
}

export interface MediaItemEpisode {
	RatingKey: string;
	Title: string;
	SeasonNumber: number;
	EpisodeNumber: number;
	File: File;
}

export interface MediaItemFile {
	Path: string;
	Size: number;
	Duration: number;
}
