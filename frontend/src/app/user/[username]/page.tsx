"use client";

import {
	TMDBLookupMap,
	createTMDBLookupMap,
	getAllLibrarySectionsFromIDB,
	searchWithLookupMap,
} from "@/helper/searchIDBForTMDBID";
import { fetchAllUserSets } from "@/services/api.mediux";
import { ReturnErrorMessage } from "@/services/api.shared";

import { useEffect, useRef, useState } from "react";

import { useParams } from "next/navigation";

import { RenderBoxSetDisplay } from "@/components/shared/boxset-display";
import { CustomPagination } from "@/components/shared/custom-pagination";
import { ErrorMessage } from "@/components/shared/error-message";
import { SelectItemsPerPage } from "@/components/shared/items-per-page-select";
import Loader from "@/components/shared/loader";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ToggleGroup } from "@/components/ui/toggle-group";

import { usePageStore, useSearchQueryStore } from "@/lib/homeSearchStore";
import { log } from "@/lib/logger";
import { storage } from "@/lib/storage";

import { APIResponse } from "@/types/apiResponse";
import { MediaItem } from "@/types/mediaItem";
import {
	MediuxUserBoxset,
	MediuxUserCollectionMovie,
	MediuxUserCollectionSet,
	MediuxUserMovieSet,
	MediuxUserShowSet,
} from "@/types/mediuxUserAllSets";

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

