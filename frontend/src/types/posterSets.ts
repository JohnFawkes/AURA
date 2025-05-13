export interface PosterSets {
	Type?: string;
	Item?: PosterItem;
	Sets?: PosterSet[];
}

export interface PosterItem {
	ID: string;
	Title: string;
	Status: string;
	Tagline: string;
	Slug: string;
	DateUpdated?: string;
	TvdbID?: string;
	ImdbID?: string;
	TraktID?: string;
	FirstAirDate?: string;
	ReleaseDate?: string;
}

export interface PosterSet {
	ID: string;
	User: {
		Name: string;
	};
	DateCreated: string;
	DateUpdated: string;
	Files: PosterFile[];
}

export interface PosterFile {
	ID: string;
	Type: string;
	Modified: string;
	FileSize: string;
	Movie?: PosterFileMovie;
	Season?: PosterFileSeason;
	Episode?: PosterFileEpisode;
}

export interface PosterFileMovie {
	ID: string;
}

export interface PosterFileSeason {
	Number?: number;
}

export interface PosterFileEpisode {
	Title?: string;
	EpisodeNumber?: number;
	SeasonNumber?: number;
}
