import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterSet } from "@/types/media-and-posters/poster-sets";

export interface DBPosterSetDetail {
	PosterSetID: string;
	PosterSet: PosterSet;
	LastDownloaded: string;
	SelectedTypes: string[];
	AutoDownload: boolean;
	ToDelete: boolean;
}

export interface DBMediaItemWithPosterSets {
	TMDB_ID: string;
	LibraryTitle: string;
	MediaItem: MediaItem;
	PosterSets: DBPosterSetDetail[];
}
