"use client";

import { setRefsToFormItems } from "@/helper/download-modal/set-to-form-item";
import { formatLastUpdatedDate } from "@/helper/format-date-last-updates";
import { TMDBLookupMap, createTMDBLookupMap, searchWithLookupMap } from "@/helper/search-idb-for-tmdb-id";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { getAllUserSets } from "@/services/mediux/sets-user";
import { ArrowDownAZ, ArrowDownZA, ClockArrowDown, ClockArrowUp, User } from "lucide-react";

import { useEffect, useRef, useState } from "react";

import Link from "next/link";
import { useParams } from "next/navigation";

import { AssetImage } from "@/components/shared/asset-image";
import { RenderBoxsetDisplay, RenderShowAndCollectionDisplay } from "@/components/shared/boxset-display";
import { CustomPagination } from "@/components/shared/custom-pagination";
import DownloadModal from "@/components/shared/download-modal";
import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { ResponsiveGrid } from "@/components/shared/responsive-grid";
import { SelectItemsPerPage } from "@/components/shared/select-items-per-page";
import { SortControl } from "@/components/shared/select-sort";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ToggleGroup } from "@/components/ui/toggle-group";
import { Lead, P } from "@/components/ui/typography";

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";
import { useUserPageStore } from "@/lib/stores/page-store-user";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { BoxsetRef, CreatorSetsResponse, IncludedItem, SetRef } from "@/types/media-and-posters/sets";

function processSets(
	sets: SetRef[],
	tmdbLookupMap: TMDBLookupMap,
	includedItems: { [key: string]: IncludedItem },
	setSetter: (sets: SetRef[]) => void
) {
	const tmdbIDs = new Set<string>();
	for (const set of sets) {
		if (Array.isArray(set.item_ids)) {
			for (const id of set.item_ids) {
				tmdbIDs.add(id);
			}
		}
	}
	const mediaItemByTMDB = new Map<string, any>();
	for (const tmdbId of tmdbIDs) {
		const mediaItem = searchWithLookupMap(tmdbId, tmdbLookupMap);
		if (mediaItem && typeof mediaItem !== "boolean") {
			mediaItemByTMDB.set(tmdbId, mediaItem);
		}
	}
	const includedByTMDB: { [tmdb_id: string]: IncludedItem } = {};
	for (const tmdbId of mediaItemByTMDB.keys()) {
		const included = Object.values(includedItems || {}).find((ii) => ii?.mediux_info?.tmdb_id === tmdbId);
		if (!included) continue;
		includedByTMDB[tmdbId] = {
			...included,
			media_item: mediaItemByTMDB.get(tmdbId),
		} as IncludedItem;
	}
	const filteredSets: SetRef[] = [];
	for (const set of sets) {
		if (Array.isArray(set.item_ids)) {
			for (const itemID of set.item_ids) {
				if (includedByTMDB[itemID]) {
					filteredSets.push(set);
					break;
				}
			}
		}
	}
	setSetter(filteredSets);
}

function getBoxsetSetsForLibrary(
	boxset: BoxsetRef,
	userResponse: CreatorSetsResponse,
	libraryType: "show" | "movie",
	validSetIds: Set<string>
): SetRef[] {
	const sets: SetRef[] = [];

	if (libraryType === "show" && boxset.set_ids.show) {
		for (const setId of boxset.set_ids.show) {
			if (validSetIds.has(setId)) {
				const set = userResponse.show_sets.find((s) => s.id === setId);
				if (set) sets.push(set);
			}
		}
	}

	if (libraryType === "movie") {
		if (boxset.set_ids.movie) {
			for (const setId of boxset.set_ids.movie) {
				if (validSetIds.has(setId)) {
					const set = userResponse.movie_sets.find((s) => s.id === setId);
					if (set) sets.push(set);
				}
			}
		}
		if (boxset.set_ids.collection) {
			for (const setId of boxset.set_ids.collection) {
				if (validSetIds.has(setId)) {
					const set = userResponse.collection_sets.find((s) => s.id === setId);
					if (set) sets.push(set);
				}
			}
		}
	}

	return sets;
}

export interface BoxsetsWithSetInfo extends BoxsetRef {
	sets?: SetRef[];
}

