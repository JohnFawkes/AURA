import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterSet } from "@/types/media-and-posters/poster-sets";

export interface DBSavedItem {
	MediaItemID: string;
	MediaItem: MediaItem;
	PosterSetID: string;
	PosterSet: PosterSet;
	LastDownloaded: string;
	SelectedTypes: string[];
	AutoDownload: boolean;
}
