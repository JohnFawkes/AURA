"use client";

import { ReturnErrorMessage } from "@/services/api-error-return";
import { fetchMediaServerLibrarySectionItems } from "@/services/mediaserver/api-mediaserver-fetch-library-section-items";
import { fetchMediaServerLibrarySections } from "@/services/mediaserver/api-mediaserver-fetch-library-sections";
import { Loader } from "lucide-react";

import { useCallback, useEffect, useRef, useState } from "react";

import { CustomPagination } from "@/components/shared/custom-pagination";
import { ErrorMessage } from "@/components/shared/error-message";
import { FilterHome } from "@/components/shared/filter-home";
import HomeMediaItemCard from "@/components/shared/media-item-card";
import { HomeMediaItemCardSkeletonGrid } from "@/components/shared/media-item-card-skeleton";
import { RefreshButton } from "@/components/shared/refresh-button";
import { ResponsiveGrid } from "@/components/shared/responsive-grid";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";

import { cn } from "@/lib/cn";
import { log } from "@/lib/logger";
import { MAX_CACHE_DURATION, useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";
import { useHomePageStore } from "@/lib/stores/page-store-home";

import { searchItems } from "@/hooks/search-query";

import { APIResponse } from "@/types/api/api-response";
import { LibrarySection } from "@/types/media-and-posters/media-item-and-library";

export default function Home() {
	useEffect(() => {
		document.title = "aura | Home";
	}, []);
	const isMounted = useRef(false);

	// -------------------------------
	// States
	// -------------------------------
	// Search
	const { searchQuery } = useSearchQueryStore();
	const prevSearchQuery = useRef(searchQuery);

	// Loading & Error
	const [error, setError] = useState<APIResponse<unknown> | null>(null);
	const [fullyLoaded, setFullyLoaded] = useState<boolean>(false);

	// Library sections & progress
	const [librarySections, setLibrarySections] = useState<LibrarySection[]>([]);
	const [sectionProgress, setSectionProgress] = useState<{
		[key: string]: { loaded: number; total: number };
	}>({});

	// State to track the HomePageStore values
	const {
		filteredLibraries,
		setFilteredLibraries,
		filterInDB,
		setFilterInDB,
		currentPage,
		setCurrentPage,
		itemsPerPage,
		setItemsPerPage,
		sortOption,
		setSortOption,
		sortOrder,
		setSortOrder,
		filteredAndSortedMediaItems,
		setFilteredAndSortedMediaItems,
	} = useHomePageStore();

	const { sections, setSections, timestamp } = useLibrarySectionsStore();
	const hasHydrated = useLibrarySectionsStore((state) => state.hasHydrated);

	// -------------------------------
	// Derived values
	// -------------------------------
	const paginatedItems = filteredAndSortedMediaItems.slice(
		(currentPage - 1) * itemsPerPage,
		currentPage * itemsPerPage
	);
	const totalPages = Math.ceil(filteredAndSortedMediaItems.length / itemsPerPage);

	// Set sortOption to "dateAdded" if its not title or dateUpdated or dateAdded or dateReleased
	useEffect(() => {
		if (
			sortOption !== "title" &&
			sortOption !== "dateUpdated" &&
			sortOption !== "dateAdded" &&
			sortOption !== "dateReleased"
		) {
			setSortOption("dateAdded");
			setSortOrder("desc");
		}
	}, [sortOption, setSortOption, setSortOrder]);

	// Fetch data from cache or API
	const getMediaItems = useCallback(
		async (useCache: boolean) => {
			if (isMounted.current && useCache) return;
			setSectionProgress({});
			setLibrarySections([]);
			setError(null);
			setFullyLoaded(false);
			try {
				// Check if we want to use cache
				if (useCache) {
					const isCacheAgeValid = timestamp ? Date.now() - timestamp < MAX_CACHE_DURATION : false;
					const cacheContainsSectionsAndTimestamp = sections && timestamp && Object.keys(sections).length > 0;
					log("INFO", "Home Page", "Library Cache", "Attempting to load sections from cache", {
						"Current Time": Date.now(),
						"Cache Timestamp": timestamp,
						"Cache Age Max (ms)": MAX_CACHE_DURATION,
						"Cache Age (ms)": timestamp ? Date.now() - timestamp : "N/A",
						"Is Cache Age Valid": isCacheAgeValid,
						"Cache Contains Sections & Timestamp": cacheContainsSectionsAndTimestamp,
					});
					if (cacheContainsSectionsAndTimestamp) {
						if (isCacheAgeValid) {
							setLibrarySections(Object.values(sections));
							setFullyLoaded(true);
							log("INFO", "Home Page", "Library Cache", "Using cached sections", sections);
							return;
						} else {
							log("WARN", "Home Page", "Library Cache", "Cache expired, fetching fresh data");
						}
					} else {
						log("WARN", "Home Page", "Library Cache", "No valid cache found, fetching fresh data");
					}
				}

				// Fetch fresh data
				const response = await fetchMediaServerLibrarySections();
				if (response.status === "error") {
					setError(response);
					setFullyLoaded(true);
					return;
				}

				const fetchedSections = response.data || [];
				if (!fetchedSections || fetchedSections.length === 0) {
					setError(ReturnErrorMessage<unknown>(new Error("No sections found, please check the logs.")));
					return;
				}

				// Initialize each section's MediaItems to an empty array
				fetchedSections.forEach((section) => (section.MediaItems = []));
				setLibrarySections(fetchedSections.slice().sort((a, b) => a.Title.localeCompare(b.Title)));

				// Process each section concurrently
				await Promise.all(
					fetchedSections.map(async (section) => {
						let itemsFetched = 0;
						let totalSize = Infinity;
						let allItems: LibrarySection["MediaItems"] = [];

						while (itemsFetched < totalSize) {
							const itemsResponse = await fetchMediaServerLibrarySectionItems(section, itemsFetched);
							if (itemsResponse.status === "error") {
								setError(itemsResponse);
								return;
							}

							const data = itemsResponse.data;
							const items = data?.MediaItems || [];
							allItems = allItems.concat(items);
							if (totalSize === Infinity) {
								totalSize = data?.TotalSize ?? 0;
							}
							itemsFetched += items.length;
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
						section.MediaItems = allItems;
						section.TotalSize = totalSize;
					})
				);

				// Build the sections object for the store
				const sectionsObj = fetchedSections.reduce<Record<string, LibrarySection>>((acc, section) => {
					acc[section.Title] = section;
					return acc;
				}, {});
				const librarySections = fetchedSections.slice().sort((a, b) => a.Title.localeCompare(b.Title));
				// Store in zustand and update timestamp
				setSections(sectionsObj, Date.now());
				setFullyLoaded(true);
				log("INFO", "Home Page", "", "Sections fetched successfully from server", {
					"Library Sections": librarySections,
					Sections: sectionsObj,
				});
				setLibrarySections(librarySections);
			} catch (error) {
				setError(ReturnErrorMessage<unknown>(error));
			} finally {
				isMounted.current = false;
			}
		},
		[sections, setSections, timestamp]
	);

	useEffect(() => {
		if (!hasHydrated) return;
		getMediaItems(true);
		isMounted.current = true;
	}, [getMediaItems, hasHydrated]);

	useEffect(() => {
		if (searchQuery !== prevSearchQuery.current) {
			setCurrentPage(1);
			prevSearchQuery.current = searchQuery;
		}
	}, [searchQuery, setCurrentPage]);

	// Filter items based on the search query
	useEffect(() => {
		const filterAndSortItems = async () => {
			let items = librarySections.flatMap((section) => section.MediaItems || []);

			// Sort items by Title
			if (sortOption === "title") {
				if (sortOrder === "asc") {
					items.sort((a, b) => a.Title.localeCompare(b.Title));
				} else if (sortOrder === "desc") {
					items.sort((a, b) => b.Title.localeCompare(a.Title));
				}
			} else if (sortOption === "dateUpdated") {
				if (sortOrder === "asc") {
					items.sort((a, b) => (a.UpdatedAt ?? 0) - (b.UpdatedAt ?? 0));
				} else if (sortOrder === "desc") {
					items.sort((a, b) => (b.UpdatedAt ?? 0) - (a.UpdatedAt ?? 0));
				}
			} else if (sortOption === "dateAdded") {
				if (sortOrder === "asc") {
					items.sort((a, b) => (a.AddedAt ?? 0) - (b.AddedAt ?? 0));
				} else if (sortOrder === "desc") {
					items.sort((a, b) => (b.AddedAt ?? 0) - (a.AddedAt ?? 0));
				}
			} else if (sortOption === "dateReleased") {
				if (sortOrder === "asc") {
					items.sort((a, b) => (a.ReleasedAt ?? 0) - (b.ReleasedAt ?? 0));
				} else if (sortOrder === "desc") {
					items.sort((a, b) => (b.ReleasedAt ?? 0) - (a.ReleasedAt ?? 0));
				}
			}

			// Filter by selected libraries
			if (filteredLibraries.length > 0) {
				items = items.filter((item) => filteredLibraries.includes(item.LibraryTitle));
			}

			// Filter out items already in the DB
			if (filterInDB === "notInDB") {
				items = items.filter((item) => !item.ExistInDatabase);
			} else if (filterInDB === "inDB") {
				items = items.filter((item) => item.ExistInDatabase);
			}

			// Filter out items by search
			const filteredItems = searchItems(items, searchQuery, {
				getTitle: (item) => item.Title,
				getYear: (item) => item.Year,
				getLibraryTitle: (item) => item.LibraryTitle,
				getID: (item) => item.TMDB_ID || item.RatingKey,
			});

			// Store the filtered items in local storage
			setFilteredAndSortedMediaItems(filteredItems);
		};
		filterAndSortItems();
	}, [
		librarySections,
		filteredLibraries,
		setFilteredAndSortedMediaItems,
		searchQuery,
		filterInDB,
		sortOption,
		sortOrder,
	]);

	if (error) {
		return <ErrorMessage error={error} />;
	}

	const hasUpdatedAt = paginatedItems.some((item) => item.UpdatedAt !== undefined && item.UpdatedAt !== null);

	return (
		<div className="flex items-center justify-center">
			{!fullyLoaded && librarySections.length > 0 ? (
				<div className="min-h-screen pb-4 px-4 sm:px-10 w-full">
					{/* Progress bars */}
					<div className="flex flex-col items-center w-full px-4">
						{[...librarySections]
							.sort((a, b) => {
								const progressA = sectionProgress[a.ID];
								const percentA =
									progressA && progressA.total > 0
										? Math.min((progressA.loaded / progressA.total) * 100, 100)
										: 0;
								const progressB = sectionProgress[b.ID];
								const percentB =
									progressB && progressB.total > 0
										? Math.min((progressB.loaded / progressB.total) * 100, 100)
										: 0;
								return percentB - percentA; // Sort descending
							})
							.map((section) => {
								const progressInfo = sectionProgress[section.ID];
								const percentage =
									progressInfo && progressInfo.total > 0
										? Math.min((progressInfo.loaded / progressInfo.total) * 100, 100)
										: 0;

								return (
									<div
										key={section.ID}
										className="mb-6 w-full max-w-xl flex flex-col items-center px-2"
									>
										<Label className="text-lg font-semibold text-center mb-2">
											Loading {section.Title}
										</Label>
										<Progress
											value={percentage}
											className={cn(
												"w-full max-w-lg h-2 rounded-md overflow-hidden",
												percentage < 100 && "animate-pulse",
												percentage >= 0 && percentage < 20 && "[&>div]:bg-yellow-100",
												percentage >= 20 && percentage < 40 && "[&>div]:bg-yellow-300",
												percentage >= 40 && percentage < 60 && "[&>div]:bg-green-200",
												percentage >= 60 && percentage < 80 && "[&>div]:bg-green-300",
												percentage >= 80 && percentage < 100 && "[&>div]:bg-green-400",
												percentage === 100 && "[&>div]:bg-green-500"
											)}
										/>

										{percentage < 100 && <Loader className="animate-spin mt-2" />}
										<span className="mt-2 text-base text-muted-foreground font-medium">
											{Math.round(percentage)}%
											{typeof progressInfo?.total === "number" && progressInfo.total > 0
												? ` - ${progressInfo.loaded} / ${progressInfo.total} items`
												: ""}
										</span>
									</div>
								);
							})}
					</div>
					<HomeMediaItemCardSkeletonGrid />
				</div>
			) : (
				<div className="min-h-screen pb-4 px-4 sm:px-10 w-full">
					{/* Filter & Sort Controls */}
					<div className="w-full flex items-center justify-center mb-4 mt-4">
						<FilterHome
							librarySections={librarySections}
							filteredLibraries={filteredLibraries}
							setFilteredLibraries={setFilteredLibraries}
							filterInDB={filterInDB}
							setFilterInDB={setFilterInDB}
							hasUpdatedAt={hasUpdatedAt}
							sortOption={sortOption}
							setSortOption={setSortOption}
							sortOrder={sortOrder}
							setSortOrder={setSortOrder}
							setCurrentPage={setCurrentPage}
							itemsPerPage={itemsPerPage}
							setItemsPerPage={setItemsPerPage}
						/>
					</div>

					{/* Grid of Cards */}
					<ResponsiveGrid size="regular">
						{paginatedItems.length === 0 && fullyLoaded && (searchQuery || filteredLibraries.length > 0) ? (
							<div className="col-span-full text-center text-red-500">
								<ErrorMessage
									error={ReturnErrorMessage<string>(
										`No items found${searchQuery ? ` matching "${searchQuery}"` : ""} in ${
											filteredLibraries.length > 0 ? filteredLibraries.join(", ") : "any library"
										}${
											filterInDB === "notInDB"
												? " that are not in the database."
												: filterInDB === "inDB"
													? " that are already in the database."
													: ""
										}`
									)}
								/>
							</div>
						) : (
							paginatedItems.map((item) => <HomeMediaItemCard key={item.RatingKey} item={item} />)
						)}
					</ResponsiveGrid>

					{/* Pagination */}
					<CustomPagination
						currentPage={currentPage}
						totalPages={totalPages}
						setCurrentPage={setCurrentPage}
						scrollToTop={true}
						filterItemsLength={filteredAndSortedMediaItems.length}
						itemsPerPage={itemsPerPage}
					/>
					{/* Refresh Button */}
					<RefreshButton onClick={() => getMediaItems(false)} />
				</div>
			)}
		</div>
	);
}
