import { storage } from "@/lib/storage";

import { MediaItem } from "@/types/mediaItem";

export const getAllLibrarySectionsFromIDB = async (): Promise<
	{ title: string; type: string }[]
> => {
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

	const sections = (await Promise.all(cachedSectionsPromises)).filter(
		(section) => section !== null
	);

	if (sections.length === 0) {
		return [];
	}

	return sections.map((section) => ({
		title: section!.data.Title,
		type: section!.data.Type,
	}));
};

export const searchIDBForTMDBID = async (
	tmdbID: string,
	sectionTitle: string
): Promise<MediaItem | boolean> => {
	// Get section from storage
	const librarySection = await storage.getItem<{
		data: {
			MediaItems: MediaItem[];
		};
	}>(sectionTitle);

	if (!librarySection?.data?.MediaItems) {
		return false;
	}

	// Find media item with matching TMDB ID
	const mediaItem = librarySection.data.MediaItems.find((item) =>
		item.Guids.some((guid) => guid.ID === tmdbID && guid.Provider === "tmdb")
	);

	return mediaItem || false;
};

export const searchIDBForTMDBIDNoLibrary = async (tmdbID: string): Promise<MediaItem | boolean> => {
	// Get all cached sections
	const keys = await storage.keys();
	const cachedSectionsPromises = keys.map((key) =>
		storage.getItem<{
			data: {
				MediaItems: MediaItem[];
			};
		}>(key)
	);

	const sections = (await Promise.all(cachedSectionsPromises)).filter(
		(section) => section !== null
	);

	if (sections.length === 0) {
		return false;
	}

	// Search through all sections for matching TMDB ID
	for (const section of sections) {
		if (section?.data?.MediaItems) {
			const mediaItem = section.data.MediaItems.find((item) =>
				item.Guids.some((guid) => guid.ID === tmdbID && guid.Provider === "tmdb")
			);
			if (mediaItem) {
				return mediaItem;
			}
		}
	}

	return false;
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
