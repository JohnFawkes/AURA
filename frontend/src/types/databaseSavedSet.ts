import { MediaItem } from "./mediaItem";
import { PosterSet } from "./posterSets";

export interface DBSavedItem {
	MediaItemID: string;
	MediaItem: MediaItem;
	PosterSetID: string;
	PosterSet: PosterSet;
	LastDownloaded: string;
	SelectedTypes: string[];
	AutoDownload: boolean;
}

export interface DBPosterSetDetail {
	PosterSetID: string;
	PosterSet: PosterSet;
	PosterSetJSON: string;
	LastDownloaded: string;
	SelectedTypes: string[];
	AutoDownload: boolean;
	ToDelete: boolean;
}

export interface DBMediaItemWithPosterSets {
	MediaItemID: string;
	MediaItem: MediaItem;
	MediaItemJSON: string;
	PosterSets: DBPosterSetDetail[];
}
