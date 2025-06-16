import { storage } from "@/lib/storage";

import { MediaItem } from "@/types/mediaItem";

export const getAllLibrarySectionsFromIDB = async (): Promise<{ title: string; type: string }[]> => {
	// Get all cached sections from storage
	const keys = await storage.keys();
	const cachedSectionsPromises = keys.map((key) =>
		storage.getItem<{
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
	libraryTitle: string,
	currentRatingKey: string,
	direction: Direction
): Promise<MediaItem | null> => {
	const librarySection = await storage.getItem<{
		data: {
			MediaItems: MediaItem[];
		};
	}>(libraryTitle);

	if (!librarySection?.data?.MediaItems) {
		return null;
	}

	const mediaItems = librarySection.data.MediaItems;
	const currentIndex = mediaItems.findIndex((item) => item.RatingKey === currentRatingKey);

	if (currentIndex === -1) {
		return null; // Current item not found
	}

	let nextIndex: number;
	if (direction === "next") {
		// If at the end, wrap to beginning, otherwise move forward
		nextIndex = currentIndex === mediaItems.length - 1 ? 0 : currentIndex + 1;
	} else {
		// If at the beginning, wrap to end, otherwise move backward
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
export const searchWithLookupMap = (tmdbID: string, lookupMap: TMDBLookupMap): MediaItem | boolean => {
	return lookupMap[tmdbID] || false;
};
