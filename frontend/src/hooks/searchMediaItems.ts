import { MediaItem } from "@/types/mediaItem";

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

	// Extract rating key filter (e.g., ID:239:)
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
		filteredItems = filteredItems.filter((item) => item.RatingKey === ratingFilter);
	}

	return filteredItems.slice(0, limit);
}
