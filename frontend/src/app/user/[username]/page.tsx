"use client";

import { BoxsetMovieToPosterSet } from "@/helper/boxsets/boxset-to-movie-poster-set";
import { formatLastUpdatedDate } from "@/helper/format-date-last-updates";
import { TMDBLookupMap, createTMDBLookupMap, searchWithLookupMap } from "@/helper/search-idb-for-tmdb-id";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { fetchAllUserSets } from "@/services/mediux/api-mediux-fetch-username-sets";
import { ArrowDownAZ, ArrowDownZA, CircleAlert, ClockArrowDown, ClockArrowUp, Database, User } from "lucide-react";

import { useEffect, useRef, useState } from "react";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useParams } from "next/navigation";

import { AssetImage } from "@/components/shared/asset-image";
import { RenderBoxSetDisplay } from "@/components/shared/boxset-display";
import { CustomPagination } from "@/components/shared/custom-pagination";
import DownloadModal from "@/components/shared/download-modal";
import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { ResponsiveGrid } from "@/components/shared/responsive-grid";
import { SelectItemsPerPage } from "@/components/shared/select-items-per-page";
import { SortControl } from "@/components/shared/select-sort";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ToggleGroup } from "@/components/ui/toggle-group";
import { Lead, P } from "@/components/ui/typography";

import { cn } from "@/lib/cn";
import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";
import { useUserPageStore } from "@/lib/stores/page-store-user";

import { APIResponse } from "@/types/api/api-response";
import {
	MediuxUserBoxset,
	MediuxUserCollectionMovie,
	MediuxUserCollectionSet,
	MediuxUserMovieSet,
	MediuxUserShowSet,
} from "@/types/mediux/mediux-sets";

const processBatch = async <T extends MediuxUserMovieSet | MediuxUserShowSet | MediuxUserCollectionMovie>(
	items: MediuxUserMovieSet[] | MediuxUserShowSet[] | MediuxUserCollectionMovie[],
	lookupMap: TMDBLookupMap,
	type: "movie" | "show" | "collection",
	updateProgress: (current: number) => void
): Promise<T[]> => {
	const results: T[] = [];
	items.forEach((item, index) => {
		let tmdbId: string | undefined;

		// Handle different item types
		if (type === "movie" && "movie_id" in item) {
			tmdbId = item.movie_id.id;
		} else if (type === "collection" && "movie" in item) {
			tmdbId = item.movie.id;
		} else if (type === "show" && "show_id" in item) {
			tmdbId = item.show_id.id;
		}

		if (!tmdbId) return;

		const mediaItem = searchWithLookupMap(tmdbId, lookupMap);
		updateProgress(index + 1);

		if (mediaItem && mediaItem !== true) {
			// Handle different item types
			if (type === "movie" && "movie_id" in item) {
				item.movie_id.MediaItem = mediaItem;
			} else if (type === "show" && "show_id" in item) {
				item.show_id.MediaItem = mediaItem;
			} else if (type === "collection" && "movie" in item) {
				item.movie.MediaItem = mediaItem;
			}
			results.push(item as T);
		}
	});

	return results;
};

const processCollectionSetItems = async (
	set: MediuxUserCollectionSet,
	lookupMap: TMDBLookupMap,
	updateProgress: (current: number) => void
) => {
	const posters = await processBatch(set.movie_posters || [], lookupMap, "collection", updateProgress);

	const backdrops = await processBatch(set.movie_backdrops || [], lookupMap, "collection", updateProgress);

	if (posters.length !== 0 || backdrops.length !== 0) {
		return {
			...set,
			movie_posters: posters,
			movie_backdrops: backdrops,
		};
	}
};

