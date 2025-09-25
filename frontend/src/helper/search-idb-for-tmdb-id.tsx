import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export interface TMDBLookupMap {
	[tmdbId: string]: MediaItem;
}

export const createTMDBLookupMap = (mediaItems: MediaItem[]): TMDBLookupMap =>
	mediaItems.reduce((map: TMDBLookupMap, item) => {
		const tmdbGuid = item.Guids?.find((g) => g.Provider === "tmdb");
		if (tmdbGuid?.ID) map[tmdbGuid.ID] = item;
		return map;
	}, {});

export const searchWithLookupMap = (id: string, lookupMap: TMDBLookupMap): MediaItem | boolean =>
	lookupMap[id] || false;
