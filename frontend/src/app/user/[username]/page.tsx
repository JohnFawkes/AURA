"use client";

import { CarouselShow } from "@/components/carousel-show";
import {
	Carousel,
	CarouselContent,
	CarouselNext,
	CarouselPrevious,
} from "@/components/ui/carousel";
import ErrorMessage from "@/components/ui/error-message";
import Loader from "@/components/ui/loader";
import { Lead } from "@/components/ui/typography";
import { formatLastUpdatedDate } from "@/helper/formatDate";
import {
	getAllLibrarySectionsFromIDB,
	searchIDBForTMDBID,
} from "@/helper/searchIDBForTMDBID";
import { log } from "@/lib/logger";
import { fetchAllUserSets } from "@/services/api.mediux";
import {
	MediuxUserBoxset,
	MediuxUserCollectionMovie,
	MediuxUserCollectionSet,
	MediuxUserMovieSet,
	MediuxUserShowSet,
} from "@/types/mediuxUserAllSets";
import { PosterSet } from "@/types/posterSets";
import { useParams } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { CarouselMovie } from "@/components/carousel-movie";
import DownloadModalShow from "@/components/download-modal-show";
import DownloadModalMovie from "@/components/download-modal-movie";
import { useRouter } from "next/navigation";
import { usePosterSetStore } from "@/lib/posterSetStore";
import { useMediaStore } from "@/lib/mediaStore";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
	Pagination,
	PaginationContent,
	PaginationItem,
	PaginationLink,
	PaginationPrevious,
	PaginationNext,
	PaginationEllipsis,
} from "@/components/ui/pagination";
import {
	Accordion,
	AccordionItem,
	AccordionTrigger,
	AccordionContent,
} from "@/components/ui/accordion";
import { Label } from "@/components/ui/label";
import { ToggleGroup } from "@/components/ui/toggle-group";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2 as Checkmark } from "lucide-react";
import { BoxsetDisplay } from "@/components/boxset-display";

const itemsPerPage = 10; // adjust as needed

