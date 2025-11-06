// --- Helper functions ---
function normalizeString(str: string): string {
	return str.replace(/[^\w\s]/g, "").toLowerCase();
}

function extractYearFromQuery(query: string): number | null {
	const yearMatch = query.match(/[Yy]:(\d{4}):/);
	return yearMatch ? parseInt(yearMatch[1], 10) : null;
}

function extractLibraryFromQuery(query: string): string | null {
	const libraryMatch = query.match(/[Ll]:(.+?):/);
	return libraryMatch ? libraryMatch[1].trim() : null;
}

function extractIDFromQuery(query: string): string | null {
	const idMatch = query.match(/[Ii][Dd]:(.+?):/);
	return idMatch ? idMatch[1].trim() : null;
}

function removeOtherFiltersFromQuery(query: string): string {
	return query
		.replace(/[Yy]:(\d{4}):/, "")
		.replace(/[Ll]:(.+?):/, "")
		.replace(/[Ii][Dd]:(.+?):/, "")
		.trim();
}

// --- Generic search function ---
type Extractors<T> = {
	getTitle: (item: T) => string;
	getLibraryTitle?: (item: T) => string | undefined;
	getYear?: (item: T) => number | undefined;
	getID?: (item: T) => string | undefined;
};

export function searchItems<T>(items: T[], query: string, extractors: Extractors<T>, limit?: number): T[] {
	let filteredItems = [...items];
	const trimmedQuery = query.trim();
	if (trimmedQuery === "") {
		return items.slice(0, limit);
	}

	const yearFilter = extractYearFromQuery(trimmedQuery);
	const libraryFilter = extractLibraryFromQuery(trimmedQuery);
	const idFilter = extractIDFromQuery(trimmedQuery);
	const partialQuery = removeOtherFiltersFromQuery(trimmedQuery);

	// Exact match if wrapped in quotes
	if (
		(partialQuery.startsWith('"') && partialQuery.endsWith('"')) ||
		(partialQuery.startsWith("'") && partialQuery.endsWith("'")) ||
		(partialQuery.startsWith("‘") && partialQuery.endsWith("’")) ||
		(partialQuery.startsWith("“") && partialQuery.endsWith("”"))
	) {
		const rawQuery = partialQuery.slice(1, partialQuery.length - 1);
		const normalizedQuery = normalizeString(rawQuery);
		filteredItems = filteredItems.filter((item) => normalizeString(extractors.getTitle(item)) === normalizedQuery);
	} else if (partialQuery !== "") {
		const normalizedQuery = normalizeString(partialQuery);
		const queryWords = normalizedQuery.split(/\s+/);
		filteredItems = filteredItems.filter((item) => {
			const normalizedTitle = normalizeString(extractors.getTitle(item));
			return queryWords.every((word) => normalizedTitle.includes(word));
		});
	}

	if (yearFilter && extractors.getYear) {
		filteredItems = filteredItems.filter((item) => extractors.getYear!(item) === yearFilter);
	}

	if (libraryFilter && extractors.getLibraryTitle) {
		const normalizedLibrary = normalizeString(libraryFilter);
		filteredItems = filteredItems.filter((item) =>
			normalizeString(extractors.getLibraryTitle!(item) || "").includes(normalizedLibrary)
		);
	}

	if (idFilter && extractors.getID) {
		filteredItems = filteredItems.filter((item) => extractors.getID!(item) === idFilter);
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
