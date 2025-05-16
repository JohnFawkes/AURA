import { MediaItem } from "./mediaItem";
import { PosterSet } from "./posterSets";

export interface SavedSet {
	ID: string;
	MediaItem: MediaItem;
	Sets: Database_Set[];
}

export interface Database_Set {
	ID: string;
	Set: PosterSet;
	SelectedTypes: string[];
	AutoDownload: boolean;
	LastUpdate?: string; // Optional field
	ToDelete?: boolean; // Optional field
}
