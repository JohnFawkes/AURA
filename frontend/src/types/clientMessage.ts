import { MediaItem } from "./mediaItem";
import { PosterSet } from "./posterSets";

export interface ClientMessage {
	MediaItem: MediaItem;
	Set: PosterSet;
	SelectedTypes: string[];
	AutoDownload: boolean;
	LastUpdate?: string; // Optional field
}
