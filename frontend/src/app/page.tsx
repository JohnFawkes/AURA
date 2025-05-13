"use client";
import { LibrarySection, MediaItem } from "@/types/mediaItem";
import { openDB } from "idb";
import { useEffect, useState, useContext, useCallback } from "react";
import { SearchContext } from "@/app/layout";
import ErrorMessage from "@/components/ui/error-message";
import HomeMediaItemCard from "@/components/ui/home-media-item-card";
import {
	Pagination,
	PaginationContent,
	PaginationEllipsis,
	PaginationItem,
	PaginationLink,
	PaginationNext,
	PaginationPrevious,
} from "@/components/ui/pagination";
import { Button } from "@/components/ui/button";
import { RefreshCcw as RefreshIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { ToggleGroup } from "@/components/ui/toggle-group";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import {
	CACHE_DB_NAME,
	CACHE_STORE_NAME,
	CACHE_EXPIRY,
} from "@/constants/cache";
import {
	fetchMediaServerLibrarySectionItems,
	fetchMediaServerLibrarySections,
} from "@/services/api.mediaserver";
import { log } from "@/lib/logger";
import { Progress } from "@/components/ui/progress";

export default function Home() {
	const [isMounted, setIsMounted] = useState(false);

	// -------------------------------
	// Contexts
	// -------------------------------
	const { searchQuery } = useContext(SearchContext);

	// -------------------------------
	// States
	// -------------------------------
	// Search
	const [debouncedQuery, setDebouncedQuery] = useState<string>("");

	// Loading & Error
	const [errorMessage, setErrorMessage] = useState<string>("");
	const [fullyLoaded, setFullyLoaded] = useState<boolean>(false);

	// Library sections & progress
	const [librarySections, setLibrarySections] = useState<LibrarySection[]>(
		[]
	);
	const [loadingLibraryName, setLoadingLibraryName] = useState<string>("");
	const [loadingLibraryProgress, setLoadingProgress] = useState<number>(0);
	const [loadingLibraryTotalSize, setLoadingTotalSize] = useState<number>(0);

	// Filtering & Pagination
	const [filteredLibraries, setFilteredLibraries] = useState<string[]>([]);
	const [filteredItems, setFilteredItems] = useState<MediaItem[]>([]);
	const [currentPage, setCurrentPage] = useState<number>(1);
	const itemsPerPage = 20;

	// -------------------------------
	// Derived values
	// -------------------------------
	const paginatedItems = filteredItems.slice(
		(currentPage - 1) * itemsPerPage,
		currentPage * itemsPerPage
	);
	const totalPages = Math.ceil(filteredItems.length / itemsPerPage);

	// Fetch data from cache or API
	const getMediaItems = useCallback(
		async (useCache: boolean) => {
			if (isMounted) return;
			setIsMounted(true);
			try {
				const db = await openDB(CACHE_DB_NAME, 1, {
					upgrade(db) {
						if (!db.objectStoreNames.contains(CACHE_STORE_NAME)) {
							db.createObjectStore(CACHE_STORE_NAME);
						}
					},
				});

				let sections: LibrarySection[] = [];

				// If cache is allowed, try loading all sections from DB (using title as key)
				if (useCache) {
					// Get all cached sections
					const cachedSections = await db.getAll(CACHE_STORE_NAME);
					if (cachedSections.length > 0) {
						// Filter valid cached sections
						const validSections = cachedSections.filter(
							(section) =>
								Date.now() - section.timestamp < CACHE_EXPIRY
						);
						if (validSections.length > 0) {
							sections = validSections.map((s) => s.data);
							setLibrarySections(sections);
							setFullyLoaded(true);
							log(
								"Home Page - Using cached sections",
								validSections
							);
							return;
						}
					}
					// If no valid cached sections, clear the store.
					if (sections.length === 0) {
						await db.clear(CACHE_STORE_NAME);
					}
				}

				setFullyLoaded(false);

				// If sections were not loaded from cache, fetch them from the API.
				if (sections.length === 0) {
					const sectionsResponse =
						await fetchMediaServerLibrarySections();
					if (sectionsResponse.status !== "success") {
						throw new Error(sectionsResponse.message);
					}
					sections = sectionsResponse.data || [];
					if (!sections || sections.length === 0) {
						throw new Error(
							"No sections found, please check the logs."
						);
					}
					// Initialize media items for each section.
					sections.forEach((section) => {
						section.MediaItems = [];
					});
					setLibrarySections(sections);
				}

				await Promise.all(
					sections.map(async (section, idx) => {
						let itemsFetched = 0;
						let totalSize = Infinity;
						let allItems: LibrarySection["MediaItems"] = [];
						while (itemsFetched < totalSize) {
							setLoadingLibraryName(section.Title);
							const itemsResponse =
								await fetchMediaServerLibrarySectionItems(
									section,
									itemsFetched
								);
							if (itemsResponse.status !== "success") {
								console.error(itemsResponse.message);
								break;
							}
							const data = itemsResponse.data;
							allItems = allItems.concat(data?.MediaItems || []);
							if (totalSize === Infinity) {
								totalSize = data?.TotalSize ?? 0;
								setLoadingTotalSize(totalSize);
							}
							itemsFetched += data?.MediaItems?.length || 0;
							setLoadingProgress(itemsFetched);
							if ((data?.MediaItems?.length || 0) === 0) {
								break;
							}
						}
						// Update section with fetched media items.
						section.MediaItems = allItems;
						setLibrarySections((prev) => {
							const updated = [...prev];
							updated[idx] = section;
							return updated;
						});
						// Cache the complete section (using title as key).
						const db = await openDB(CACHE_DB_NAME, 1);
						await db.put(
							CACHE_STORE_NAME,
							{
								data: section,
								timestamp: Date.now(),
							},
							section.Title
						);
					})
				);

				log("Home Page - Sections fetched successfully", sections);
				setFullyLoaded(true);
			} catch (error) {
				setErrorMessage(
					error instanceof Error
						? error.message
						: "An unknown error occurred"
				);
			} finally {
				setIsMounted(false);
			}
		},
		[isMounted]
	);

	useEffect(() => {
		getMediaItems(true);
	}, [getMediaItems]);

	// Debounce the search query
	useEffect(() => {
		const handler = setTimeout(() => {
			setDebouncedQuery(searchQuery);
		}, 300);

		return () => clearTimeout(handler);
	}, [searchQuery]);

	// Filter items based on the search query
	useEffect(() => {
		let items = librarySections.flatMap(
			(section) => section.MediaItems || []
		);

		// Filter by selected libraries
		if (filteredLibraries.length > 0) {
			items = items.filter((item) =>
				filteredLibraries.includes(item.LibraryTitle)
			);
		}

		// Handle search query
		if (debouncedQuery.trim() !== "") {
			const query = debouncedQuery.trim();
			let exactQuery = "";
			let yearFilter = null;

			// Check if the query contains a year filter (e.g., y:2012)
			const yearMatch = query.match(/y:(\d{4})/);
			if (yearMatch) {
				yearFilter = parseInt(yearMatch[1], 10);
			}

			// Check if the query is wrapped in quotes for exact match
			if (
				(query.startsWith('"') && query.endsWith('"')) ||
				(query.startsWith("'") && query.endsWith("'")) ||
				(query.startsWith("‘") && query.endsWith("’")) ||
				(query.startsWith("'“") && query.endsWith("”'"))
			) {
				// Normalize quotes
				const normalizedQuery = query
					.replace(/‘|’/g, "'")
					.replace(/“|”/g, '"')
					.replace(/'/g, '"');

				exactQuery = normalizedQuery
					.slice(1, normalizedQuery.length - 1) // Remove surrounding quotes
					.toLowerCase();
				items = items.filter(
					(item) => item.Title.toLowerCase() === exactQuery
				);
			} else {
				// Partial match (search for the query in the title)
				const partialQuery = query
					.replace(/y:\d{4}/, "")
					.trim()
					.toLowerCase();
				items = items.filter((item) =>
					item.Title.toLowerCase().includes(partialQuery)
				);
			}

			// Apply year filter if present
			if (yearFilter) {
				items = items.filter((item) => item.Year === yearFilter);
			}
		}

		setFilteredItems(items);
		setCurrentPage(1); // Reset to the first page on new search
	}, [librarySections, filteredLibraries, debouncedQuery]);

	if (errorMessage) {
		return <ErrorMessage message={errorMessage} />;
	}

	return (
		<div className="min-h-screen px-8 pb-20 sm:px-20">
			{!fullyLoaded && (
				<div className="w-full mt-2">
					<div className="flex items-center justify-between">
						<Label
							htmlFor="library-filter"
							className="text-lg font-semibold"
						>
							Loading {loadingLibraryName || "Media Items"}
						</Label>
						<Progress
							value={
								(loadingLibraryProgress /
									loadingLibraryTotalSize) *
									100 || 0
							}
							className="flex-1 ml-2"
						/>
						<span className="ml-2 text-sm text-muted-foreground">
							{Math.round(
								(loadingLibraryProgress /
									loadingLibraryTotalSize) *
									100 || 0
							)}
							%
						</span>
					</div>
				</div>
			)}

			{/* Filter and Sort Section */}
			<div className="flex flex-col sm:flex-row mb-4 mt-2">
				{/* Label */}
				<Label
					htmlFor="library-filter"
					className="text-lg font-semibold mb-2 sm:mb-0 sm:mr-4"
				>
					Filter by Library Name:
				</Label>

				{/* ToggleGroup */}
				<ToggleGroup
					type="multiple"
					className="flex flex-wrap sm:flex-nowrap gap-2"
					value={filteredLibraries}
					onValueChange={setFilteredLibraries}
				>
					{librarySections.map((section) => (
						<Badge
							key={section.ID}
							className="cursor-pointer"
							variant={
								filteredLibraries.includes(section.Title)
									? "default"
									: "outline"
							}
							onClick={() => {
								if (filteredLibraries.includes(section.Title)) {
									setFilteredLibraries((prev) =>
										prev.filter(
											(lib) => lib !== section.Title
										)
									);
								} else {
									setFilteredLibraries((prev) => [
										...prev,
										section.Title,
									]);
								}
							}}
						>
							{section.Title}
						</Badge>
					))}
				</ToggleGroup>
			</div>

			{/* Grid of Cards */}
			<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-1">
				{paginatedItems.length === 0 && searchQuery && fullyLoaded ? (
					<div className="col-span-full text-center text-red-500">
						No items found matching '{searchQuery}' in{" "}
						{filteredLibraries.length > 0
							? filteredLibraries.join(", ")
							: "any library"}
					</div>
				) : (
					paginatedItems.map((item) => (
						<HomeMediaItemCard
							key={item.RatingKey}
							mediaItem={item}
						/>
					))
				)}
			</div>

			{/* Pagination */}
			<div className="flex justify-center mt-8">
				<Pagination>
					<PaginationContent>
						{/* Previous Page Button */}
						{totalPages > 1 && (
							<PaginationItem>
								<PaginationPrevious
									onClick={() =>
										setCurrentPage((prev) =>
											Math.max(prev - 1, 1)
										)
									}
								/>
							</PaginationItem>
						)}

						{/* Current Page */}
						<PaginationItem>
							<PaginationLink isActive>
								{currentPage}
							</PaginationLink>
						</PaginationItem>

						{/* Next Page Button */}
						{totalPages > 1 && currentPage < totalPages && (
							<PaginationItem>
								<PaginationNext
									onClick={() =>
										setCurrentPage((prev) =>
											Math.min(prev + 1, totalPages)
										)
									}
								/>
							</PaginationItem>
						)}

						{/* Ellipsis and End Page */}
						{totalPages > 3 && currentPage < totalPages - 1 && (
							<>
								<PaginationItem>
									<PaginationEllipsis />
								</PaginationItem>
								<PaginationItem>
									<PaginationLink
										onClick={() =>
											setCurrentPage(totalPages)
										}
									>
										{totalPages}
									</PaginationLink>
								</PaginationItem>
							</>
						)}
					</PaginationContent>
				</Pagination>
			</div>

			<Button
				variant="outline"
				size="sm"
				className={cn(
					"fixed z-100 right-3 bottom-10 sm:bottom-15 rounded-full shadow-lg transition-all duration-300 bg-background border-primary-dynamic text-primary-dynamic hover:bg-primary-dynamic hover:text-primary cursor-pointer"
				)}
				onClick={() => getMediaItems(false)}
				aria-label="refresh"
			>
				<RefreshIcon className="h-3 w-3 mr-1" />
				<span className="text-xs hidden sm:inline">Refresh</span>
			</Button>
		</div>
	);
}
