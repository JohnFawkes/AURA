"use client";

import { ReturnErrorMessage } from "@/services/api-error-return";
import { fetchCollectionItems } from "@/services/mediaserver/api-mediaserver-fetch-collection-items";

import { useCallback, useEffect, useRef, useState } from "react";

import { useRouter } from "next/navigation";

import { AssetImage } from "@/components/shared/asset-image";
import { CustomPagination } from "@/components/shared/custom-pagination";
import { ErrorMessage } from "@/components/shared/error-message";
import { FilterCollections } from "@/components/shared/filter-collections";
import Loader from "@/components/shared/loader";
import { RefreshButton } from "@/components/shared/refresh-button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { H4 } from "@/components/ui/typography";

import { log } from "@/lib/logger";
import { useCollectionStore } from "@/lib/stores/global-store-collection-store";
import { MAX_CACHE_DURATION } from "@/lib/stores/global-store-library-sections";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";
import { useCollectionsPageStore } from "@/lib/stores/page-store-collections";

import { searchItems } from "@/hooks/search-query";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export interface CollectionItem {
	RatingKey: string;
	Title: string;
	Summary?: string;
	ChildCount: number;
	MediaItems: MediaItem[];
	LibraryTitle?: string;
}

export default function CollectionsPage() {
	const router = useRouter();

	useEffect(() => {
		document.title = "aura | Collections";
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

	const [collectionItems, setCollectionItems] = useState<CollectionItem[]>([]);

	const {
		collectionItems: storedCollectionItems,
		setCollectionItems: setStoredCollectionItems,
		timestamp: storedTimestamp,
		setCollectionItem,
	} = useCollectionStore();

	// State to track the CollectionsPageStore values
	const {
		filteredLibraries,
		setFilteredLibraries,
		currentPage,
		setCurrentPage,
		itemsPerPage,
		setItemsPerPage,
		sortOption,
		setSortOption,
		sortOrder,
		setSortOrder,
		filteredAndSortedCollectionItems,
		setFilteredAndSortedCollectionItems,
	} = useCollectionsPageStore();

	// -------------------------------
	// Derived values
	// -------------------------------
	const paginatedItems = filteredAndSortedCollectionItems.slice(
		(currentPage - 1) * itemsPerPage,
		currentPage * itemsPerPage
	);
	const totalPages = Math.ceil(filteredAndSortedCollectionItems.length / itemsPerPage);

	// Set sortOption to "title" if its not title, numberOfItems
	useEffect(() => {
		if (sortOption !== "title" && sortOption !== "numberOfItems") {
			setSortOption("title");
			setSortOrder("desc");
		}
	}, [sortOption, setSortOption, setSortOrder]);

	// Fetch data from cache or API
	const getCollectionItems = useCallback(
		async (useCache: boolean) => {
			if (isMounted.current && useCache) return;
			setError(null);
			setFullyLoaded(false);
			setCollectionItems([]);
			try {
				// Check if we want to use cache
				if (useCache) {
					const isCacheAgeValid = storedTimestamp ? Date.now() - storedTimestamp < MAX_CACHE_DURATION : false;
					const cacheContainsCollectionItemsAndTimestamp =
						storedCollectionItems && storedTimestamp && storedCollectionItems.length > 0;
					log("INFO", "Collections Page", "Library Cache", "Attempting to load collection items from cache", {
						"Current Time": Date.now(),
						"Cache Timestamp": storedTimestamp,
						"Cache Age Max (ms)": MAX_CACHE_DURATION,
						"Cache Age (ms)": storedTimestamp ? Date.now() - storedTimestamp : "N/A",
						"Is Cache Age Valid": isCacheAgeValid,
						"Cache Contains Collection Items & Timestamp": cacheContainsCollectionItemsAndTimestamp,
					});
					if (cacheContainsCollectionItemsAndTimestamp) {
						if (isCacheAgeValid) {
							setCollectionItems(storedCollectionItems);
							setFullyLoaded(true);
							log("INFO", "Collections Page", "Library Cache", "Loaded collection items from cache", {
								"Number of Items": storedCollectionItems.length,
							});
							return;
						} else {
							log(
								"WARN",
								"Collections Page",
								"Library Cache",
								"Cache is stale, fetching fresh collection items from API"
							);
						}
					} else {
						log(
							"WARN",
							"Collections Page",
							"Library Cache",
							"No valid cache found, fetching collection items from API"
						);
					}
				}

				const response = await fetchCollectionItems();
				if (response.status === "error") {
					setError(response);
					setFullyLoaded(true);
					return;
				}

				const fetchedCollectionItems = response.data || [];
				if (!fetchedCollectionItems || fetchedCollectionItems.length === 0) {
					setError(ReturnErrorMessage("No Collection Items found in Media Server"));
					return;
				}

				log(
					"INFO",
					"Collections Page",
					"Fetched Collection Items:",
					`Fetched ${fetchedCollectionItems.length} items`,
					{ fetchedCollectionItems }
				);

				// Store in global store for caching
				setStoredCollectionItems(fetchedCollectionItems, Date.now());

				setCollectionItems(fetchedCollectionItems);
				setFullyLoaded(true);
			} catch (error) {
				setError(ReturnErrorMessage<unknown>(error));
			} finally {
				isMounted.current = false;
			}
		},
		[setStoredCollectionItems, storedCollectionItems, storedTimestamp]
	);

	useEffect(() => {
		getCollectionItems(true);
		isMounted.current = true;
	}, [getCollectionItems]);

	useEffect(() => {
		if (searchQuery !== prevSearchQuery.current) {
			setCurrentPage(1);
			prevSearchQuery.current = searchQuery;
		}
	}, [searchQuery, setCurrentPage]);

	// Filter Items
	useEffect(() => {
		const filterAndSortItems = async () => {
			let items = [...collectionItems];

			// Sort items by Title
			if (sortOption === "title") {
				if (sortOrder === "asc") {
					items.sort((a, b) => a.Title.localeCompare(b.Title));
				} else if (sortOrder === "desc") {
					items.sort((a, b) => b.Title.localeCompare(a.Title));
				}
			} else if (sortOption === "numberOfItems") {
				if (sortOrder === "asc") {
					items.sort((a, b) => a.ChildCount - b.ChildCount);
				} else if (sortOrder === "desc") {
					items.sort((a, b) => b.ChildCount - a.ChildCount);
				}
			}

			// Filter by Libraries
			if (filteredLibraries.length > 0) {
				items = items.filter((item) => item.LibraryTitle && filteredLibraries.includes(item.LibraryTitle));
			}

			// Filter out items by search
			const filteredItems = searchItems(items, searchQuery, {
				getTitle: (item) => item.Title,
				getLibraryTitle: (item) => item.LibraryTitle,
				getID: (item) => item.RatingKey,
			});

			// Store the filtered and sorted items in local storage
			setFilteredAndSortedCollectionItems(filteredItems);
		};

		filterAndSortItems();
	}, [collectionItems, filteredLibraries, sortOption, sortOrder, setFilteredAndSortedCollectionItems, searchQuery]);

	const handleCardClick = (collectionItem: CollectionItem) => {
		setCollectionItem(collectionItem);
		router.push("/collection-item/");
	};

	if (error) {
		return <ErrorMessage error={error} />;
	}

	return (
		<div className="min-h-screen px-0 sm:px-0 pb-0 flex items-center justify-center">
			<div className="min-h-screen px-8 pb-20 sm:px-20 w-full">
				{fullyLoaded ? (
					<div className="min-h-screen px-8 pb-20 sm:px-20 w-full">
						{/* Filter & Sort Controls */}
						<div className="w-full flex items-center justify-center mb-4 mt-4">
							<FilterCollections
								librarySections={[
									...new Set(collectionItems.map((item) => item.LibraryTitle || "Unknown")),
								]}
								filteredLibraries={filteredLibraries}
								setFilteredLibraries={setFilteredLibraries}
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
						<div className="w-full grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4">
							{paginatedItems.length === 0 &&
							fullyLoaded &&
							(searchQuery || filteredLibraries.length > 0) ? (
								<div className="col-span-full text-center text-red-500">
									<ErrorMessage
										error={ReturnErrorMessage<string>(
											`No collection items found${searchQuery ? ` matching "${searchQuery}"` : ""} in ${
												filteredLibraries.length > 0
													? filteredLibraries.join(", ")
													: "any library"
											}`
										)}
									/>
								</div>
							) : (
								paginatedItems.map((item) => (
									<Card
										key={item.RatingKey}
										className="relative items-center cursor-pointer hover:shadow-xl transition-shadow"
										style={{ backgroundColor: "var(--card)" }}
										onClick={() => handleCardClick(item)}
									>
										{/* Poster Image */}
										<AssetImage
											image={`/api/mediaserver/image?ratingKey=${item.RatingKey}&imageType=poster`}
											className="w-[170px] h-auto transition-transform hover:scale-105"
										/>

										{/* Title */}
										<H4 className="text-center font-semibold mb-2 px-2">
											{item.Title.length > 55 ? `${item.Title.slice(0, 55)}...` : item.Title}
										</H4>

										<div className="flex justify-center gap-2">
											<Badge variant="default" className="mt-2 mb-2">
												{item.ChildCount} Items
											</Badge>
											<Badge variant="default" className="mt-2 mb-2">
												{item.LibraryTitle}
											</Badge>
										</div>
									</Card>
								))
							)}
						</div>

						{/* Pagination */}
						<CustomPagination
							currentPage={currentPage}
							totalPages={totalPages}
							setCurrentPage={setCurrentPage}
							scrollToTop={true}
							filterItemsLength={filteredAndSortedCollectionItems.length}
							itemsPerPage={itemsPerPage}
						/>

						{/* Refresh Button */}
						<RefreshButton onClick={() => getCollectionItems(false)} />
					</div>
				) : (
					<Loader className="mt-20" message="Loading Collection Items" />
				)}
			</div>
		</div>
	);
}