const UserSetPage = () => {
	// Get the username from the URL
	const { username } = useParams();
	const hasFetchedInfo = useRef(false);
	// Error Handling
	const [hasError, setHasError] = useState(false);
	const [errorMessage, setErrorMessage] = useState("");
	const [isLoading, setIsLoading] = useState(true);
	const [loadMessage, setLoadMessage] = useState("");

	const [respShowSets, setRespShowSets] = useState<MediuxUserShowSet[]>([]);
	const [respMovieSets, setRespMovieSets] = useState<MediuxUserMovieSet[]>(
		[]
	);
	const [respCollectionSets, setRespCollectionSets] = useState<
		MediuxUserCollectionSet[]
	>([]);
	const [respBoxsets, setRespBoxsets] = useState<MediuxUserBoxset[]>([]);
	const [idbShowSets, setIdbShowSets] = useState<MediuxUserShowSet[]>([]);
	const [idbMovieSets, setIdbMovieSets] = useState<MediuxUserMovieSet[]>([]);
	const [idbCollectionSets, setIdbCollectionSets] = useState<
		MediuxUserCollectionSet[]
	>([]);
	const [idbBoxsets, setIdbBoxsets] = useState<MediuxUserBoxset[]>([]);
	const [showSets, setShowSets] = useState<MediuxUserShowSet[]>([]);
	const [movieSets, setMovieSets] = useState<MediuxUserMovieSet[]>([]);
	const [collectionSets, setCollectionSets] = useState<
		MediuxUserCollectionSet[]
	>([]);
	const [boxsets, setBoxSets] = useState<MediuxUserBoxset[]>([]);

	// Pagination and Active Tab state
	const [activeTab, setActiveTab] = useState("boxSets"); // Default to boxSets
	const [currentPage, setCurrentPage] = useState(1);
	const [totalPages, setTotalPages] = useState(0);

	// Library sections & progress
	const [librarySections, setLibrarySections] = useState<
		{ title: string; type: string }[]
	>([]);
	const [selectedLibrarySection, setSelectedLibrarySection] = useState<{
		title: string;
		type: string;
	} | null>(null);
	const [filterOutInDB, setFilterOutInDB] = useState<
		"all" | "inDB" | "notInDB"
	>("all");

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
		// Perform any side effects or data fetching here
		log(`Fetching data for user: ${username}`);
		setLoadMessage(`Loading sets for ${username}`);
		const getAllUserSets = async () => {
			try {
				setIsLoading(true);
				const resp = await fetchAllUserSets(username as string);
				if (!resp) {
					throw new Error("No response from server");
				}
				log("User Page - Response from fetchAllUserSets", resp);

				setRespShowSets(resp.data?.show_sets || []);
				setRespMovieSets(resp.data?.movie_sets || []);
				setRespCollectionSets(resp.data?.collection_sets || []);
				setRespBoxsets(resp.data?.boxsets || []);
				setIsLoading(false);
			} catch (error) {
				log("User Page - Error fetching user sets", error);
				setHasError(true);
				if (error instanceof Error) {
					setErrorMessage(error.message);
				} else {
					setErrorMessage("An unknown error occurred");
				}
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

		setIsLoading(true);
		setLoadMessage("Filtering sets based on selected library");
		const filterOutItems = async () => {
			if (!selectedLibrarySection) {
				setIsLoading(false);
				return;
			}
			log(
				"Setting items based on",
				selectedLibrarySection,
				filterOutInDB
			);
			if (respBoxsets && respBoxsets.length > 0) {
				setLoadMessage("Filtering Boxsets");
				const userBoxsets: MediuxUserBoxset[] = [];
				log("User Page - Filtering Boxsets", respBoxsets);
				await Promise.all(
					respBoxsets.map(async (origBoxset) => {
						const boxset = JSON.parse(JSON.stringify(origBoxset));
						if (selectedLibrarySection.type === "show") {
							boxset.movie_sets = []; // Clear movie sets since its not needed
							boxset.collection_sets = []; // Clear collection sets since its not needed
							if (
								boxset.show_sets &&
								boxset.show_sets.length > 0
							) {
								setLoadMessage(
									"Filtering Show Sets in Boxsets"
								);
								const updatedShowSets: MediuxUserShowSet[] = [];
								await Promise.all(
									boxset.show_sets.map(
										async (showSet: MediuxUserShowSet) => {
											const dbResult =
												await searchIDBForTMDBID(
													showSet.show_id.id,
													selectedLibrarySection.title
												);
											console.log(
												"SHOW RESULT",
												dbResult,
												showSet.set_title
											);
											if (dbResult && dbResult !== true) {
												showSet.MediaItem = dbResult;
												if (filterOutInDB === "all") {
													updatedShowSets.push(
														showSet
													);
												} else if (
													filterOutInDB === "inDB" &&
													dbResult.ExistInDatabase
												) {
													updatedShowSets.push(
														showSet
													);
												} else if (
													filterOutInDB ===
														"notInDB" &&
													!dbResult.ExistInDatabase
												) {
													updatedShowSets.push(
														showSet
													);
												}
											}
										}
									)
								);
								boxset.show_sets = updatedShowSets;
							}
						} else if (selectedLibrarySection.type === "movie") {
							boxset.show_sets = []; // Clear show sets since its not needed
							// Process boxset.movie_sets if they exist.
							if (
								boxset.movie_sets &&
								boxset.movie_sets.length > 0
							) {
								setLoadMessage(
									"Filtering Movie Sets in Boxsets"
								);
								const updatedMovieSets: MediuxUserMovieSet[] =
									[];
								await Promise.all(
									boxset.movie_sets.map(
										async (
											movieSet: MediuxUserMovieSet
										) => {
											const dbResult =
												await searchIDBForTMDBID(
													movieSet.movie_id.id,
													selectedLibrarySection.title
												);
											if (dbResult && dbResult !== true) {
												movieSet.MediaItem = dbResult;
												if (filterOutInDB === "all") {
													updatedMovieSets.push(
														movieSet
													);
												} else if (
													filterOutInDB === "inDB" &&
													dbResult.ExistInDatabase
												) {
													updatedMovieSets.push(
														movieSet
													);
												} else if (
													filterOutInDB ===
														"notInDB" &&
													!dbResult.ExistInDatabase
												) {
													updatedMovieSets.push(
														movieSet
													);
												}
											}
										}
									)
								);
								boxset.movie_sets = updatedMovieSets;
							}

							if (
								boxset.collection_sets &&
								boxset.collection_sets.length > 0
							) {
								const updatedCollectionSets: MediuxUserCollectionSet[] =
									[];
								await Promise.all(
									boxset.collection_sets.map(
										async (
											collectionSet: MediuxUserCollectionSet
										) => {
											const posterMovies: MediuxUserCollectionMovie[] =
												[];
											await Promise.all(
												collectionSet.movie_posters.map(
													async (poster) => {
														const dbResult =
															await searchIDBForTMDBID(
																poster.movie.id,
																selectedLibrarySection.title
															);
														if (
															dbResult &&
															dbResult !== true
														) {
															poster.movie.MediaItem =
																dbResult;
															if (
																filterOutInDB ===
																"all"
															) {
																posterMovies.push(
																	poster
																);
															} else if (
																filterOutInDB ===
																	"inDB" &&
																dbResult.ExistInDatabase
															) {
																posterMovies.push(
																	poster
																);
															} else if (
																filterOutInDB ===
																	"notInDB" &&
																!dbResult.ExistInDatabase
															) {
																posterMovies.push(
																	poster
																);
															}
														}
													}
												)
											);
											const backdropMovies: MediuxUserCollectionMovie[] =
												[];
											await Promise.all(
												collectionSet.movie_backdrops.map(
													async (backdrop) => {
														const dbResult =
															await searchIDBForTMDBID(
																backdrop.movie
																	.id,
																selectedLibrarySection.title
															);
														if (
															dbResult &&
															dbResult !== true
														) {
															backdrop.movie.MediaItem =
																dbResult;
															if (
																filterOutInDB ===
																"all"
															) {
																backdropMovies.push(
																	backdrop
																);
															} else if (
																filterOutInDB ===
																	"inDB" &&
																dbResult.ExistInDatabase
															) {
																backdropMovies.push(
																	backdrop
																);
															} else if (
																filterOutInDB ===
																	"notInDB" &&
																!dbResult.ExistInDatabase
															) {
																backdropMovies.push(
																	backdrop
																);
															}
														}
													}
												)
											);
											// Only include the collectionSet if at least one poster or backdrop was found.
											if (
												posterMovies.length > 0 ||
												backdropMovies.length > 0
											) {
												updatedCollectionSets.push({
													...collectionSet,
													movie_posters: posterMovies,
													movie_backdrops:
														backdropMovies,
												});
											}
										}
									)
								);
								boxset.collection_sets = updatedCollectionSets;
							}
						}
						if (
							boxset.movie_sets.length > 0 ||
							boxset.show_sets.length > 0 ||
							boxset.collection_sets.length > 0
						) {
							userBoxsets.push(boxset);
						}
					})
				);
				log("User Page - Filtered Box Sets", userBoxsets);
				setBoxSets(userBoxsets);
				setIdbBoxsets(userBoxsets);
			}

			// Handle the show sets
			if (
				respShowSets &&
				respShowSets.length > 0 &&
				selectedLibrarySection.type === "show"
			) {
				const userShowSets: MediuxUserShowSet[] = [];
				log("User Page - Filtering Show Sets", respShowSets);
				await Promise.all(
					respShowSets.map(async (showSet) => {
						const dbResult = await searchIDBForTMDBID(
							showSet.show_id.id,
							selectedLibrarySection.title
						);
						if (dbResult && dbResult !== true) {
							showSet.MediaItem = dbResult;
							if (filterOutInDB === "all") {
								userShowSets.push(showSet);
							} else if (
								filterOutInDB === "inDB" &&
								dbResult.ExistInDatabase
							) {
								userShowSets.push(showSet);
							} else if (
								filterOutInDB === "notInDB" &&
								!dbResult.ExistInDatabase
							) {
								userShowSets.push(showSet);
							}
						}
					})
				);
				log("User Page - Filtered Show Sets", userShowSets);
				setShowSets(userShowSets);
				setIdbShowSets(userShowSets);
			}

			// Handle the movie sets
			if (
				respMovieSets &&
				respMovieSets.length > 0 &&
				selectedLibrarySection.type === "movie"
			) {
				const userMovieSets: MediuxUserMovieSet[] = [];
				log("User Page - Filtering Movie Sets", respMovieSets);
				await Promise.all(
					respMovieSets.map(async (movieSet) => {
						const dbResult = await searchIDBForTMDBID(
							movieSet.movie_id.id,
							selectedLibrarySection.title
						);
						if (dbResult && dbResult !== true) {
							movieSet.MediaItem = dbResult;
							if (filterOutInDB === "all") {
								userMovieSets.push(movieSet);
							} else if (
								filterOutInDB === "inDB" &&
								dbResult.ExistInDatabase
							) {
								userMovieSets.push(movieSet);
							} else if (
								filterOutInDB === "notInDB" &&
								!dbResult.ExistInDatabase
							) {
								userMovieSets.push(movieSet);
							}
						}
					})
				);
				log("User Page - Filtered Movie Sets", userMovieSets);
				setMovieSets(userMovieSets);
				setIdbMovieSets(userMovieSets);
			}

			// Handle the collection sets
			if (
				respCollectionSets &&
				respCollectionSets.length > 0 &&
				selectedLibrarySection.type === "movie"
			) {
				const userCollectionSets: MediuxUserCollectionSet[] = [];
				log(
					"User Page - Filtering Collection Sets",
					respCollectionSets
				);
				await Promise.all(
					respCollectionSets.map(async (collectionSet) => {
						const posterMovies: MediuxUserCollectionMovie[] = [];
						await Promise.all(
							collectionSet.movie_posters.map(async (poster) => {
								const dbResult = await searchIDBForTMDBID(
									poster.movie.id,
									selectedLibrarySection.title
								);
								if (dbResult && dbResult !== true) {
									poster.movie.MediaItem = dbResult;
									if (filterOutInDB === "all") {
										posterMovies.push(poster);
									} else if (
										filterOutInDB === "inDB" &&
										dbResult.ExistInDatabase
									) {
										posterMovies.push(poster);
									} else if (
										filterOutInDB === "notInDB" &&
										!dbResult.ExistInDatabase
									) {
										posterMovies.push(poster);
									}
								}
							})
						);
						const backdropMovies: MediuxUserCollectionMovie[] = [];
						await Promise.all(
							collectionSet.movie_backdrops.map(
								async (backdrop) => {
									const dbResult = await searchIDBForTMDBID(
										backdrop.movie.id,
										selectedLibrarySection.title
									);
									if (dbResult && dbResult !== true) {
										backdrop.movie.MediaItem = dbResult;
										if (filterOutInDB === "all") {
											backdropMovies.push(backdrop);
										} else if (
											filterOutInDB === "inDB" &&
											dbResult.ExistInDatabase
										) {
											backdropMovies.push(backdrop);
										} else if (
											filterOutInDB === "notInDB" &&
											!dbResult.ExistInDatabase
										) {
											backdropMovies.push(backdrop);
										}
									}
								}
							)
						);
						if (
							posterMovies.length > 0 ||
							backdropMovies.length > 0
						) {
							userCollectionSets.push({
								...collectionSet,
								movie_posters: posterMovies,
								movie_backdrops: backdropMovies,
							});
						}
					})
				);
				log("User Page - Filtered Collection Sets", userCollectionSets);
				setCollectionSets(userCollectionSets);
				setIdbCollectionSets(userCollectionSets);
			}

			setIsLoading(false);
		};

		filterOutItems();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [selectedLibrarySection]);

	// Add this effect to handle filterOutInDB changes
	useEffect(() => {
		if (!selectedLibrarySection) return;

		const filterByDatabaseStatus = (item: {
			MediaItem: { ExistInDatabase: boolean };
		}) => {
			if (filterOutInDB === "all") return true;
			if (filterOutInDB === "inDB")
				return item.MediaItem?.ExistInDatabase;
			if (filterOutInDB === "notInDB")
				return !item.MediaItem?.ExistInDatabase;
			return false;
		};
		// Filter boxsets
		if (idbBoxsets.length > 0) {
			const filteredBoxsets = idbBoxsets
				.map((boxset) => {
					const newBoxset = { ...boxset };

					if (selectedLibrarySection.type === "show") {
						newBoxset.show_sets = boxset.show_sets.filter(
							filterByDatabaseStatus
						);
					} else if (selectedLibrarySection.type === "movie") {
						newBoxset.movie_sets = boxset.movie_sets.filter(
							filterByDatabaseStatus
						);
						newBoxset.collection_sets = boxset.collection_sets
							.map((collectionSet) => ({
								...collectionSet,
								movie_posters:
									collectionSet.movie_posters.filter(
										(poster) =>
											filterByDatabaseStatus(poster.movie)
									),
								movie_backdrops:
									collectionSet.movie_backdrops.filter(
										(backdrop) =>
											filterByDatabaseStatus(
												backdrop.movie
											)
									),
							}))
							.filter(
								(set) =>
									set.movie_posters.length > 0 ||
									set.movie_backdrops.length > 0
							);
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
			const filteredShowSets = idbShowSets.filter(filterByDatabaseStatus);
			setShowSets(filteredShowSets);
		}

		// Filter movie sets
		if (idbMovieSets.length > 0) {
			const filteredMovieSets = idbMovieSets.filter(
				filterByDatabaseStatus
			);
			setMovieSets(filteredMovieSets);
		}

		// Filter collection sets
		if (idbCollectionSets.length > 0) {
			const filteredCollectionSets = idbCollectionSets
				.map((collectionSet) => ({
					...collectionSet,
					movie_posters: collectionSet.movie_posters.filter(
						(poster) => filterByDatabaseStatus(poster.movie)
					),
					movie_backdrops: collectionSet.movie_backdrops.filter(
						(backdrop) => filterByDatabaseStatus(backdrop.movie)
					),
				}))
				.filter(
					(set) =>
						set.movie_posters.length > 0 ||
						set.movie_backdrops.length > 0
				);

			setCollectionSets(filteredCollectionSets);
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [filterOutInDB]);

	// Update document title accordingly.
	useEffect(() => {
		if (hasError) {
			if (typeof window !== "undefined") document.title = "Aura | Error";
		} else {
			if (typeof window !== "undefined")
				document.title = `AURA | ${username} Sets`;
		}
	}, [hasError, username]);

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
	}, [
		activeTab,
		boxsets.length,
		collectionSets.length,
		movieSets.length,
		showSets.length,
		totalPages,
	]);

	const paginatedShowSets = showSets.slice(
		(currentPage - 1) * itemsPerPage,
		currentPage * itemsPerPage
	);
	const paginatedMovieSets = movieSets.slice(
		(currentPage - 1) * itemsPerPage,
		currentPage * itemsPerPage
	);
	const paginatedCollectionSets = collectionSets.slice(
		(currentPage - 1) * itemsPerPage,
		currentPage * itemsPerPage
	);
	const paginatedBoxSets = boxsets.slice(
		(currentPage - 1) * itemsPerPage,
		currentPage * itemsPerPage
	);

	return (
		<div className="flex flex-col">
			{/* Show loading message */}
			{isLoading && (
				<div className="flex justify-center mt-4">
					<Loader message={loadMessage} />
				</div>
			)}

			{/* Show error message if there is an error */}
			{hasError && (
				<div className="flex justify-center">
					<ErrorMessage message={errorMessage} />
				</div>
			)}

			{/* Show message when no sets are found */}
			{!isLoading &&
				!hasError &&
				respShowSets.length === 0 &&
				respMovieSets.length === 0 &&
				respCollectionSets.length === 0 &&
				respBoxsets.length === 0 && (
					<div className="flex justify-center">
						<ErrorMessage
							message={`No sets found for ${username}`}
						/>
					</div>
				)}

			{/* Main content when sets exist */}
			{!isLoading &&
				!hasError &&
				(respShowSets.length > 0 ||
					respMovieSets.length > 0 ||
					respCollectionSets.length > 0 ||
					respBoxsets.length > 0) && (
					<div className="min-h-screen px-8 pb-20 sm:px-20">
						{/* User Sets Header */}
						<div className="flex flex-col items-center mt-4 mb-4">
							<h1 className="text-3xl font-bold text-center mb-2">
								Sets by {username}
							</h1>
							{!selectedLibrarySection && (
								<div>
									{respShowSets.length > 0 && (
										<p className="text-muted-foreground text-sm">
											Show Sets: {respShowSets.length}
										</p>
									)}
									{respMovieSets.length > 0 && (
										<p className="text-muted-foreground text-sm">
											Movie Sets: {respMovieSets.length}
										</p>
									)}
									{respCollectionSets.length > 0 && (
										<p className="text-muted-foreground text-sm">
											Collection Sets:{" "}
											{respCollectionSets.length}
										</p>
									)}
									{respBoxsets.length > 0 && (
										<p className="text-muted-foreground text-sm">
											Box Sets: {respBoxsets.length}
										</p>
									)}
								</div>
							)}
						</div>

						<div className="w-full max-w-3xl">
							{/* Filter Section */}
							<div className="flex flex-col sm:flex-row mb-4 mt-2">
								<Label
									htmlFor="library-filter"
									className="text-lg font-semibold mb-2 sm:mb-0 sm:mr-4"
								>
									Libraries:
								</Label>

								<ToggleGroup
									type="single"
									className="flex flex-wrap sm:flex-nowrap gap-2"
									value={
										selectedLibrarySection &&
										selectedLibrarySection.title
											? selectedLibrarySection.title
											: ""
									}
									onValueChange={(val: string) => {
										setSelectedLibrarySection(
											val
												? librarySections.find(
														(section) =>
															section.title ===
															val
												  ) || null
												: null
										);
									}}
								>
									{librarySections.map((section) => (
										<Badge
											key={section.title}
											variant={
												selectedLibrarySection?.title ===
												section.title
													? "default"
													: "outline"
											}
											onClick={() => {
												setFilterOutInDB("all");
												if (
													selectedLibrarySection?.title ===
													section.title
												) {
													setSelectedLibrarySection(
														null
													);
													setCurrentPage(1);
												} else {
													setSelectedLibrarySection(
														section
													);
													setCurrentPage(1);
												}
											}}
										>
											{section.title}
										</Badge>
									))}
								</ToggleGroup>
							</div>

							<Badge
								key="filter-out-in-db"
								className={`cursor-pointer ${
									filterOutInDB === "inDB"
										? "bg-green-600 text-white"
										: filterOutInDB === "notInDB"
										? "bg-red-600 text-white"
										: ""
								}`}
								variant={
									filterOutInDB !== "all"
										? "default"
										: "outline"
								}
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

						{/* No library selected message */}
						{!selectedLibrarySection && (
							<div className="flex justify-center mt-8">
								<ErrorMessage message="Select a library to view sets" />
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
										message={`No sets found in ${selectedLibrarySection.title} library`}
									/>
								</div>
							) : (
								<div className="flex flex-col items-center mt-4 mb-4">
									<Tabs
										defaultValue="boxSets"
										value={activeTab}
										onValueChange={(val) => {
											setActiveTab(val);
											setCurrentPage(1);
										}}
										className="mt-2 w-full"
									>
										<TabsList className="flex flex-wrap justify-center">
											{showSets.length > 0 && (
												<TabsTrigger value="showSets">
													Show Sets ({showSets.length}
													)
												</TabsTrigger>
											)}
											{movieSets.length > 0 && (
												<TabsTrigger value="movieSets">
													Movie Sets (
													{movieSets.length})
												</TabsTrigger>
											)}
											{collectionSets.length > 0 && (
												<TabsTrigger value="collectionSets">
													Collection Sets (
													{collectionSets.length})
												</TabsTrigger>
											)}
											{boxsets.length > 0 && (
												<TabsTrigger value="boxSets">
													Box Sets ({boxsets.length})
												</TabsTrigger>
											)}
										</TabsList>

										<div className="mt-4">
											{paginatedShowSets.length > 0 && (
												<TabsContent value="showSets">
													<div className="divide-y divide-primary-dynamic/20 space-y-6">
														{paginatedShowSets.map(
															(showSet) => (
																<div
																	key={`${showSet.id}-showset`}
																	className="pb-6"
																>
																	<RenderShowSetsCarousel
																		key={
																			showSet.id
																		}
																		showSet={
																			showSet
																		}
																	/>
																</div>
															)
														)}
													</div>
												</TabsContent>
											)}

											{paginatedMovieSets.length > 0 && (
												<TabsContent value="movieSets">
													<div className="divide-y divide-primary-dynamic/20 space-y-6">
														{paginatedMovieSets.map(
															(movieSet) => (
																<div
																	key={`${movieSet.id}-movieset`}
																	className="pb-6"
																>
																	<RenderMovieSetsCarousel
																		key={
																			movieSet.id
																		}
																		movieSet={
																			movieSet
																		}
																	/>
																</div>
															)
														)}
													</div>
												</TabsContent>
											)}

											{paginatedCollectionSets.length >
												0 && (
												<TabsContent value="collectionSets">
													<div className="divide-y divide-primary-dynamic/20 space-y-6">
														{paginatedCollectionSets.map(
															(collectionSet) => (
																<div
																	key={`${collectionSet.id}-collectionset`}
																	className="pb-6"
																>
																	<RenderCollectionSetsCarousel
																		key={
																			collectionSet.id
																		}
																		collectionSet={
																			collectionSet
																		}
																	/>
																</div>
															)
														)}
													</div>
												</TabsContent>
											)}

											{paginatedBoxSets.length > 0 && (
												<TabsContent value="boxSets">
													<div className="divide-y divide-primary-dynamic/20 space-y-6">
														{paginatedBoxSets.map(
															(boxset) => (
																<div
																	key={`${boxset.id}-boxset`}
																	className="pb-6"
																>
																	<RenderBoxsetsSection
																		key={
																			boxset.id
																		}
																		boxset={
																			boxset
																		}
																		libraryType={
																			selectedLibrarySection.type
																		}
																	/>
																</div>
															)
														)}
													</div>
												</TabsContent>
											)}
										</div>
									</Tabs>

									{/* Pagination Component */}
									<div className="flex justify-center mt-8">
										<Pagination>
											<PaginationContent>
												{totalPages > 1 && (
													<PaginationItem>
														<PaginationPrevious
															onClick={() => {
																const newPage =
																	Math.max(
																		currentPage -
																			1,
																		1
																	);
																setCurrentPage(
																	newPage
																);
																window.scrollTo(
																	{
																		top: 0,
																		behavior:
																			"smooth",
																	}
																);
															}}
														/>
													</PaginationItem>
												)}
												<PaginationItem>
													<PaginationLink isActive>
														{currentPage}
													</PaginationLink>
												</PaginationItem>
												{totalPages > 1 &&
													currentPage <
														totalPages && (
														<PaginationItem>
															<PaginationNext
																onClick={() => {
																	const newPage =
																		Math.min(
																			currentPage +
																				1,
																			totalPages
																		);
																	setCurrentPage(
																		newPage
																	);
																	window.scrollTo(
																		{
																			top: 0,
																			behavior:
																				"smooth",
																		}
																	);
																}}
															/>
														</PaginationItem>
													)}
												{totalPages > 3 &&
													currentPage <
														totalPages - 1 && (
														<>
															<PaginationItem>
																<PaginationEllipsis />
															</PaginationItem>
															<PaginationItem>
																<PaginationLink
																	onClick={() => {
																		setCurrentPage(
																			totalPages
																		);
																		window.scrollTo(
																			{
																				top: 0,
																				behavior:
																					"smooth",
																			}
																		);
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
								</div>
							))}
					</div>
				)}
		</div>
	);
};

export default UserSetPage;

const RenderShowSetsCarousel = ({
	showSet,
}: {
	showSet: MediuxUserShowSet;
}) => {
	const router = useRouter();
	const { setPosterSet } = usePosterSetStore();
	const { setMediaItem } = useMediaStore();

	const posterSet: PosterSet = {
		ID: showSet.id,
		Title: showSet.set_title,
		Type: "show",
		User: {
			Name: showSet?.user_created?.username || "Unknown User",
		},
		DateCreated: showSet.date_created,
		DateUpdated: showSet.date_updated,
		Poster:
			showSet.show_poster && showSet.show_poster.length > 0
				? {
						ID: showSet.show_poster[0].id,
						Type: "poster",
						Modified: showSet.show_poster[0].modified_on,
						FileSize: Number(showSet.show_poster[0].filesize),
				  }
				: undefined,
		Backdrop:
			showSet.show_backdrop && showSet.show_backdrop.length > 0
				? {
						ID: showSet.show_backdrop[0].id,
						Type: "backdrop",
						Modified: showSet.show_backdrop[0].modified_on,
						FileSize: Number(showSet.show_backdrop[0].filesize),
				  }
				: undefined,
		SeasonPosters: showSet.season_posters.map((poster) => ({
			ID: poster.id,
			Type: "seasonPoster",
			Modified: poster.modified_on,
			FileSize: Number(poster.filesize),
			Season: {
				Number: poster.season.season_number,
			},
		})),
		TitleCards: showSet.titlecards.map((titlecard) => ({
			ID: titlecard.id,
			Type: "titlecard",
			Modified: titlecard.modified_on,
			FileSize: Number(titlecard.filesize),
			Episode: {
				Title: titlecard.episode.episode_title,
				EpisodeNumber: titlecard.episode.episode_number,
				SeasonNumber: titlecard.episode.season_id.season_number,
			},
		})),
		Status: showSet.show_id.status,
	};

	return (
		<Carousel
			opts={{
				align: "start",
				dragFree: true,
				slidesToScroll: "auto",
			}}
			className="w-full"
		>
			<div className="flex flex-col">
				<div className="flex flex-row items-center">
					<div className="flex flex-row items-center">
						<div
							className="text-primary-dynamic hover:text-primary cursor-pointer text-md font-semibold"
							onClick={() => {
								setPosterSet(posterSet);
								setMediaItem(showSet.MediaItem);
								router.push(`/sets/${posterSet.ID}`);
							}}
						>
							{showSet.set_title}
						</div>
						{showSet.MediaItem.ExistInDatabase && (
							<Checkmark
								className="ml-2 text-green-500"
								size={20}
							/>
						)}
					</div>
					<div className="ml-auto flex space-x-2">
						<button className="btn">
							<DownloadModalShow
								posterSet={posterSet}
								mediaItem={showSet.MediaItem}
							/>
						</button>
					</div>
				</div>
				<Lead className="text-sm text-muted-foreground flex items-center mb-1">
					Last Update:{" "}
					{formatLastUpdatedDate(
						showSet.date_updated,
						showSet.date_created
					)}
				</Lead>
			</div>

			<CarouselContent>
				<CarouselShow set={posterSet as PosterSet} />
			</CarouselContent>
			<CarouselNext className="right-2 bottom-0" />
			<CarouselPrevious className="right-8 bottom-0" />
		</Carousel>
	);
};

const RenderMovieSetsCarousel = ({
	movieSet,
}: {
	movieSet: MediuxUserMovieSet;
}) => {
	const router = useRouter();
	const { setPosterSet } = usePosterSetStore();
	const { setMediaItem } = useMediaStore();

	const posterSet: PosterSet = {
		ID: movieSet.id,
		Title: movieSet.set_title,
		Type: "movie",
		User: {
			Name: movieSet.user_created.username,
		},
		DateCreated: movieSet.date_created,
		DateUpdated: movieSet.date_updated,
		Poster:
			movieSet.movie_poster && movieSet.movie_poster.length > 0
				? {
						ID: movieSet.movie_poster[0].id,
						Type: "poster",
						Modified: movieSet.movie_poster[0].modified_on,
						FileSize: Number(movieSet.movie_poster[0].filesize),
				  }
				: undefined,
		Backdrop:
			movieSet.movie_backdrop && movieSet.movie_backdrop.length > 0
				? {
						ID: movieSet.movie_backdrop[0].id,
						Type: "backdrop",
						Modified: movieSet.movie_backdrop[0].modified_on,
						FileSize: Number(movieSet.movie_backdrop[0].filesize),
				  }
				: undefined,

		Status: movieSet.movie_id.status,
	};

	return (
		<Carousel
			opts={{
				align: "start",
				dragFree: true,
				slidesToScroll: "auto",
			}}
			className="w-full"
		>
			<div className="flex flex-col">
				<div className="flex flex-row items-center">
					<div className="flex flex-row items-center">
						<div
							className="text-primary-dynamic hover:text-primary cursor-pointer text-md font-semibold"
							onClick={() => {
								setPosterSet(posterSet);
								setMediaItem(movieSet.MediaItem);
								router.push(`/sets/${posterSet.ID}`);
							}}
						>
							{movieSet.set_title}
						</div>
						{movieSet.MediaItem.ExistInDatabase && (
							<Checkmark
								className="ml-2 text-green-500"
								size={20}
							/>
						)}
					</div>
					<div className="ml-auto flex space-x-2">
						<button className="btn">
							{
								<DownloadModalMovie
									posterSet={posterSet}
									mediaItem={movieSet.MediaItem}
								/>
							}
						</button>
					</div>
				</div>
				<Lead className="text-sm text-muted-foreground flex items-center mb-1">
					Last Update:{" "}
					{formatLastUpdatedDate(
						movieSet.date_updated,
						movieSet.date_created
					)}
				</Lead>
			</div>

			<CarouselContent>
				<CarouselMovie
					set={posterSet as PosterSet}
					librarySection={movieSet.MediaItem.LibraryTitle}
				/>
			</CarouselContent>
			<CarouselNext className="right-2 bottom-0" />
			<CarouselPrevious className="right-8 bottom-0" />
		</Carousel>
	);
};

const RenderCollectionSetsCarousel = ({
	collectionSet,
}: {
	collectionSet: MediuxUserCollectionSet;
}) => {
	const router = useRouter();
	const { setPosterSet } = usePosterSetStore();

	const posterSet: PosterSet = {
		ID: collectionSet.id,
		Title: collectionSet.set_title,
		Type: "collection",
		User: {
			Name: collectionSet.user_created.username,
		},
		DateCreated: collectionSet.date_created,
		DateUpdated: collectionSet.date_updated,
		OtherPosters: collectionSet.movie_posters.map((poster) => ({
			ID: poster.id,
			Type: "poster",
			Modified: poster.modified_on,
			FileSize: Number(poster.filesize),
			Movie: {
				ID: poster.movie.id,
				Title: poster.movie.title,
				Status: poster.movie.status,
				Tagline: poster.movie.tagline,
				Slug: poster.movie.slug,
				DateUpdated: poster.movie.date_updated,
				TVbdID: poster.movie.tvdb_id,
				ImdbID: poster.movie.imdb_id,
				TraktID: poster.movie.trakt_id,
				ReleaseDate: poster.movie.release_date,
				RatingKey: poster.movie.MediaItem?.RatingKey || "",
				LibrarySection: poster.movie?.MediaItem?.LibraryTitle || "",
			},
		})),
		OtherBackdrops: collectionSet.movie_backdrops.map((backdrop) => ({
			ID: backdrop.id,
			Type: "backdrop",
			Modified: backdrop.modified_on,
			FileSize: Number(backdrop.filesize),
			Movie: {
				ID: backdrop.movie.id,
				Title: backdrop.movie.title,
				Status: backdrop.movie.status,
				Tagline: backdrop.movie.tagline,
				Slug: backdrop.movie.slug,
				DateUpdated: backdrop.movie.date_updated,
				TVbdID: backdrop.movie.tvdb_id,
				ImdbID: backdrop.movie.imdb_id,
				TraktID: backdrop.movie.trakt_id,
				ReleaseDate: backdrop.movie.release_date,
				RatingKey: backdrop.movie.MediaItem?.RatingKey || "",
				LibrarySection: backdrop.movie?.MediaItem?.LibraryTitle || "",
			},
		})),
		Status: "none",
	};
	return (
		<Carousel
			opts={{
				align: "start",
				dragFree: true,
				slidesToScroll: "auto",
			}}
			className="w-full"
		>
			<div className="flex flex-col">
				<div className="flex flex-row items-center">
					<div className="flex flex-row items-center">
						<div
							className="text-primary-dynamic hover:text-primary cursor-pointer text-md font-semibold"
							onClick={() => {
								setPosterSet(posterSet);
								router.push(`/sets/${posterSet.ID}`);
							}}
						>
							{collectionSet.set_title}
						</div>
					</div>
					<div className="ml-auto flex space-x-2">
						<button className="btn">
							{
								<DownloadModalMovie
									posterSet={posterSet}
									mediaItem={
										collectionSet.movie_posters[0].movie
											.MediaItem
									}
								/>
							}
						</button>
					</div>
				</div>
				<Lead className="text-sm text-muted-foreground flex items-center mb-1">
					Last Update:{" "}
					{formatLastUpdatedDate(
						collectionSet.date_updated,
						collectionSet.date_created
					)}
				</Lead>
			</div>

			<CarouselContent>
				<CarouselMovie
					set={posterSet as PosterSet}
					librarySection={(() => {
						const sections =
							posterSet.OtherPosters?.map(
								(p) => p.Movie?.LibrarySection
							).filter(Boolean) || [];
						if (sections.length === 0) return "";
						const freq: Record<string, number> = {};
						sections.forEach((s) => {
							if (typeof s === "string") {
								freq[s] = (freq[s] || 0) + 1;
							}
						});
						return Object.entries(freq).sort(
							(a, b) => b[1] - a[1]
						)[0][0];
					})()}
				/>
			</CarouselContent>
			<CarouselNext className="right-2 bottom-0" />
			<CarouselPrevious className="right-8 bottom-0" />
		</Carousel>
	);
};

const RenderBoxsetsSection = ({
	boxset,
	libraryType,
}: {
	boxset: MediuxUserBoxset;
	libraryType: string;
}) => {
	return (
		<div>
			<Accordion type="single" collapsible className="w-full">
				<AccordionItem value={boxset.id}>
					<AccordionTrigger className="flex items-center justify-between">
						<div className="text-primary-dynamic hover:text-primary cursor-pointer text-lg font-semibold">
							{boxset.boxset_title}
						</div>
					</AccordionTrigger>
					<AccordionContent>
						<div className="flex flex-col space-y-4">
							<BoxsetDisplay
								boxset={boxset}
								libraryType={libraryType}
							/>
						</div>
					</AccordionContent>
				</AccordionItem>
			</Accordion>
		</div>
	);
};
