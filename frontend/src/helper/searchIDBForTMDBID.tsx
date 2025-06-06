import { CACHE_DB_NAME, CACHE_STORE_NAME } from "@/constants/cache";
import { MediaItem } from "@/types/mediaItem";
import { openDB } from "idb";

export const getAllLibrarySectionsFromIDB = async (): Promise<
	{ title: string; type: string }[]
> => {
	const db = await openDB(CACHE_DB_NAME, 1, {
		upgrade(db) {
			if (!db.objectStoreNames.contains(CACHE_STORE_NAME)) {
				db.createObjectStore(CACHE_STORE_NAME);
			}
		},
	});

	// Retrieve all library records
	const librarySections = await db.getAll(CACHE_STORE_NAME);
	if (!librarySections || librarySections.length === 0) {
		return [];
	}
	console.log("Library sections found:", librarySections);

	return librarySections.map((section) => {
		return {
			title: section.data.Title,
			type: section.data.Type,
		};
	});
};

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

export const searchIDBForTMDBIDNoLibrary = async (
	tmdbID: string
): Promise<MediaItem | boolean> => {
	const db = await openDB(CACHE_DB_NAME, 1, {
		upgrade(db) {
			if (!db.objectStoreNames.contains(CACHE_STORE_NAME)) {
				db.createObjectStore(CACHE_STORE_NAME);
			}
		},
	});

	// Retrieve all library records
	const librarySections = await db.getAll(CACHE_STORE_NAME);
	if (!librarySections || librarySections.length === 0) {
		return false;
	}

	// Find and update the media item with the matching ratingKey
	const mediaItems: MediaItem[] = [];
	librarySections.forEach((librarySection) => {
		if (
			librarySection &&
			librarySection.data &&
			librarySection.data.MediaItems
		) {
			const items: MediaItem[] = librarySection.data.MediaItems;
			const index = items.findIndex((item) =>
				item.Guids.some(
					(guid) => guid.ID === tmdbID && guid.Provider === "tmdb"
				)
			);
			if (index !== -1) {
				mediaItems.push(items[index]);
			}
		}
	});
	if (mediaItems.length === 0) {
		return false;
	}
	return mediaItems[0];
};