const UserSetPage = () => {
	// Get the username from the URL
	const { username } = useParams();
	const hasFetchedInfo = useRef(false);
	// Error Handling
	const [error, setError] = useState<APIResponse<unknown> | null>(null);
	const [isLoading, setIsLoading] = useState(true);
	const [loadMessage, setLoadMessage] = useState("");
	const { currentPage, setCurrentPage, itemsPerPage, setItemsPerPage } = useUserPageStore();

	const { searchQuery, setSearchQuery } = useSearchQueryStore();
	const prevSearchQuery = useRef(searchQuery);

	// Pagination and Active Tab state
	const [activeTab, setActiveTab] = useState("boxSets");
	const [totalPages, setTotalPages] = useState(0);

	// Library sections & progress
	const [librarySections, setLibrarySections] = useState<{ title: string; type: string }[]>([]);
	const [selectedLibrarySection, setSelectedLibrarySection] = useState<{
		title: string;
		type: "show" | "movie";
	} | null>(null);
	const [filterOutInDB, setFilterOutInDB] = useState<"all" | "inDB" | "notInDB">("all");

	// State to track the selected sorting option
	const { sortOption, setSortOption, sortOrder, setSortOrder } = useUserPageStore();

	const { sections, getSectionSummaries, hasHydrated } = useLibrarySectionsStore();

	// User Response From Server
	const [creatorSetsResponse, setCreatorSetsResponse] = useState<CreatorSetsResponse | null>(null);

	const [showSets, setShowSets] = useState<SetRef[]>([]);
	const [filteredShowSets, setFilteredShowSets] = useState<SetRef[]>([]);
	const [movieSets, setMovieSets] = useState<SetRef[]>([]);
	const [filteredMovieSets, setFilteredMovieSets] = useState<SetRef[]>([]);
	const [collectionSets, setCollectionSets] = useState<SetRef[]>([]);
	const [filteredCollectionSets, setFilteredCollectionSets] = useState<SetRef[]>([]);
	const [boxsets, setBoxsets] = useState<BoxsetsWithSetInfo[]>([]);
	const [filteredBoxsets, setFilteredBoxsets] = useState<BoxsetsWithSetInfo[]>([]);
	const [tmdbLookupMap, setTMDBLookupMap] = useState<TMDBLookupMap>({});
	const [setIncludedItems, setSetIncludedItems] = useState<{
		[tmdb_id: string]: IncludedItem;
	} | null>(null);

	useEffect(() => {
		if (!hasHydrated) return;
		const sections = getSectionSummaries();
		setLibrarySections(sections);
		setSelectedLibrarySection(null);
		log("INFO", "User Page", "Library Sections", "Fetched library sections from cache", sections);
	}, [getSectionSummaries, hasHydrated]);

	// Set sortOption to "dateLastUpdate" if it's not title or dateLastUpdate
	if (sortOption !== "title" && sortOption !== "dateLastUpdate") {
		setSortOption("dateLastUpdate");
		setSortOrder("desc");
	}

	// Get all of the sets for the user
	useEffect(() => {
		if (hasFetchedInfo.current) return;
		hasFetchedInfo.current = true;
		setLoadMessage(`Loading sets for ${username}`);
		const fetchAllUserSets = async () => {
			try {
				setIsLoading(true);
				setLoadMessage(`Loading sets for ${username}`);
				const response = await getAllUserSets(username as string);

				if (response.status === "error") {
					setError(response);
					return;
				}
				setSelectedLibrarySection(null);

				setCreatorSetsResponse(response.data || null);
			} catch (error) {
				log("ERROR", "User Page", "Fetch User Sets", "Failed to fetch user sets:", error);
				setError(ReturnErrorMessage<unknown>(error));
			} finally {
				setIsLoading(false);
			}
		};
		fetchAllUserSets();
	}, [username]);

	// When a library section is selected, set the showSets, movieSets, collectionSets, boxsets based on that section
	useEffect(() => {
		if (!creatorSetsResponse) return;
		if (
			creatorSetsResponse &&
			creatorSetsResponse.show_sets.length === 0 &&
			creatorSetsResponse.movie_sets.length === 0 &&
			creatorSetsResponse.collection_sets.length === 0 &&
			creatorSetsResponse.boxsets.length === 0
		) {
			return;
		}
		if (!selectedLibrarySection) {
			return;
		}

		setShowSets([]);
		setMovieSets([]);
		setCollectionSets([]);
		setBoxsets([]);
		setTMDBLookupMap({});
		setSetIncludedItems(null);

		const filterOutItems = async () => {
			setIsLoading(true);

			const librarySection = sections[selectedLibrarySection.title];
			if (!librarySection || !librarySection.media_items || librarySection.media_items.length === 0) {
				log(
					"ERROR",
					"User Page",
					"Fetch User Sets",
					`No data found for library section: ${selectedLibrarySection.title}`
				);
				setIsLoading(false);
				return;
			}

			const tmdbLookupMap = createTMDBLookupMap(librarySection.media_items);

			const includedItemsArray = Object.values(creatorSetsResponse.included_items ?? {});
			if (Object.keys(tmdbLookupMap).length > 0 && includedItemsArray.length > 0) {
				const nextIncludedByTMDB = includedItemsArray.reduce(
					(acc, includedItem) => {
						const tmdbId = includedItem?.mediux_info?.tmdb_id;
						if (!tmdbId) return acc;

						const mediaItem = searchWithLookupMap(tmdbId, tmdbLookupMap); // MediaItem | boolean
						acc[tmdbId] =
							mediaItem && typeof mediaItem !== "boolean"
								? ({
										...includedItem,
										media_item: mediaItem,
									} as IncludedItem)
								: includedItem;

						return acc;
					},
					{} as { [tmdb_id: string]: IncludedItem }
				);

				setSetIncludedItems(nextIncludedByTMDB);
			} else {
				setSetIncludedItems(null);
			}

			// Process Boxsets
			if (creatorSetsResponse.boxsets.length > 0) {
				// Get all valid set IDs for this library section
				let validSetIds = new Set<string>();
				let validTmdbIds = new Set<string>(librarySection.media_items.map((mi) => mi.tmdb_id));

				if (selectedLibrarySection.type === "show") {
					creatorSetsResponse.show_sets.forEach((set) => {
						if (Array.isArray(set.item_ids) && set.item_ids.some((id) => validTmdbIds.has(id))) {
							validSetIds.add(set.id);
						}
					});
				} else if (selectedLibrarySection.type === "movie") {
					[...creatorSetsResponse.movie_sets, ...creatorSetsResponse.collection_sets].forEach((set) => {
						if (Array.isArray(set.item_ids) && set.item_ids.some((id) => validTmdbIds.has(id))) {
							validSetIds.add(set.id);
						}
					});
				}

				// Filter boxsets to only include sets present in the current library section
				const filteredBoxsets = creatorSetsResponse.boxsets
					.map((boxset) => ({
						...boxset,
						sets: getBoxsetSetsForLibrary(
							boxset,
							creatorSetsResponse,
							selectedLibrarySection.type,
							validSetIds
						),
					}))
					.filter((boxset) => boxset.sets.length > 0);

				setBoxsets(filteredBoxsets);

				// Build a map of tmdb_id -> MediaItem for the current library section
				const libraryMediaItemMap = new Map<string, MediaItem>();
				librarySection.media_items.forEach((mi) => {
					libraryMediaItemMap.set(mi.tmdb_id, mi);
				});

				// Collect all tmdb_ids used in the sets of filtered boxsets
				const allBoxsetTmdbIds = new Set<string>();
				filteredBoxsets.forEach((boxset) => {
					boxset.sets.forEach((set) => {
						set.item_ids.forEach((id) => {
							if (libraryMediaItemMap.has(id)) {
								allBoxsetTmdbIds.add(id);
							}
						});
					});
				});
			}

			// Process Show Sets
			if (selectedLibrarySection.type === "show" && creatorSetsResponse.show_sets.length > 0) {
				processSets(
					creatorSetsResponse.show_sets,
					tmdbLookupMap,
					creatorSetsResponse.included_items,
					setShowSets
				);
			} else if (selectedLibrarySection.type === "movie") {
				// Process Movie Sets
				if (creatorSetsResponse.movie_sets.length > 0) {
					processSets(
						creatorSetsResponse.movie_sets,
						tmdbLookupMap,
						creatorSetsResponse.included_items,
						setMovieSets
					);
				}

				// Process Collection Sets
				if (creatorSetsResponse.collection_sets.length > 0) {
					processSets(
						creatorSetsResponse.collection_sets,
						tmdbLookupMap,
						creatorSetsResponse.included_items,
						setCollectionSets
					);
				}
			}

			setTMDBLookupMap(tmdbLookupMap);
			setIsLoading(false);
		};

		filterOutItems();
	}, [sections, selectedLibrarySection, creatorSetsResponse]);

	useEffect(() => {
		if (!creatorSetsResponse) return;
		if (
			creatorSetsResponse &&
			creatorSetsResponse.show_sets.length === 0 &&
			creatorSetsResponse.movie_sets.length === 0 &&
			creatorSetsResponse.collection_sets.length === 0 &&
			creatorSetsResponse.boxsets.length === 0
		) {
			return;
		}

		// Always reset filtered sets to all sets if search is blank or filterOutInDB is "all"
		if (searchQuery.trim() === "" && filterOutInDB === "all") {
			setFilteredShowSets(showSets);
			setFilteredMovieSets(movieSets);
			setFilteredCollectionSets(collectionSets);
			setFilteredBoxsets(boxsets);
			return;
		}

		if (searchQuery !== prevSearchQuery.current) {
			setCurrentPage(1);
			prevSearchQuery.current = searchQuery;

			if (!selectedLibrarySection) {
				return;
			}
			// Filter out the Show Sets on set title and mediaItem.Title
			const filteredShowSets = showSets.filter((showSet) => {
				const query = searchQuery.toLowerCase();

				// Match on set title
				if (showSet.title.toLowerCase().includes(query)) return true;

				// Match on any media item's title in the set
				return showSet.item_ids.some((tmdbId) => {
					const mediaItem = searchWithLookupMap(tmdbId, tmdbLookupMap);
					return (
						mediaItem &&
						typeof mediaItem !== "boolean" &&
						mediaItem.title &&
						mediaItem.title.toLowerCase().includes(query)
					);
				});
			});
			setFilteredShowSets(filteredShowSets);

			// Filter out the Movie Sets on set title and mediaItem.Title
			const filteredMovieSets = movieSets.filter((movieSet) => {
				const query = searchQuery.toLowerCase();
				// Match on set title
				if (movieSet.title.toLowerCase().includes(query)) return true;
				// Match on any media item's title in the set
				return movieSet.item_ids.some((tmdbId) => {
					const mediaItem = searchWithLookupMap(tmdbId, tmdbLookupMap);
					return (
						mediaItem &&
						typeof mediaItem !== "boolean" &&
						mediaItem.title &&
						mediaItem.title.toLowerCase().includes(query)
					);
				});
			});
			setFilteredMovieSets(filteredMovieSets);

			// Filter out the Collection Sets on set title and mediaItem.Title
			const filteredCollectionSets = collectionSets.filter((collectionSet) => {
				const query = searchQuery.toLowerCase();
				// Match on set title
				if (collectionSet.title.toLowerCase().includes(query)) return true;
				// Match on any media item's title in the set
				return collectionSet.item_ids.some((tmdbId) => {
					const mediaItem = searchWithLookupMap(tmdbId, tmdbLookupMap);
					return (
						mediaItem &&
						typeof mediaItem !== "boolean" &&
						mediaItem.title &&
						mediaItem.title.toLowerCase().includes(query)
					);
				});
			});
			setFilteredCollectionSets(filteredCollectionSets);

			// Filter out the Box Sets on set title and mediaItem.Title
			// Boxsets don't have sets, they have
			const filteredBoxSets = boxsets.filter((boxset) => {
				const query = searchQuery.toLowerCase();
				// Match on boxset title
				if (boxset.title.toLowerCase().includes(query)) return true;
				// Match on any set's title in the boxset
				if (boxset.sets) {
					for (const set of boxset.sets) {
						if (set.title.toLowerCase().includes(query)) return true;

						// Match on any media item's title in the set
						for (const tmdbId of set.item_ids) {
							const mediaItem = searchWithLookupMap(tmdbId, tmdbLookupMap);
							if (
								mediaItem &&
								typeof mediaItem !== "boolean" &&
								mediaItem.title &&
								mediaItem.title.toLowerCase().includes(query)
							) {
								return true;
							}
						}
					}
				}
				return false;
			});
			setFilteredBoxsets(filteredBoxSets);
		}
	}, [
		showSets,
		movieSets,
		collectionSets,
		boxsets,
		searchQuery,
		filterOutInDB,
		selectedLibrarySection,
		setCurrentPage,
		creatorSetsResponse,
		filteredShowSets,
		filteredMovieSets,
		filteredCollectionSets,
		filteredBoxsets,
		tmdbLookupMap,
	]);

	useEffect(() => {
		if (filterOutInDB === "all") {
			// Show all items, do not filter further
			setFilteredShowSets(showSets);
			setFilteredMovieSets(movieSets);
			setFilteredCollectionSets(collectionSets);
			setFilteredBoxsets(boxsets);
			return;
		}

		const filterInDB = (sets: SetRef[]) =>
			sets.filter((set) =>
				set.item_ids.every((tmdbId) => {
					const mediaItem = searchWithLookupMap(tmdbId, tmdbLookupMap);
					return mediaItem && typeof mediaItem !== "boolean" && mediaItem.db_saved_sets.length > 0;
				})
			);

		const filterNotInDB = (sets: SetRef[]) =>
			sets.filter((set) =>
				set.item_ids.some((tmdbId) => {
					const mediaItem = searchWithLookupMap(tmdbId, tmdbLookupMap);
					return !mediaItem || (typeof mediaItem !== "boolean" && mediaItem.db_saved_sets.length === 0);
				})
			);

		if (filterOutInDB === "inDB") {
			setFilteredShowSets((prev) => filterInDB(prev));
			setFilteredMovieSets((prev) => filterInDB(prev));
			setFilteredCollectionSets((prev) => filterInDB(prev));
			setFilteredBoxsets((prevBoxsets) =>
				prevBoxsets.filter((boxset) =>
					boxset.sets?.every((set) =>
						set.item_ids.every((tmdbId) => {
							const mediaItem = searchWithLookupMap(tmdbId, tmdbLookupMap);
							return mediaItem && typeof mediaItem !== "boolean" && mediaItem.db_saved_sets.length > 0;
						})
					)
				)
			);
		} else if (filterOutInDB === "notInDB") {
			setFilteredShowSets((prev) => filterNotInDB(prev));
			setFilteredMovieSets((prev) => filterNotInDB(prev));
			setFilteredCollectionSets((prev) => filterNotInDB(prev));
			setFilteredBoxsets((prevBoxsets) =>
				prevBoxsets.filter((boxset) =>
					boxset.sets?.some((set) =>
						set.item_ids.some((tmdbId) => {
							const mediaItem = searchWithLookupMap(tmdbId, tmdbLookupMap);
							return (
								!mediaItem || (typeof mediaItem !== "boolean" && mediaItem.db_saved_sets.length === 0)
							);
						})
					)
				)
			);
		}
	}, [
		filterOutInDB,
		tmdbLookupMap,
		setFilteredShowSets,
		setFilteredMovieSets,
		setFilteredCollectionSets,
		setFilteredBoxsets,
		showSets,
		movieSets,
		collectionSets,
		boxsets,
	]);

	useEffect(() => {
		setTotalPages(
			Math.ceil(
				(activeTab === "showSets"
					? showSets.length
					: activeTab === "movieSets"
						? movieSets.length
						: activeTab === "collectionSets"
							? collectionSets.length
							: boxsets.length) / itemsPerPage
			)
		);
		log("INFO", "User Page", "Fetch User Sets", "User Page - Total Pages", totalPages);
		setCurrentPage(1); // Reset to first page when tab changes
	}, [
		activeTab,
		boxsets.length,
		collectionSets.length,
		itemsPerPage,
		movieSets.length,
		showSets.length,
		totalPages,
		setCurrentPage,
	]);

	const paginatedShowSets = filteredShowSets.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);
	const paginatedMovieSets = filteredMovieSets.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);
	const paginatedCollectionSets = filteredCollectionSets.slice(
		(currentPage - 1) * itemsPerPage,
		currentPage * itemsPerPage
	);
	const paginatedBoxSets = filteredBoxsets.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

	return (
		<div className="flex flex-col">
			{/* Show loading message */}
			{isLoading && (
				<div className="flex justify-center mt-4">
					<Loader message={loadMessage} />
				</div>
			)}

			{/* Show error message if there is an error */}
			{error && (
				<div className="flex justify-center">
					<ErrorMessage error={error} />
				</div>
			)}

			{/* Show message when no sets are found */}
			{!isLoading &&
				!error &&
				creatorSetsResponse &&
				creatorSetsResponse.show_sets.length === 0 &&
				creatorSetsResponse.movie_sets.length === 0 &&
				creatorSetsResponse.collection_sets.length === 0 &&
				creatorSetsResponse.boxsets.length === 0 && (
					<div className="flex justify-center mt-4">
						<P>No sets found for {username}</P>
					</div>
				)}

			{/* Main content when sets exist */}
			{!isLoading &&
				!error &&
				creatorSetsResponse &&
				(creatorSetsResponse.show_sets.length > 0 ||
					creatorSetsResponse.movie_sets.length > 0 ||
					creatorSetsResponse.collection_sets.length > 0 ||
					creatorSetsResponse.boxsets.length > 0) && (
					<div className="min-h-screen px-4 sm:px-8 pb-20">
						{/* User Sets Header */}
						<div className="flex flex-col items-center mt-8 mb-6">
							<h1 className="text-4xl font-extrabold text-center mb-2 tracking-tight text-primary flex items-center justify-center gap-2">
								<span className="text-white opacity-80">Sets by</span>
								<span className="text-primary">{username}</span>
								<Avatar className="rounded-lg w-7 h-7 ml-2 align-middle">
									<AvatarImage
										src={`/api/images/mediux/avatar?username=${username}`}
										className="w-7 h-7"
									/>
									<AvatarFallback>
										<User className="w-7 h-7" />
									</AvatarFallback>
								</Avatar>
							</h1>
							{!selectedLibrarySection ||
								(selectedLibrarySection.title == "" && (
									<div className="flex flex-wrap gap-3 mt-2 justify-center">
										{showSets.length > 0 && (
											<div className="flex items-center gap-2 bg-background border border-border rounded-lg px-4 py-2 shadow-sm">
												<span className="font-semibold text-primary">Show Sets</span>
												<Badge variant="secondary" className="text-xs px-2 py-1">
													{showSets.length}
												</Badge>
											</div>
										)}
										{movieSets.length > 0 && (
											<div className="flex items-center gap-2 bg-background border border-border rounded-lg px-4 py-2 shadow-sm">
												<span className="font-semibold text-primary">Movie Sets</span>
												<Badge variant="secondary" className="text-xs px-2 py-1">
													{movieSets.length}
												</Badge>
											</div>
										)}
										{collectionSets.length > 0 && (
											<div className="flex items-center gap-2 bg-background border border-border rounded-lg px-4 py-2 shadow-sm">
												<span className="font-semibold text-primary">Collection Sets</span>
												<Badge variant="secondary" className="text-xs px-2 py-1">
													{collectionSets.length}
												</Badge>
											</div>
										)}
										{boxsets.length > 0 && (
											<div className="flex items-center gap-2 bg-background border border-border rounded-lg px-4 py-2 shadow-sm">
												<span className="font-semibold text-primary">Box Sets</span>
												<Badge variant="secondary" className="text-xs px-2 py-1">
													{boxsets.length}
												</Badge>
											</div>
										)}
									</div>
								))}
						</div>

						{/* Library Section Selection */}
						<div className="w-full max-w-3xl">
							{/* Library Section */}
							<div className="flex flex-col sm:flex-row mb-4 mt-2">
								<Label htmlFor="library-filter" className="text-lg font-semibold mb-2 sm:mb-0 sm:mr-4">
									Libraries:
								</Label>

								<ToggleGroup
									type="single"
									className="flex flex-wrap sm:flex-nowrap gap-2"
									value={
										selectedLibrarySection && selectedLibrarySection.title
											? selectedLibrarySection.title
											: ""
									}
									onValueChange={(val: string) => {
										const found = librarySections.find(
											(section) =>
												section.title === val &&
												(section.type === "show" || section.type === "movie")
										);
										setSelectedLibrarySection(
											found ? { ...found, type: found.type as "show" | "movie" } : null
										);
									}}
								>
									{librarySections.map((section) => (
										<Badge
											key={section.title}
											variant={
												selectedLibrarySection?.title === section.title ? "default" : "outline"
											}
											onClick={() => {
												const safeSection =
													section.type === "show" || section.type === "movie"
														? { ...section, type: section.type as "show" | "movie" }
														: null;
												if (selectedLibrarySection?.title === section.title) {
													setSelectedLibrarySection(null);
													setCurrentPage(1);
													setFilterOutInDB("all");
												} else {
													setSelectedLibrarySection(safeSection);
													setCurrentPage(1);
													setFilterOutInDB("all");
												}
												setSearchQuery("");
											}}
											className={`cursor-pointer text-sm active:scale-95 hover:brightness-120 ${
												!!selectedLibrarySection &&
												selectedLibrarySection.title !== section.title
													? "opacity-50 hover:opacity-100"
													: ""
											}`}
										>
											{section.title}
										</Badge>
									))}
								</ToggleGroup>
							</div>
						</div>

						{/* No library selected message */}
						{!selectedLibrarySection && (
							<div className="flex justify-center mt-8">
								<ErrorMessage
									isWarning={true}
									error={ReturnErrorMessage<string>(
										"No library selected. Select one to get started."
									)}
								/>
							</div>
						)}

						{selectedLibrarySection &&
							(showSets.length === 0 &&
							movieSets.length === 0 &&
							collectionSets.length === 0 &&
							boxsets.length === 0 ? (
								<>
									{/* If the filterOutInDB is selected, show an option to unselect it */}
									<div className="flex justify-center">
										{/* Filter Out In DB Selection */}
										<div className="w-full flex items-center mb-2">
											<Label htmlFor="filter-out-in-db" className="text-lg font-semibold mr-2">
												Filter:
											</Label>
											{/* Filter Out In DB Toggle */}

											<Badge
												key="filter-out-in-db"
												className={`cursor-pointer text-sm active:scale-95 hover:brightness-120 ${
													filterOutInDB === "inDB"
														? "bg-green-600 text-white"
														: filterOutInDB === "notInDB"
															? "bg-red-600 text-white"
															: ""
												}`}
												variant={filterOutInDB !== "all" ? "default" : "outline"}
												onClick={() => {
													const next =
														filterOutInDB === "all"
															? "inDB"
															: filterOutInDB === "inDB"
																? "notInDB"
																: "all";
													setFilterOutInDB(next);
													setCurrentPage(1);
												}}
											>
												{filterOutInDB === "all"
													? "All Items"
													: filterOutInDB === "inDB"
														? "Items In DB"
														: "Items Not in DB"}
											</Badge>
										</div>
									</div>
									<div className="flex justify-center mt-8">
										<ErrorMessage
											error={ReturnErrorMessage<string>(
												`No Sets found in ${selectedLibrarySection.title} library${
													filterOutInDB === "inDB"
														? " that exist in your database"
														: filterOutInDB === "notInDB"
															? " that are missing from your database"
															: ""
												}${searchQuery ? ` for search query "${searchQuery}"` : ""}`
											)}
										/>
									</div>
								</>
							) : (
								<div className="flex flex-col items-center mt-4 mb-4">
									{/* Filter Out In DB Selection */}
									<div className="w-full flex items-center mb-2">
										<Label htmlFor="filter-out-in-db" className="text-lg font-semibold mr-2">
											Filter:
										</Label>
										{/* Filter Out In DB Toggle */}

										<Badge
											key="filter-out-in-db"
											className={`cursor-pointer text-sm active:scale-95 hover:brightness-120 ${
												filterOutInDB === "inDB"
													? "bg-green-600 text-white"
													: filterOutInDB === "notInDB"
														? "bg-red-600 text-white"
														: ""
											}`}
											variant={filterOutInDB !== "all" ? "default" : "outline"}
											onClick={() => {
												const next =
													filterOutInDB === "all"
														? "inDB"
														: filterOutInDB === "inDB"
															? "notInDB"
															: "all";
												setFilterOutInDB(next);
												setCurrentPage(1);
											}}
										>
											{filterOutInDB === "all"
												? "All Items"
												: filterOutInDB === "inDB"
													? "Items In DB"
													: "Items Not in DB"}
										</Badge>
									</div>

									{/* Items Per Page Selection */}
									<div className="w-full flex items-center mb-2">
										<SelectItemsPerPage
											setCurrentPage={setCurrentPage}
											itemsPerPage={itemsPerPage}
											setItemsPerPage={setItemsPerPage}
										/>
									</div>

									{/* Sort Control */}
									<div className="w-full flex items-center mb-4">
										{/* Sort Control */}
										<SortControl
											options={[
												{
													value: "dateLastUpdate",
													label: "Date Updated",
													ascIcon: <ClockArrowUp />,
													descIcon: <ClockArrowDown />,
													type: "date" as const,
												},

												{
													value: "title",
													label: "Title",
													ascIcon: <ArrowDownAZ />,
													descIcon: <ArrowDownZA />,
													type: "string" as const,
												},
											]}
											sortOption={sortOption}
											sortOrder={sortOrder}
											setSortOption={(value) => {
												setSortOption(value as "title" | "dateLastUpdate");
												if (value === "title") setSortOrder("asc");
												else if (value === "dateLastUpdate") setSortOrder("desc");
											}}
											setSortOrder={setSortOrder}
										/>
									</div>

									<Tabs
										defaultValue="boxSets"
										value={activeTab}
										onValueChange={(val) => {
											setActiveTab(val);
											setCurrentPage(1);
										}}
										className="mt-2 w-full"
									>
										<TabsList className="flex flex-wrap w-full rounded-none bg-transparent gap-2 justify-start px-2 mb-2 border-b">
											{showSets.length > 0 && (
												<TabsTrigger
													value="showSets"
													className="flex-1 cursor-pointer text-primary data-[state=active]:bg-primary data-[state=active]:text-background dark:data-[state=active]:bg-primary dark:data-[state=active]:text-background hover:brightness-120 active:scale-95"
												>
													Show Sets ({showSets.length})
												</TabsTrigger>
											)}
											{movieSets.length > 0 && (
												<TabsTrigger
													value="movieSets"
													className="flex-1 cursor-pointer text-primary data-[state=active]:bg-primary data-[state=active]:text-background dark:data-[state=active]:bg-primary dark:data-[state=active]:text-background hover:brightness-120 active:scale-95"
												>
													Movie Sets ({movieSets.length})
												</TabsTrigger>
											)}
											{collectionSets.length > 0 && (
												<TabsTrigger
													value="collectionSets"
													className="flex-1 cursor-pointer text-primary data-[state=active]:bg-primary data-[state=active]:text-background dark:data-[state=active]:bg-primary dark:data-[state=active]:text-background hover:brightness-120 active:scale-95"
												>
													Collection Sets ({collectionSets.length})
												</TabsTrigger>
											)}
											{boxsets.length > 0 && (
												<TabsTrigger
													value="boxSets"
													className="flex-1 cursor-pointer text-primary data-[state=active]:bg-primary data-[state=active]:text-background dark:data-[state=active]:bg-primary dark:data-[state=active]:text-background hover:brightness-120 active:scale-95"
												>
													Box Sets ({boxsets.length})
												</TabsTrigger>
											)}
										</TabsList>

										<div className="mt-4">
											{paginatedShowSets.length > 0 && (
												<TabsContent value="showSets">
													<div className="divide-y divide-primary-dynamic/20 space-y-6">
														{paginatedShowSets.map((showSet) => (
															<div key={`${showSet.id}-showset`} className="pb-6">
																<RenderShowAndCollectionDisplay
																	includedItems={setIncludedItems || {}}
																	set={showSet}
																/>
															</div>
														))}
													</div>
												</TabsContent>
											)}

											{paginatedMovieSets.length > 0 && (
												<TabsContent value="movieSets">
													<ResponsiveGrid size="regular">
														{filteredMovieSets.map((set) => (
															<div
																key={set.id}
																className="relative flex flex-col items-center p-2 border rounded-md"
																style={{
																	background: "oklch(0.16 0.0202 282.55)",
																	opacity: "0.95",
																	padding: "0.5rem",
																}}
															>
																<div className="relative w-full mb-1">
																	{/* Download Button - absolute top right */}
																	<div className="absolute top-0 right-0 z-10">
																		<DownloadModal
																			baseSetInfo={set}
																			formItems={setRefsToFormItems(
																				[set],
																				setIncludedItems || {}
																			)}
																		/>
																	</div>
																	{/* Set Name */}
																	<P className="text-primary-dynamic text-sm font-semibold w-full mb-1 truncate pr-10">
																		{set.title}
																	</P>
																</div>

																{/* Set User Name */}
																<div className="flex items-center justify-start w-full mb-1">
																	<div className="flex items-center gap-1">
																		<Avatar className="rounded-lg mr-1 w-4 h-4">
																			<AvatarImage
																				src={`/api/images/mediux/avatar?username=${set.user_created}`}
																				className="w-4 h-4"
																			/>
																			<AvatarFallback className="">
																				<User className="w-4 h-4" />
																			</AvatarFallback>
																		</Avatar>
																		<Link
																			href={`/user/${set.user_created}`}
																			className="text-sm hover:text-primary cursor-pointer underline truncate"
																			style={{ wordBreak: "break-word" }}
																		>
																			{set.user_created}
																		</Link>
																	</div>
																</div>

																{/* Last Update */}
																<Lead className="text-sm text-muted-foreground w-full mb-2">
																	Last Update:{" "}
																	{formatLastUpdatedDate(
																		set.date_updated,
																		set.date_created ||
																			set.images[0]?.modified ||
																			""
																	)}
																</Lead>

																{/* Poster */}
																{set.images.find(
																	(image) => image.type === "poster"
																) && (
																	<AssetImage
																		image={
																			set.images.find(
																				(image) => image.type === "poster"
																			)!
																		}
																		imageType="mediux"
																		aspect="poster"
																		className="w-full mb-2"
																	/>
																)}

																{/* Backdrop */}
																{set.images.find(
																	(image) => image.type === "backdrop"
																) && (
																	<AssetImage
																		image={
																			set.images.find(
																				(image) => image.type === "backdrop"
																			)!
																		}
																		imageType="mediux"
																		aspect="backdrop"
																		className="w-full"
																	/>
																)}
															</div>
														))}
													</ResponsiveGrid>
												</TabsContent>
											)}

											{paginatedCollectionSets.length > 0 && (
												<TabsContent value="collectionSets">
													<div className="divide-y divide-primary-dynamic/20 space-y-6">
														{paginatedCollectionSets.map((collectionSet) => (
															<div
																key={`${collectionSet.id}-collectionset`}
																className="pb-6"
															>
																<RenderShowAndCollectionDisplay
																	includedItems={setIncludedItems || {}}
																	set={collectionSet}
																/>
															</div>
														))}
													</div>
												</TabsContent>
											)}

											{paginatedBoxSets.length > 0 && (
												<TabsContent value="boxSets">
													<div className="divide-y divide-primary-dynamic/20 space-y-6">
														{paginatedBoxSets.slice(1).map((boxset) => (
															<div key={`${boxset.id}-boxset`} className="pb-6">
																<RenderBoxsetDisplay
																	includedItems={setIncludedItems || {}}
																	set={boxset}
																/>
															</div>
														))}
													</div>
												</TabsContent>
											)}
										</div>
									</Tabs>

									{/* Pagination */}
									<CustomPagination
										currentPage={currentPage}
										totalPages={totalPages}
										setCurrentPage={setCurrentPage}
										scrollToTop={true}
										filterItemsLength={
											activeTab === "boxSets"
												? filteredBoxsets.length
												: activeTab === "showSets"
													? filteredShowSets.length
													: activeTab === "movieSets"
														? filteredMovieSets.length
														: filteredCollectionSets.length
										}
										itemsPerPage={itemsPerPage}
									/>
								</div>
							))}
					</div>
				)}
		</div>
	);
};

export default UserSetPage;
