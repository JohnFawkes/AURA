import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterSet } from "@/types/media-and-posters/poster-sets";

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
