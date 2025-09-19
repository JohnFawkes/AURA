import { useHomePageStore } from "@/lib/stores/page-store-home";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

type Direction = "next" | "previous";

/**
 * Retrieves adjacent media item (wrap-around) from the Home page store's
 * filteredAndSortedMediaItems array.
 */
export const getAdjacentMediaItem = (currentRatingKey: string, direction: Direction): MediaItem | null => {
	const mediaItems = useHomePageStore.getState().filteredAndSortedMediaItems || [];
	if (!mediaItems.length) return null;

	const currentIndex = mediaItems.findIndex((m) => m.RatingKey === currentRatingKey);
	if (currentIndex === -1) return null;

	const nextIndex =
		direction === "next"
			? (currentIndex + 1) % mediaItems.length
			: (currentIndex - 1 + mediaItems.length) % mediaItems.length;

	return mediaItems[nextIndex] ?? null;
};

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
