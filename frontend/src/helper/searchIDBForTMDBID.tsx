import { librarySectionsStorage } from "@/lib/storage";

import { MediaItem } from "@/types/mediaItem";

export const getAllLibrarySectionsFromIDB = async (): Promise<{ title: string; type: string }[]> => {
	// Get all cached sections from librarySectionsStorage
	const keys = await librarySectionsStorage.keys();
	const cachedSectionsPromises = keys.map((key) =>
		librarySectionsStorage.getItem<{
			data: {
				Title: string;
				Type: string;
			};
		}>(key)
	);

	const sections = (await Promise.all(cachedSectionsPromises)).filter((section) => section !== null);

	if (sections.length === 0) {
		return [];
	}

	return sections.map((section) => ({
		title: section!.data.Title,
		type: section!.data.Type,
	}));
};

type Direction = "next" | "previous";
export const getAdjacentMediaItemFromIDB = async (
	currentRatingKey: string,
	direction: Direction
): Promise<MediaItem | null> => {
	// Get the sorted & filtered items from librarySectionsStorage
	const mediaItems = await librarySectionsStorage.getItem<MediaItem[]>("Home Page - Sorted and Filtered Items");

	if (!mediaItems || mediaItems.length === 0) {
		return null;
	}

	const currentIndex = mediaItems.findIndex((item) => item.RatingKey === currentRatingKey);

	if (currentIndex === -1) {
		return null; // Current item not found
	}

	let nextIndex: number;
	if (direction === "next") {
		nextIndex = currentIndex === mediaItems.length - 1 ? 0 : currentIndex + 1;
	} else {
		nextIndex = currentIndex === 0 ? mediaItems.length - 1 : currentIndex - 1;
	}

	return mediaItems[nextIndex];
};

export interface TMDBLookupMap {
	[tmdbId: string]: MediaItem;
}

// Utility function to create lookup map
export const createTMDBLookupMap = (mediaItems: MediaItem[]): TMDBLookupMap => {
	return mediaItems.reduce((map: TMDBLookupMap, item) => {
		const tmdbGuid = item.Guids?.find((g) => g.Provider === "tmdb");
		if (tmdbGuid?.ID) {
			map[tmdbGuid.ID] = item;
		}
		return map;
	}, {});
};

// Optimized search function using lookup map
export const searchWithLookupMap = (id: string, lookupMap: TMDBLookupMap): MediaItem | boolean => {
	return lookupMap[id] || false;
};
