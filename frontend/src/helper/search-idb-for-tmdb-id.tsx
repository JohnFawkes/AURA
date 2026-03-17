import type { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export interface TMDBLookupMap {
  [tmdbId: string]: MediaItem;
}

export const createTMDBLookupMap = (mediaItems: MediaItem[]): TMDBLookupMap =>
  mediaItems.reduce((map: TMDBLookupMap, item) => {
    if (item.tmdb_id) map[item.tmdb_id] = item;
    return map;
  }, {});

export const searchWithLookupMap = (id: string, lookupMap: TMDBLookupMap): MediaItem | boolean =>
  lookupMap[id] || false;
