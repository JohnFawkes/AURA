import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

// Helper to remove special characters and lowercase the string
function normalizeString(str: string): string {
	return str.replace(/[^\w\s]/g, "").toLowerCase();
}

export function searchMediaItems(items: MediaItem[], query: string, limit?: number): MediaItem[] {
	let filteredItems = [...items];
	const trimmedQuery = query.trim();
	if (trimmedQuery === "") {
		return items.slice(0, limit);
	}

	// Extract year filter (e.g., Y:2023: or y:2023:)
	let yearFilter: number | null = null;
	const yearMatch = trimmedQuery.match(/[Yy]:(\d{4}):/);
	if (yearMatch) {
		yearFilter = parseInt(yearMatch[1], 10);
	}

	// Extract library filter (e.g., L:4K Movies:)
	let libraryFilter: string | null = null;
	// This regex captures everything between "L:" and the next colon.
	const libraryMatch = trimmedQuery.match(/[Ll]:(.+?):/);
	if (libraryMatch) {
		libraryFilter = libraryMatch[1].trim().toLowerCase();
	}

	// Extract rating key/tmdb id filter (e.g., ID:239:)
	let ratingFilter: string | null = null;
	const ratingMatch = trimmedQuery.match(/[Ii][Dd]:(.+?):/);
	if (ratingMatch) {
		ratingFilter = ratingMatch[1].trim();
	}

	// Remove year, library, and rating tokens from the query before further processing
	const partialQuery = trimmedQuery
		.replace(/[Yy]:(\d{4}):/, "")
		.replace(/[Ll]:(.+?):/, "")
		.replace(/[Ii][Dd]:(.+?):/, "")
		.trim();

	// If the query is wrapped in quotes, perform an exact match ignoring special characters.
	if (
		(partialQuery.startsWith('"') && partialQuery.endsWith('"')) ||
		(partialQuery.startsWith("'") && partialQuery.endsWith("'")) ||
		(partialQuery.startsWith("‘") && partialQuery.endsWith("’")) ||
		(partialQuery.startsWith("'“") && partialQuery.endsWith("”'"))
	) {
		const rawQuery = partialQuery.slice(1, partialQuery.length - 1);
		const normalizedQuery = normalizeString(rawQuery);
		filteredItems = filteredItems.filter((item) => normalizeString(item.Title) === normalizedQuery);
	} else if (partialQuery !== "") {
		// Normalize remaining query and split into words
		const normalizedQuery = normalizeString(partialQuery);
		const queryWords = normalizedQuery.split(/\s+/);
		filteredItems = filteredItems.filter((item) => {
			const normalizedTitle = normalizeString(item.Title);
			// Check that every query word exists in the title
			return queryWords.every((word) => normalizedTitle.includes(word));
		});
	}

	// Apply year filter if present
	if (yearFilter) {
		filteredItems = filteredItems.filter((item) => item.Year === yearFilter);
	}

	// Apply library filter if present
	if (libraryFilter) {
		filteredItems = filteredItems.filter((item) => normalizeString(item.LibraryTitle).includes(libraryFilter));
	}

	// Apply rating key filter if present (exact match)
	if (ratingFilter) {
		filteredItems = filteredItems.filter(
			(item) => item.RatingKey === ratingFilter || item.TMDB_ID === ratingFilter
		);
	}

	return filteredItems.slice(0, limit);
}

// Extracts year (Y:2023:) and mediaItemID (ID:123:) and Library (L:4K Movies:) from the search query
export function extractInfoFromSearchQuery(query: string): {
	searchTMDBID: string;
	searchLibrary: string;
	searchYear: number;
	searchTitle: string;
} {
	const trimmed = query.trim();

	// Match year: Y:2023: or y:2023:
	const yearMatch = trimmed.match(/[Yy]:(\d{4}):/);
	const year = yearMatch ? parseInt(yearMatch[1], 10) : 0;

	// Match mediaItemID: ID:123: or id:123:
	const idMatch = trimmed.match(/[Ii][Dd]:(.+?):/);
	const mediaItemID = idMatch ? idMatch[1] : "";

	// Match library: L:4K Movies: or l:4K Movies:
	const libraryMatch = trimmed.match(/[Ll]:(.+?):/);
	const library = libraryMatch ? libraryMatch[1].trim() : "";

	// Remove matched tokens from the original query
	let cleanedQuery = trimmed;
	if (yearMatch) {
		cleanedQuery = cleanedQuery.replace(yearMatch[0], "").trim();
	}
	if (idMatch) {
		cleanedQuery = cleanedQuery.replace(idMatch[0], "").trim();
	}
	if (libraryMatch) {
		cleanedQuery = cleanedQuery.replace(libraryMatch[0], "").trim();
	}

	// Return the cleaned search query along with extracted year and mediaItemIDs
	return { searchTMDBID: mediaItemID, searchLibrary: library, searchYear: year, searchTitle: cleanedQuery };
}
