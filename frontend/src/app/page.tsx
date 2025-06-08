"use client";
import { LibrarySection, MediaItem } from "@/types/mediaItem";
import { useEffect, useState, useCallback, useRef } from "react";
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
	fetchMediaServerLibrarySectionItems,
	fetchMediaServerLibrarySections,
} from "@/services/api.mediaserver";
import { log } from "@/lib/logger";
import { Progress } from "@/components/ui/progress";
import { useHomeSearchStore } from "@/lib/homeSearchStore";
import { searchMediaItems } from "@/hooks/searchMediaItems";
import localforage from "localforage";

export const CACHE_DURATION = 24 * 60 * 60 * 1000;
// Initialize localforage
localforage.config({
	name: "aura",
	storeName: "LibrarySections",
	version: 1.0,
	description: "Library sections cache for Aura",
});

export default function Home() {
	const isMounted = useRef(false);
	if (typeof window !== "undefined") {
		// Safe to use document here.
		document.title = "Aura | Home";
	}
	// -------------------------------
	// States
	// -------------------------------
	// Search
	const { searchQuery } = useHomeSearchStore();
	const prevSearchQuery = useRef(searchQuery);

	// Loading & Error
	const [errorMessage, setErrorMessage] = useState<string>("");
	const [fullyLoaded, setFullyLoaded] = useState<boolean>(false);

	// Library sections & progress
	const [librarySections, setLibrarySections] = useState<LibrarySection[]>(
		[]
	);
	const [sectionProgress, setSectionProgress] = useState<{
		[key: string]: { loaded: number; total: number };
	}>({});

	// Filtering & Pagination
	const {
		filteredLibraries,
		setFilteredLibraries,
		filterOutInDB,
		setFilterOutInDB,
	} = useHomeSearchStore();
	const [filteredItems, setFilteredItems] = useState<MediaItem[]>([]);
	const { currentPage, setCurrentPage } = useHomeSearchStore();
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
	const getMediaItems = useCallback(async (useCache: boolean) => {
		if (isMounted.current) return;
		// Reset progress state before starting a new fetch
		setSectionProgress({});
		setFullyLoaded(false);
		isMounted.current = true;
		try {
			let sections: LibrarySection[] = [];

			// If cache is allowed, try loading from localforage
			if (useCache) {
				log("Home Page - Attempting to load sections from cache");
				// Get all cached sections
				const cachedSections: {
					data: LibrarySection;
					timestamp: number;
				}[] = (
					await localforage.keys().then((keys) =>
						Promise.all(
							keys.map((key) =>
								localforage.getItem<{
									data: LibrarySection;
									timestamp: number;
								}>(key)
							)
						)
					)
				).filter(
					(
						section
					): section is { data: LibrarySection; timestamp: number } =>
						section !== null
				);

				if (cachedSections && cachedSections.length > 0) {
					// Filter valid cached sections
					const validSections = cachedSections.filter(
						(section) =>
							Date.now() - section.timestamp < CACHE_DURATION
					);

					if (validSections.length > 0) {
						sections = validSections.map((s) => s.data);
						setLibrarySections(sections);
						setFullyLoaded(true);
						log("Home Page - Using cached sections", validSections);
						return;
					}
				}

				// Clear invalid cache
				if (sections.length === 0) {
					await localforage.clear();
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

			// Process each section concurrently
			await Promise.all(
				sections.map(async (section, idx) => {
					let itemsFetched = 0;
					let totalSize = Infinity;
					let allItems: LibrarySection["MediaItems"] = [];

					while (itemsFetched < totalSize) {
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
						const items = data?.MediaItems || [];
						allItems = allItems.concat(items);
						if (totalSize === Infinity) {
							totalSize = data?.TotalSize ?? 0;
						}
						itemsFetched += items.length;
						// Update the progress state for this section:
						setSectionProgress((prev) => ({
							...prev,
							[section.ID]: {
								loaded: itemsFetched,
								total: totalSize,
							},
						}));
						if (items.length === 0) {
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

					// Cache using localforage
					await localforage.setItem(`${section.Title}`, {
						data: section,
						timestamp: Date.now(),
					});
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
			isMounted.current = false;
		}
	}, []);

	useEffect(() => {
		getMediaItems(true);
	}, [getMediaItems]);

	useEffect(() => {
		if (searchQuery !== prevSearchQuery.current) {
			setCurrentPage(1);
			prevSearchQuery.current = searchQuery;
		}
	}, [searchQuery, setCurrentPage]);

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

		// Filter out items already in the DB
		if (filterOutInDB) {
			items = items.filter((item) => !item.ExistInDatabase);
		}

		// Filter out items by search
		setFilteredItems(searchMediaItems(items, searchQuery));
	}, [librarySections, filteredLibraries, searchQuery, filterOutInDB]);

	if (errorMessage) {
		return <ErrorMessage message={errorMessage} />;
	}

	return (
		<div className="min-h-screen px-8 pb-20 sm:px-20">
			{!fullyLoaded && librarySections.length > 0 && (
				<div className="mb-4">
					{librarySections.map((section) => {
						// Retrieve progress info for this section
						const progressInfo = sectionProgress[section.ID];
						const percentage =
							progressInfo && progressInfo.total > 0
								? Math.min(
										(progressInfo.loaded /
											progressInfo.total) *
											100,
										100
								  )
								: 0;

						// Render progress UI only if the percentage is not 100
						if (Math.round(percentage) !== 100) {
							return (
								<div key={section.ID} className="mb-2">
									<Label className="text-lg font-semibold">
										Loading {section.Title}
									</Label>
									<Progress
										value={percentage}
										className="mt-1"
									/>
									<span className="ml-2 text-sm text-muted-foreground">
										{Math.round(percentage)}%
									</span>
								</div>
							);
						}
					})}
				</div>
			)}

			{/* Filter Section*/}
			<div className="flex flex-col sm:flex-row mb-4 mt-2">
				{/* Label */}
				<Label
					htmlFor="library-filter"
					className="text-lg font-semibold mb-2 sm:mb-0 sm:mr-4"
				>
					Filters:
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
									setFilteredLibraries(
										filteredLibraries.filter(
											(lib: string) =>
												lib !== section.Title
										)
									);
									setCurrentPage(1);
								} else {
									setFilteredLibraries([
										...filteredLibraries,
										section.Title,
									]);
									setCurrentPage(1);
								}
							}}
						>
							{section.Title}
						</Badge>
					))}

					<Badge
						key={"filter-out-in-db"}
						className="cursor-pointer"
						variant={filterOutInDB ? "default" : "outline"}
						onClick={() => {
							setFilterOutInDB(!filterOutInDB);
							setCurrentPage(1);
						}}
					>
						{filterOutInDB ? "Items Not in DB" : "All Items"}
					</Badge>
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
									onClick={() => {
										const newPage = Math.max(
											currentPage - 1,
											1
										);
										setCurrentPage(newPage);
										window.scrollTo({
											top: 0,
											behavior: "smooth",
										});
									}}
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
									onClick={() => {
										const newPage = Math.min(
											currentPage + 1,
											totalPages
										);
										setCurrentPage(newPage);
										window.scrollTo({
											top: 0,
											behavior: "smooth",
										});
									}}
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
										onClick={() => {
											setCurrentPage(totalPages);
											window.scrollTo({
												top: 0,
												behavior: "smooth",
											});
										}}
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
