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
	Files: File[];
}

export interface File {
	ID: string;
	Type: string;
	Modified: string;
	Movie?: Movie;
	Season?: Season;
	Episode?: Episode;
}

export interface Movie {
	ID: string;
}

export interface Season {
	Number?: number;
}

export interface Episode {
	Title?: string;
	Number?: number;
	Season?: Season;
}