function sortSets<T extends object>(
	sets: T[],
	sortOption: "title" | "dateLastUpdate",
	sortOrder: "asc" | "desc",
	titleKey: keyof T = "set_title" as keyof T // default to set_title, but can be boxset_title
): T[] {
	if (sortOption === "title") {
		return sets.slice().sort((a, b) => {
			const aTitle = (a[titleKey] as string) || "";
			const bTitle = (b[titleKey] as string) || "";
			return sortOrder === "asc" ? aTitle.localeCompare(bTitle) : bTitle.localeCompare(aTitle);
		});
	} else if (sortOption === "dateLastUpdate") {
		return sets.slice().sort((a, b) => {
			const aDate =
				"date_updated" in a &&
				(typeof a.date_updated === "string" ||
					typeof a.date_updated === "number" ||
					a.date_updated instanceof Date)
					? new Date(a.date_updated).getTime()
					: 0;
			const bDate =
				"date_updated" in b &&
				(typeof b.date_updated === "string" ||
					typeof b.date_updated === "number" ||
					b.date_updated instanceof Date)
					? new Date(b.date_updated).getTime()
					: 0;
			return sortOrder === "asc" ? aDate - bDate : bDate - aDate;
		});
	}
	return sets;
}

const UserSetPage = () => {
	const router = useRouter();
	// Get the username from the URL
	const { username } = useParams();
	const hasFetchedInfo = useRef(false);
	// Error Handling
	const [error, setError] = useState<APIResponse<unknown> | null>(null);
	const [isLoading, setIsLoading] = useState(true);
	const [loadMessage, setLoadMessage] = useState("");
	const { currentPage, setCurrentPage, itemsPerPage, setItemsPerPage } = useUserPageStore();

	// Add state to track progress
	const [, setProgressCount] = useState<{
		current: number;
		total: number;
	}>({
		current: 0,
		total: 0,
	});
	const [respShowSets, setRespShowSets] = useState<MediuxUserShowSet[]>([]);
	const [respMovieSets, setRespMovieSets] = useState<MediuxUserMovieSet[]>([]);
	const [respCollectionSets, setRespCollectionSets] = useState<MediuxUserCollectionSet[]>([]);
	const [respBoxsets, setRespBoxsets] = useState<MediuxUserBoxset[]>([]);
	const [idbShowSets, setIdbShowSets] = useState<MediuxUserShowSet[]>([]);
	const [idbMovieSets, setIdbMovieSets] = useState<MediuxUserMovieSet[]>([]);
	const [idbCollectionSets, setIdbCollectionSets] = useState<MediuxUserCollectionSet[]>([]);
	const [idbBoxsets, setIdbBoxsets] = useState<MediuxUserBoxset[]>([]);
	const [showSets, setShowSets] = useState<MediuxUserShowSet[]>([]);
	const [movieSets, setMovieSets] = useState<MediuxUserMovieSet[]>([]);
	const [collectionSets, setCollectionSets] = useState<MediuxUserCollectionSet[]>([]);
	const [boxsets, setBoxSets] = useState<MediuxUserBoxset[]>([]);

	const { searchQuery, setSearchQuery } = useSearchQueryStore();
	const prevSearchQuery = useRef(searchQuery);

	// Pagination and Active Tab state
	const [activeTab, setActiveTab] = useState("");
	const [totalPages, setTotalPages] = useState(0);

	// Library sections & progress
	const [librarySections, setLibrarySections] = useState<{ title: string; type: string }[]>([]);
	const [selectedLibrarySection, setSelectedLibrarySection] = useState<{
		title: string;
		type: string;
	} | null>(null);
	const [filterOutInDB, setFilterOutInDB] = useState<"all" | "inDB" | "notInDB">("all");

	// State to track the selected sorting option
	const { sortOption, setSortOption, sortOrder, setSortOrder } = useUserPageStore();

	const { sections, getSectionSummaries } = useLibrarySectionsStore();

	// Get all the library sections from the IDB
	useEffect(() => {
		const fetchLibrarySections = async () => {
			setLoadMessage("Loading library sections from cache");
			const sections = getSectionSummaries();
			setLibrarySections(sections);
			log("INFO", "User Page", "Library Sections", "Fetched library sections from cache", sections);
		};
		fetchLibrarySections();
	}, [getSectionSummaries]);

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
		const getAllUserSets = async () => {
			try {
				setIsLoading(true);
				setLoadMessage(`Loading sets for ${username}`);
				const response = await fetchAllUserSets(username as string);

				if (response.status === "error") {
					setError(response);
					return;
				}

				// Set the response data
				setRespShowSets(response.data?.show_sets || []);
				setRespMovieSets(response.data?.movie_sets || []);
				setRespCollectionSets(response.data?.collection_sets || []);
				setRespBoxsets(response.data?.boxsets || []);
			} catch (error) {
				log("ERROR", "User Page", "Fetch User Sets", "Failed to fetch user sets:", error);
				setError(ReturnErrorMessage<unknown>(error));
			} finally {
				setIsLoading(false);
			}
		};
		getAllUserSets();
	}, [username]);

	// Filter out the sets based on which Library type is selected
	useEffect(() => {
		setShowSets([]);
		setMovieSets([]);
		setCollectionSets([]);
		setBoxSets([]);
		setIdbShowSets([]);
		setIdbMovieSets([]);
		setIdbCollectionSets([]);
		setIdbBoxsets([]);
		setFilterOutInDB("all");

		if (!respShowSets || !respMovieSets || !respCollectionSets || !respBoxsets) {
			log("ERROR", "User Page", "Fetch User Sets", "No sets found in response");
			return;
		}
		if (!selectedLibrarySection) {
			log("WARN", "User Page", "Fetch User Sets", "No library section selected");
			return;
		}

		log("INFO", "User Page", "Fetch User Sets", "Filtering sets based on selected library", selectedLibrarySection);

		const filterOutItems = async () => {
			setIsLoading(true);

			// Get the library section data once
			const librarySection = sections[selectedLibrarySection.title];
			if (!librarySection || !librarySection.MediaItems) {
				log(
					"ERROR",
					"User Page",
					"Fetch User Sets",
					`No data found for library section: ${selectedLibrarySection.title}`
				);
				setIsLoading(false);
				return;
			}

			// Create a lookup map for faster access
			const tmdbLookupMap = createTMDBLookupMap(librarySection.MediaItems);
			log("INFO", "User Page", "Fetch User Sets", "TMDB Lookup Map", tmdbLookupMap);

			log("INFO", "User Page", "Fetch User Sets", "Setting items based on", {
				selectedLibrarySection,
				filterOutInDB,
			});

			// Process Boxsets
			if (respBoxsets && respBoxsets.length > 0) {
				setActiveTab("boxSets");
				const userBoxsets: MediuxUserBoxset[] = await Promise.all(
					respBoxsets.map(async (origBoxset) => {
						const boxset = { ...origBoxset };
						if (selectedLibrarySection.type === "show") {
							boxset.movie_sets = [];
							boxset.collection_sets = [];
							if (boxset.show_sets && boxset.show_sets.length > 0) {
								const processedShowSets = await processBatch(
									boxset.show_sets,
									tmdbLookupMap,
									"show",
									(current) => setProgressCount((prev) => ({ ...prev, current }))
								);
								boxset.show_sets = processedShowSets as MediuxUserShowSet[];
							}
						} else if (selectedLibrarySection.type === "movie") {
							boxset.show_sets = [];
							if (boxset.movie_sets && boxset.movie_sets.length > 0) {
								const processedMovieSets = await processBatch(
									boxset.movie_sets,
									tmdbLookupMap,
									"movie",
									(current) => setProgressCount((prev) => ({ ...prev, current }))
								);
								boxset.movie_sets = processedMovieSets as MediuxUserMovieSet[];
							}
							if (boxset.collection_sets && boxset.collection_sets.length > 0) {
								const processedCollectionSets = await Promise.all(
									boxset.collection_sets.map(async (set) => {
										return await processCollectionSetItems(set, tmdbLookupMap, (current) =>
											setProgressCount((prev) => ({ ...prev, current }))
										);
									})
								);
								if (processedCollectionSets && processedCollectionSets.length > 0) {
									const filteredCollectionSets = processedCollectionSets.filter(
										(set): set is MediuxUserCollectionSet => set !== undefined
									);
									boxset.collection_sets = filteredCollectionSets;
								} else {
									boxset.collection_sets = [];
								}
							}
						}

						return boxset;
					})
				);

				// Only keep boxsets with at least one set
				const filteredBoxsets = userBoxsets.filter(
					(boxset) =>
						boxset.show_sets.length > 0 || boxset.movie_sets.length > 0 || boxset.collection_sets.length > 0
				);

				// Sort Boxsets
				const sortedBoxsets = sortSets(
					filteredBoxsets,
					sortOption as "title" | "dateLastUpdate",
					sortOrder,
					"boxset_title"
				) as MediuxUserBoxset[];

				log("INFO", "User Page", "Fetch User Sets", "Processed Boxsets", sortedBoxsets);
				setBoxSets(sortedBoxsets);
				setIdbBoxsets(sortedBoxsets);
			}

			// Process Show Sets
			if (selectedLibrarySection.type === "show" && respShowSets && respShowSets.length > 0) {
				setActiveTab("showSets");
				log("INFO", "User Page", "Fetch User Sets", "Processing Show Sets");
				const processedShowSets = await processBatch<MediuxUserShowSet>(
					respShowSets,
					tmdbLookupMap,
					"show",
					(current) => setProgressCount((prev) => ({ ...prev, current }))
				);

				// Sort Show Sets
				const sortedShowSets = sortSets(
					processedShowSets,
					sortOption as "title" | "dateLastUpdate",
					sortOrder as "asc" | "desc"
				);

				log("INFO", "User Page", "Fetch User Sets", "Processed Show Sets", sortedShowSets);
				setShowSets(sortedShowSets);
				setIdbShowSets(sortedShowSets);
			}

			// Process Movie Sets
			if (selectedLibrarySection.type === "movie" && respMovieSets && respMovieSets.length > 0) {
				setActiveTab("movieSets");
				log("INFO", "User Page", "Fetch User Sets", "Processing Movie Sets");
				const processedMovieSets = await processBatch<MediuxUserMovieSet>(
					respMovieSets,
					tmdbLookupMap,
					"movie",
					(current) => setProgressCount((prev) => ({ ...prev, current }))
				);

				// Sort Movie Sets
				const sortedMovieSets = sortSets(
					processedMovieSets,
					sortOption as "title" | "dateLastUpdate",
					sortOrder as "asc" | "desc"
				);

				log("INFO", "User Page", "Fetch User Sets", "Processed Movie Sets", sortedMovieSets);
				setMovieSets(sortedMovieSets);
				setIdbMovieSets(sortedMovieSets);
			}

			// Process Collection Sets
			if (selectedLibrarySection.type === "movie" && respCollectionSets && respCollectionSets.length > 0) {
				const processedCollectionSets = await Promise.all(
					respCollectionSets.map(async (set) => {
						return await processCollectionSetItems(set, tmdbLookupMap, (current) =>
							setProgressCount((prev) => ({ ...prev, current }))
						);
					})
				);
				if (processedCollectionSets && processedCollectionSets.length > 0) {
					const filteredCollectionSets = processedCollectionSets.filter(
						(set): set is MediuxUserCollectionSet => set !== undefined
					);

					// Sort Collection Sets
					const sortedCollectionSets = sortSets(
						filteredCollectionSets,
						sortOption as "title" | "dateLastUpdate",
						sortOrder as "asc" | "desc"
					);

					log("INFO", "User Page", "Fetch User Sets", "Processed Collection Sets", sortedCollectionSets);
					setCollectionSets(sortedCollectionSets);
					setIdbCollectionSets(sortedCollectionSets);
				}
			}

			setIsLoading(false);
		};

		filterOutItems();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [selectedLibrarySection]);

	useEffect(() => {
		if (searchQuery !== prevSearchQuery.current) {
			setCurrentPage(1);
			prevSearchQuery.current = searchQuery;

			if (!selectedLibrarySection) {
				return;
			}
			// Filter out the Show Sets on set title and mediaItem.Title
			const filteredShowSets = idbShowSets.filter(
				(showSet) =>
					showSet.set_title.toLowerCase().includes(searchQuery.toLowerCase()) ||
					showSet.show_id.MediaItem.Title.toLowerCase().includes(searchQuery.toLowerCase())
			);
			setShowSets(filteredShowSets);

			// Filter out the Movie Sets on set title and mediaItem.Title
			const filteredMovieSets = idbMovieSets.filter(
				(movieSet) =>
					movieSet.set_title.toLowerCase().includes(searchQuery.toLowerCase()) ||
					movieSet.movie_id.MediaItem.Title.toLowerCase().includes(searchQuery.toLowerCase())
			);
			setMovieSets(filteredMovieSets);

			// Filter out the Collection Sets on set title and mediaItem.Title
			const filteredCollectionSets = idbCollectionSets.filter(
				(collectionSet) =>
					collectionSet.set_title.toLowerCase().includes(searchQuery.toLowerCase()) ||
					collectionSet.movie_posters.some((poster) =>
						poster.movie.MediaItem.Title.toLowerCase().includes(searchQuery.toLowerCase())
					) ||
					collectionSet.movie_backdrops.some((backdrop) =>
						backdrop.movie.MediaItem.Title.toLowerCase().includes(searchQuery.toLowerCase())
					)
			);
			setCollectionSets(filteredCollectionSets);

			// Filter out the Box Sets on set title and mediaItem.Title
			const filteredBoxSets = idbBoxsets.filter((boxset) => {
				const boxsetTitleMatch = boxset.boxset_title.toLowerCase().includes(searchQuery.toLowerCase());
				const showSetsMatch = boxset.show_sets.some(
					(showSet) =>
						showSet.set_title.toLowerCase().includes(searchQuery.toLowerCase()) ||
						showSet.show_id.MediaItem.Title.toLowerCase().includes(searchQuery.toLowerCase())
				);
				const movieSetsMatch = boxset.movie_sets.some(
					(movieSet) =>
						movieSet.set_title.toLowerCase().includes(searchQuery.toLowerCase()) ||
						movieSet.movie_id.MediaItem.Title.toLowerCase().includes(searchQuery.toLowerCase())
				);
				const collectionSetsMatch = boxset.collection_sets.some(
					(collectionSet) =>
						collectionSet.set_title.toLowerCase().includes(searchQuery.toLowerCase()) ||
						collectionSet.movie_posters.some((poster) =>
							poster.movie.MediaItem.Title.toLowerCase().includes(searchQuery.toLowerCase())
						) ||
						collectionSet.movie_backdrops.some((backdrop) =>
							backdrop.movie.MediaItem.Title.toLowerCase().includes(searchQuery.toLowerCase())
						)
				);
				const boxsetMatches = boxsetTitleMatch || showSetsMatch || movieSetsMatch || collectionSetsMatch;
				return boxsetMatches;
			});
			setBoxSets(filteredBoxSets);
		}
	}, [idbBoxsets, idbCollectionSets, idbMovieSets, idbShowSets, searchQuery, selectedLibrarySection, setCurrentPage]);

	// Add this effect to handle filterOutInDB changes
	useEffect(() => {
		if (!selectedLibrarySection) return;

		const filterByDatabaseStatus = (item: { MediaItem: { ExistInDatabase: boolean } }) => {
			if (filterOutInDB === "all") return true;
			if (filterOutInDB === "inDB") return item.MediaItem?.ExistInDatabase;
			if (filterOutInDB === "notInDB") return !item.MediaItem?.ExistInDatabase;
			return false;
		};
		// Filter boxsets
		if (idbBoxsets.length > 0) {
			const filteredBoxsets = idbBoxsets
				.map((boxset) => {
					const newBoxset = { ...boxset };
					if (selectedLibrarySection.type === "show") {
						newBoxset.show_sets = boxset.show_sets.filter((showSet) =>
							filterByDatabaseStatus(showSet.show_id)
						);
					} else if (selectedLibrarySection.type === "movie") {
						newBoxset.movie_sets = boxset.movie_sets.filter((movieSet) =>
							filterByDatabaseStatus(movieSet.movie_id)
						);
						newBoxset.collection_sets = boxset.collection_sets
							.map((collectionSet) => ({
								...collectionSet,
								movie_posters: collectionSet.movie_posters.filter((poster) =>
									filterByDatabaseStatus(poster.movie)
								),
								movie_backdrops: collectionSet.movie_backdrops.filter((backdrop) =>
									filterByDatabaseStatus(backdrop.movie)
								),
							}))
							.filter((set) => set.movie_posters.length > 0 || set.movie_backdrops.length > 0);
					}

					return newBoxset;
				})
				.filter(
					(boxset) =>
						boxset.movie_sets?.length > 0 ||
						boxset.show_sets?.length > 0 ||
						boxset.collection_sets?.length > 0
				);

			const sortedBoxsets = sortSets(
				filteredBoxsets,
				sortOption as "title" | "dateLastUpdate",
				sortOrder,
				"boxset_title"
			);

			setBoxSets(sortedBoxsets);
		}

		// Filter show sets
		if (idbShowSets.length > 0) {
			log("INFO", "User Page", "Fetch User Sets", "Filtering Show Sets by Database Status", idbShowSets);
			const filteredShowSets = idbShowSets.filter((idbShowSets) => filterByDatabaseStatus(idbShowSets.show_id));
			const sortedShowSets = sortSets<MediuxUserShowSet>(
				filteredShowSets,
				sortOption as "dateLastUpdate" | "title",
				sortOrder
			);
			setShowSets(sortedShowSets);
		}

		// Filter movie sets
		if (idbMovieSets.length > 0) {
			log("INFO", "User Page", "Fetch User Sets", "Filtering Movie Sets by Database Status", idbMovieSets);
			const filteredMovieSets = idbMovieSets.filter((idbMovieSet) =>
				filterByDatabaseStatus(idbMovieSet.movie_id)
			);
			const sortedMovieSets = sortSets<MediuxUserMovieSet>(
				filteredMovieSets,
				sortOption as "dateLastUpdate" | "title",
				sortOrder
			);
			setMovieSets(sortedMovieSets);
		}

		// Filter collection sets
		if (idbCollectionSets.length > 0) {
			const filteredCollectionSets = idbCollectionSets
				.map((collectionSet) => ({
					...collectionSet,
					movie_posters: collectionSet.movie_posters.filter((poster) => filterByDatabaseStatus(poster.movie)),
					movie_backdrops: collectionSet.movie_backdrops.filter((backdrop) =>
						filterByDatabaseStatus(backdrop.movie)
					),
				}))
				.filter((set) => set.movie_posters.length > 0 || set.movie_backdrops.length > 0);
			const sortedCollectionSets = sortSets<MediuxUserCollectionSet>(
				filteredCollectionSets,
				sortOption as "dateLastUpdate" | "title",
				sortOrder
			);
			setCollectionSets(sortedCollectionSets);
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [filterOutInDB, sortOrder, sortOption]);

	// Update document title accordingly.
	useEffect(() => {
		if (error) {
			if (typeof window !== "undefined") document.title = "aura | Error";
		} else {
			if (typeof window !== "undefined") document.title = `aura | ${username} Sets`;
		}
	}, [error, username]);

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

	const paginatedShowSets = showSets.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);
	const paginatedMovieSets = movieSets.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);
	const paginatedCollectionSets = collectionSets.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);
	const paginatedBoxSets = boxsets.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

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
				respShowSets.length === 0 &&
				respMovieSets.length === 0 &&
				respCollectionSets.length === 0 &&
				respBoxsets.length === 0 && (
					<div className="flex justify-center">
						<ErrorMessage error={ReturnErrorMessage<string>(`No sets found for user ${username}.`)} />
					</div>
				)}

			{/* Main content when sets exist */}
			{!isLoading &&
				!error &&
				(respShowSets.length > 0 ||
					respMovieSets.length > 0 ||
					respCollectionSets.length > 0 ||
					respBoxsets.length > 0) && (
					<div className="min-h-screen px-4 sm:px-8 pb-20">
						{/* User Sets Header */}
						<div className="flex flex-col items-center mt-8 mb-6">
							<h1 className="text-4xl font-extrabold text-center mb-2 tracking-tight text-primary flex items-center justify-center gap-2">
								<span className="text-white opacity-80">Sets by</span>
								<span className="text-primary">{username}</span>
								<Avatar className="rounded-lg w-7 h-7 ml-2 align-middle">
									<AvatarImage
										src={`/api/mediux/avatar-image?username=${username}`}
										className="w-7 h-7"
									/>
									<AvatarFallback>
										<User className="w-7 h-7" />
									</AvatarFallback>
								</Avatar>
							</h1>
							{!selectedLibrarySection && (
								<div className="flex flex-wrap gap-3 mt-2 justify-center">
									{respShowSets.length > 0 && (
										<div className="flex items-center gap-2 bg-background border border-border rounded-lg px-4 py-2 shadow-sm">
											<span className="font-semibold text-primary">Show Sets</span>
											<Badge variant="secondary" className="text-xs px-2 py-1">
												{respShowSets.length}
											</Badge>
										</div>
									)}
									{respMovieSets.length > 0 && (
										<div className="flex items-center gap-2 bg-background border border-border rounded-lg px-4 py-2 shadow-sm">
											<span className="font-semibold text-primary">Movie Sets</span>
											<Badge variant="secondary" className="text-xs px-2 py-1">
												{respMovieSets.length}
											</Badge>
										</div>
									)}
									{respCollectionSets.length > 0 && (
										<div className="flex items-center gap-2 bg-background border border-border rounded-lg px-4 py-2 shadow-sm">
											<span className="font-semibold text-primary">Collection Sets</span>
											<Badge variant="secondary" className="text-xs px-2 py-1">
												{respCollectionSets.length}
											</Badge>
										</div>
									)}
									{respBoxsets.length > 0 && (
										<div className="flex items-center gap-2 bg-background border border-border rounded-lg px-4 py-2 shadow-sm">
											<span className="font-semibold text-primary">Box Sets</span>
											<Badge variant="secondary" className="text-xs px-2 py-1">
												{respBoxsets.length}
											</Badge>
										</div>
									)}
								</div>
							)}
						</div>

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
										setSelectedLibrarySection(
											val
												? librarySections.find((section) => section.title === val) || null
												: null
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
												if (selectedLibrarySection?.title === section.title) {
													setSelectedLibrarySection(null);
													setCurrentPage(1);
													setFilterOutInDB("all");
												} else {
													setSelectedLibrarySection(section);
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

						{/* Content when library is selected */}
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
																<RenderBoxSetDisplay
																	key={showSet.id}
																	set={showSet}
																	type="show"
																/>
															</div>
														))}
													</div>
												</TabsContent>
											)}

											{paginatedMovieSets.length > 0 && (
												<TabsContent value="movieSets">
													<ResponsiveGrid size="regular">
														{paginatedMovieSets.map((set) => {
															const posterSets = BoxsetMovieToPosterSet(
																set as MediuxUserMovieSet
															);
															const posterSet = posterSets[0];
															return (
																<div
																	key={posterSet.ID}
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
																				setType={"movie"}
																				setTitle={posterSet.Title}
																				setID={posterSet.ID}
																				setAuthor={posterSet.User.Name}
																				posterSets={posterSets}
																			/>
																		</div>
																		{/* Set Name */}
																		<P className="text-primary-dynamic text-sm font-semibold w-full mb-1 truncate pr-10">
																			{posterSet.Title}
																		</P>
																	</div>

																	{/* Set User Name */}
																	<div className="flex items-center justify-start w-full mb-1">
																		<div className="flex items-center gap-1">
																			<Avatar className="rounded-lg mr-1 w-4 h-4">
																				<AvatarImage
																					src={`/api/mediux/avatar-image?username=${posterSet.User.Name}`}
																					className="w-4 h-4"
																				/>
																				<AvatarFallback className="">
																					<User className="w-4 h-4" />
																				</AvatarFallback>
																			</Avatar>
																			<Link
																				href={`/user/${posterSet.User.Name}`}
																				className="text-sm hover:text-primary cursor-pointer underline truncate"
																				style={{ wordBreak: "break-word" }}
																			>
																				{posterSet.User.Name}
																			</Link>
																		</div>
																	</div>

																	{/* Last Update */}
																	<>
																		<Lead className="text-sm text-muted-foreground w-full mb-2 flex items-center gap-2">
																			Last Update:{" "}
																			{formatLastUpdatedDate(
																				posterSet.Poster?.Modified || "",
																				posterSet.Backdrop?.Modified || ""
																			)}
																			{set.movie_id.MediaItem?.DBSavedSets &&
																				set.movie_id.MediaItem?.DBSavedSets
																					.length > 0 && (
																					<Popover>
																						<PopoverTrigger asChild>
																							<Database
																								className={cn(
																									"cursor-pointer active:scale-95",
																									(
																										set as MediuxUserMovieSet
																									).movie_id.MediaItem?.DBSavedSets?.some(
																										(dbSet) =>
																											dbSet.PosterSetID ===
																											posterSet.ID
																									)
																										? "text-green-500"
																										: "text-yellow-500"
																								)}
																								size={16}
																							/>
																						</PopoverTrigger>
																						<PopoverContent
																							side="top"
																							sideOffset={5}
																							className="bg-secondary border border-2 border-primary p-2"
																						>
																							<div className="flex items-center mb-2">
																								<CircleAlert className="h-5 w-5 text-yellow-500 mr-2" />
																								<span className="text-sm text-yellow-600">
																									This media item
																									already exists in
																									your database
																								</span>
																							</div>
																							<div className="text-xs text-muted-foreground mb-2">
																								You have previously
																								saved it in the
																								following sets
																							</div>
																							<ul className="space-y-2">
																								{set.movie_id.MediaItem.DBSavedSets.map(
																									(dbSet) => (
																										<li
																											key={
																												dbSet.PosterSetID
																											}
																											className="flex items-center rounded-md px-2 py-1 shadow-sm"
																										>
																											<Button
																												variant="outline"
																												className={cn(
																													"flex items-center transition-colors rounded-md px-2 py-1 cursor-pointer text-sm",
																													dbSet.PosterSetID.toString() ===
																														posterSet.ID.toString()
																														? "text-green-600  hover:bg-green-100  hover:text-green-600"
																														: "text-yellow-600 hover:bg-yellow-100 hover:text-yellow-700"
																												)}
																												aria-label={`View saved set ${dbSet.PosterSetID} ${dbSet.PosterSetUser ? `by ${dbSet.PosterSetUser}` : ""}`}
																												onClick={(
																													e
																												) => {
																													e.stopPropagation();
																													setSearchQuery(
																														`${set.movie_id.MediaItem.Title} Y:${set.movie_id.MediaItem.Year}: ID:${set.movie_id.MediaItem.TMDB_ID}: L:${set.movie_id.MediaItem.LibraryTitle}:`
																													);
																													router.push(
																														"/saved-sets"
																													);
																												}}
																											>
																												Set ID:{" "}
																												{
																													dbSet.PosterSetID
																												}
																												{dbSet.PosterSetUser
																													? ` by ${dbSet.PosterSetUser}`
																													: ""}
																											</Button>
																										</li>
																									)
																								)}
																							</ul>
																						</PopoverContent>
																					</Popover>
																				)}
																		</Lead>
																	</>

																	{/* Poster */}
																	{posterSet.Poster && (
																		<AssetImage
																			image={posterSet.Poster}
																			aspect="poster"
																			className="w-full mb-2"
																		/>
																	)}

																	{/* Backdrop */}
																	{posterSet.Backdrop && (
																		<AssetImage
																			image={posterSet.Backdrop}
																			aspect="backdrop"
																			className="w-full"
																		/>
																	)}
																</div>
															);
														})}
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
																<RenderBoxSetDisplay
																	key={collectionSet.id}
																	set={collectionSet}
																	type="collection"
																/>
															</div>
														))}
													</div>
												</TabsContent>
											)}

											{paginatedBoxSets.length > 0 && (
												<TabsContent value="boxSets">
													<div className="divide-y divide-primary-dynamic/20 space-y-6">
														{paginatedBoxSets.map((boxset) => (
															<div key={`${boxset.id}-boxset`} className="pb-6">
																<RenderBoxSetDisplay
																	key={boxset.id}
																	set={boxset}
																	type="boxset"
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
												? boxsets.length
												: activeTab === "showSets"
													? showSets.length
													: activeTab === "movieSets"
														? movieSets.length
														: collectionSets.length
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
