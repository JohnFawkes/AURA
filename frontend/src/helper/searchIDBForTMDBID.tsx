import { CACHE_DB_NAME, CACHE_STORE_NAME } from "@/constants/cache";
import { MediaItem } from "@/types/mediaItem";
import { openDB } from "idb";

export const searchIDBForTMDBID = async (
	tmdbID: string,
	sectionTitle: string
): Promise<MediaItem | boolean> => {
	const db = await openDB(CACHE_DB_NAME, 1, {
		upgrade(db) {
			if (!db.objectStoreNames.contains(CACHE_STORE_NAME)) {
				db.createObjectStore(CACHE_STORE_NAME);
			}
		},
	});

	// Retrieve the library record
	const librarySection = await db.get(CACHE_STORE_NAME, sectionTitle);
	if (
		!librarySection ||
		!librarySection.data ||
		!librarySection.data.MediaItems
	) {
		return false;
	}

	// Find and update the media item with the matching ratingKey
	const mediaItems: MediaItem[] = librarySection.data.MediaItems;
	const index = mediaItems.findIndex((item) =>
		item.Guids.some(
			(guid) => guid.ID === tmdbID && guid.Provider === "tmdb"
		)
	);
	if (index === -1) {
		return false;
	}
	return mediaItems[index];
};
