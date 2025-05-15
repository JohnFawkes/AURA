import { MediaItem } from "@/types/mediaItem";

// Helper to remove special characters and lowercase the string
function normalizeString(str: string): string {
	return str.replace(/[^\w\s]/g, "").toLowerCase();
}

export function searchMediaItems(
	items: MediaItem[],
	query: string,
	limit: number = 10000
): MediaItem[] {
	let filteredItems = [...items];
	const trimmedQuery = query.trim();
	if (trimmedQuery === "") {
		return items.slice(0, limit);
	}

	let yearFilter: number | null = null;

	// Check for a year filter (e.g., y:2012)
	const yearMatch = trimmedQuery.match(/y:(\d{4})/);
	if (yearMatch) {
		yearFilter = parseInt(yearMatch[1], 10);
	}

	// If the query is wrapped in quotes, perform an exact match ignoring special characters.
	if (
		(trimmedQuery.startsWith('"') && trimmedQuery.endsWith('"')) ||
		(trimmedQuery.startsWith("'") && trimmedQuery.endsWith("'")) ||
		(trimmedQuery.startsWith("‘") && trimmedQuery.endsWith("’")) ||
		(trimmedQuery.startsWith("'“") && trimmedQuery.endsWith("”'"))
	) {
		const rawQuery = trimmedQuery.slice(1, trimmedQuery.length - 1);
		const normalizedQuery = normalizeString(rawQuery);
		filteredItems = filteredItems.filter(
			(item) => normalizeString(item.Title) === normalizedQuery
		);
	} else {
		// Remove year syntax from query and normalize it
		const partialQuery = trimmedQuery.replace(/y:\d{4}/, "").trim();
		const normalizedQuery = normalizeString(partialQuery);
		const queryWords = normalizedQuery.split(/\s+/);
		filteredItems = filteredItems.filter((item) => {
			const normalizedTitle = normalizeString(item.Title);
			// Check that every word in the query exists in the title
			return queryWords.every((word) => normalizedTitle.includes(word));
		});
	}

	if (yearFilter) {
		filteredItems = filteredItems.filter(
			(item) => item.Year === yearFilter
		);
	}

	return filteredItems.slice(0, limit);
}