const UserSetPage = () => {
	// Get the username from the URL
	const { username } = useParams();
	const hasFetchedInfo = useRef(false);
	// Error Handling
	const [error, setError] = useState<APIResponse<unknown> | null>(null);
	const [isLoading, setIsLoading] = useState(true);
	const [loadMessage, setLoadMessage] = useState("");
	const { itemsPerPage } = usePageStore();

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
	const [activeTab, setActiveTab] = useState("boxSets");
	const [currentPage, setCurrentPage] = useState(1);
	const [totalPages, setTotalPages] = useState(0);

	// Library sections & progress
	const [librarySections, setLibrarySections] = useState<{ title: string; type: string }[]>([]);
	const [selectedLibrarySection, setSelectedLibrarySection] = useState<{
		title: string;
		type: string;
	} | null>(null);
	const [filterOutInDB, setFilterOutInDB] = useState<"all" | "inDB" | "notInDB">("all");

	// Get all the library sections from the IDB
	useEffect(() => {
		const fetchLibrarySections = async () => {
			setLoadMessage("Loading library sections from cache");
			const sections = await getAllLibrarySectionsFromIDB();
			setLibrarySections(sections);
			log("User Page - Library Sections", sections);
		};
		fetchLibrarySections();
	}, []);

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
				log("User Page - Error fetching user sets:", error);
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
			log("No sets found in response");
			return;
		}
		if (!selectedLibrarySection) {
			log("No library section selected");
			return;
		}

		log("Filtering sets based on selected library", selectedLibrarySection);

		const filterOutItems = async () => {
			setIsLoading(true);

			// Get the library section data once
			const librarySection = await storage.getItem<{
				data: {
					MediaItems: MediaItem[];
				};
			}>(selectedLibrarySection.title);
			if (!librarySection || !librarySection.data?.MediaItems) {
				log(`User Page - No data found for library section: ${selectedLibrarySection.title}`);
				setIsLoading(false);
				return;
			}

			// Create a lookup map for faster access
			const tmdbLookupMap = createTMDBLookupMap(librarySection.data.MediaItems);
			log("User Page - TMDB Lookup Map", tmdbLookupMap);

			log("Setting items based on", selectedLibrarySection, filterOutInDB);

			// Process Boxsets
			if (respBoxsets && respBoxsets.length > 0) {
				const userBoxsets: MediuxUserBoxset[] = [];
				respBoxsets.map(async (origBoxset) => {
					const boxset = { ...origBoxset };
					if (selectedLibrarySection.type === "show") {
						boxset.movie_sets = []; // Clear movie sets since its not needed
						boxset.collection_sets = []; // Clear collection sets since its not needed
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
						boxset.show_sets = []; // Clear show sets since its not needed
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

					if (
						boxset.show_sets.length > 0 ||
						boxset.movie_sets.length > 0 ||
						boxset.collection_sets.length > 0
					) {
						userBoxsets.push(boxset);
					}
				});
				log("Processed Boxsets", userBoxsets);
				setBoxSets(userBoxsets);
				setIdbBoxsets(userBoxsets);
			}

			// Process Show Sets
			if (selectedLibrarySection.type === "show" && respShowSets && respShowSets.length > 0) {
				log("Processing Show Sets");
				const processedShowSets = await processBatch(respShowSets, tmdbLookupMap, "show", (current) =>
					setProgressCount((prev) => ({ ...prev, current }))
				);
				log("Processed Show Sets", processedShowSets);
				setShowSets(processedShowSets as MediuxUserShowSet[]);
				setIdbShowSets(processedShowSets as MediuxUserShowSet[]);
			}

			// Process Movie Sets
			if (selectedLibrarySection.type === "movie" && respMovieSets && respMovieSets.length > 0) {
				log("Processing Movie Sets");
				const processedMovieSets = await processBatch(respMovieSets, tmdbLookupMap, "movie", (current) =>
					setProgressCount((prev) => ({ ...prev, current }))
				);
				log("Processed Movie Sets", processedMovieSets);
				setMovieSets(processedMovieSets as MediuxUserMovieSet[]);
				setIdbMovieSets(processedMovieSets as MediuxUserMovieSet[]);
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
					log("Processed Collection Sets", filteredCollectionSets);
					setCollectionSets(filteredCollectionSets);
					setIdbCollectionSets(filteredCollectionSets);
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

			setBoxSets(filteredBoxsets);
		}

		// Filter show sets
		if (idbShowSets.length > 0) {
			log("Filtering Show Sets by Database Status", idbShowSets);
			const filteredShowSets = idbShowSets.filter((idbShowSets) => filterByDatabaseStatus(idbShowSets.show_id));
			setShowSets(filteredShowSets);
		}

		// Filter movie sets
		if (idbMovieSets.length > 0) {
			log("Filtering Movie Sets by Database Status", idbMovieSets);
			const filteredMovieSets = idbMovieSets.filter((idbMovieSet) =>
				filterByDatabaseStatus(idbMovieSet.movie_id)
			);
			setMovieSets(filteredMovieSets);
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

			setCollectionSets(filteredCollectionSets);
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [filterOutInDB]);

	// Update document title accordingly.
	useEffect(() => {
		if (error) {
			if (typeof window !== "undefined") document.title = "Aura | Error";
		} else {
			if (typeof window !== "undefined") document.title = `AURA | ${username} Sets`;
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
		log("User Page - Total Pages", totalPages);
		setCurrentPage(1); // Reset to first page when tab changes
	}, [activeTab, boxsets.length, collectionSets.length, itemsPerPage, movieSets.length, showSets.length, totalPages]);

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
							<h1 className="text-4xl font-extrabold text-center mb-2 tracking-tight text-primary">
								<span className="text-white opacity-80">Sets by</span>{" "}
								<span className="text-primary">{username}</span>
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
											className={`cursor-pointer text-sm ${
												!!selectedLibrarySection &&
												selectedLibrarySection.title !== section.title
													? "opacity-50 pointer-events-none"
													: ""
											}`}
										>
											{section.title}
										</Badge>
									))}
								</ToggleGroup>
							</div>
						</div>

						{/* Filter Out In DB Selection */}
						<div className="w-full flex items-center mb-2">
							<Label htmlFor="filter-out-in-db" className="text-lg font-semibold mr-2">
								Filter:
							</Label>
							{/* Filter Out In DB Toggle */}

							<Badge
								key="filter-out-in-db"
								className={`cursor-pointer text-sm ${
									filterOutInDB === "inDB"
										? "bg-green-600 text-white"
										: filterOutInDB === "notInDB"
											? "bg-red-600 text-white"
											: ""
								}`}
								variant={filterOutInDB !== "all" ? "default" : "outline"}
								onClick={() => {
									const next =
										filterOutInDB === "all" ? "inDB" : filterOutInDB === "inDB" ? "notInDB" : "all";
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

						{/* No library selected message */}
						{!selectedLibrarySection && (
							<div className="flex justify-center mt-8">
								<ErrorMessage error={ReturnErrorMessage<string>("No library selected")} />
							</div>
						)}

						{/* Content when library is selected */}
						{selectedLibrarySection &&
							(showSets.length === 0 &&
							movieSets.length === 0 &&
							collectionSets.length === 0 &&
							boxsets.length === 0 ? (
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
							) : (
								<div className="flex flex-col items-center mt-4 mb-4">
									{/* Items Per Page Selection */}
									<div className="w-full flex items-center mb-2">
										<SelectItemsPerPage setCurrentPage={setCurrentPage} />
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
													className="data-[state=active]:bg-background"
												>
													Show Sets ({showSets.length})
												</TabsTrigger>
											)}
											{movieSets.length > 0 && (
												<TabsTrigger
													value="movieSets"
													className="data-[state=active]:bg-background"
												>
													Movie Sets ({movieSets.length})
												</TabsTrigger>
											)}
											{collectionSets.length > 0 && (
												<TabsTrigger
													value="collectionSets"
													className="data-[state=active]:bg-background"
												>
													Collection Sets ({collectionSets.length})
												</TabsTrigger>
											)}
											{boxsets.length > 0 && (
												<TabsTrigger
													value="boxSets"
													className="data-[state=active]:bg-background"
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
													<div className="divide-y divide-primary-dynamic/20 space-y-6">
														{paginatedMovieSets.map((movieSet) => (
															<div key={`${movieSet.id}-movieset`} className="pb-6">
																<RenderBoxSetDisplay
																	key={movieSet.id}
																	set={movieSet}
																	type="movie"
																/>
															</div>
														))}
													</div>
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
									/>
								</div>
							))}
					</div>
				)}
		</div>
	);
};

export default UserSetPage;
