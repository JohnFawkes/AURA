import { formatDownloadSize } from "@/helper/formatDownloadSize";
import { postAddItemToDB } from "@/services/api.db";
import {
	fetchMediaServerItemContent,
	patchDownloadPosterFileAndUpdateMediaServer,
} from "@/services/api.mediaserver";
import { zodResolver } from "@hookform/resolvers/zod";
import { Check, Download, LoaderIcon, TriangleAlert, X } from "lucide-react";
import { z } from "zod";

import { useEffect, useState } from "react";
import { useForm, useWatch } from "react-hook-form";

import Link from "next/link";

import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogOverlay,
	DialogPortal,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Progress } from "@/components/ui/progress";

import { log } from "@/lib/logger";

import { DBSavedItem } from "@/types/databaseSavedSet";
import { MediaItem } from "@/types/mediaItem";
import {
	MediuxUserBoxset,
	MediuxUserCollectionMovie,
	MediuxUserCollectionSet,
	MediuxUserMovieSet,
	MediuxUserShowSet,
} from "@/types/mediuxUserAllSets";
import { PosterFile } from "@/types/posterSets";

const formSchema = z
	.object({
		selectedTypesByItem: z.record(
			z.object({
				types: z.array(z.string()),
				autodownload: z.boolean().optional(),
				futureUpdatesOnly: z.boolean().optional(),
				source: z.enum(["movie", "collection"]).optional(),
				addToDBOnly: z.boolean().optional(),
			})
		),
	})
	.refine(
		(data) =>
			Object.values(data.selectedTypesByItem).some(
				(item) => Array.isArray(item.types) && item.types.length > 0
			),
		{
			message: "Please select at least one image type to download.",
			path: ["selectedTypesByItem"],
		}
	);

type ShowProgressSteps = {
	poster: number;
	backdrop: number;
	seasonPoster: number;
	specialSeasonPoster: number;
	titleCard: number;
	addToDB: number; // 1 per show set
};

function calculateShowSteps(
	showSets: MediuxUserShowSet[],
	selectedShows: Array<[string, { types: string[]; futureUpdatesOnly: boolean }]>
): ShowProgressSteps {
	return showSets.reduce<ShowProgressSteps>(
		(acc, show) => {
			// Find the corresponding selected item data
			const selectedItem = selectedShows.find(([key]) => key === show.MediaItem.RatingKey);

			// If item not found or futureUpdatesOnly is true, don't count its files
			if (!selectedItem || selectedItem[1].futureUpdatesOnly) {
				return {
					...acc,
					addToDB: acc.addToDB + 1, // Always count DB entries
				};
			}

			const selectedTypes = selectedItem[1].types;

			return {
				poster:
					acc.poster + (selectedTypes.includes("poster") ? show.show_poster.length : 0),
				backdrop:
					acc.backdrop +
					(selectedTypes.includes("backdrop") ? show.show_backdrop.length : 0),
				seasonPoster:
					acc.seasonPoster +
					(selectedTypes.includes("seasonPoster")
						? show.season_posters.filter((p) => p.season.season_number !== 0).length
						: 0),
				specialSeasonPoster:
					acc.specialSeasonPoster +
					(selectedTypes.includes("specialSeasonPoster")
						? show.season_posters.filter((p) => p.season.season_number === 0).length
						: 0),
				titleCard:
					acc.titleCard +
					(selectedTypes.includes("titlecard") ? show.titlecards.length : 0),
				addToDB: acc.addToDB + 1, // Always count DB entries
			};
		},
		{
			poster: 0,
			backdrop: 0,
			seasonPoster: 0,
			specialSeasonPoster: 0,
			titleCard: 0,
			addToDB: 0,
		}
	);
}

type MovieProgressSteps = {
	poster: number;
	backdrop: number;
	addToDB: number;
};

function calculateMovieSteps(
	selectedItems: Array<[string, { types: string[]; source?: "movie" | "collection" }]>
): MovieProgressSteps {
	return selectedItems.reduce<MovieProgressSteps>(
		(acc, [, itemData]) => {
			// Skip if no types selected
			if (!itemData.types || itemData.types.length === 0) {
				return acc;
			}

			// Count poster and backdrop based on selected types
			// Add to DB count if poster or backdrop is selected
			if (itemData.types.includes("poster") || itemData.types.includes("backdrop")) {
				if (itemData.types.includes("poster")) {
					acc.poster += 1;
				}
				if (itemData.types.includes("backdrop")) {
					acc.backdrop += 1;
				}

				acc.addToDB += 1;
			}

			return acc;
		},
		{
			poster: 0,
			backdrop: 0,
			addToDB: 0,
		}
	);
}

const DownloadModalBoxset: React.FC<{
	boxset: MediuxUserBoxset;
	libraryType: string;
}> = ({ boxset, libraryType }) => {
	const [isMounted, setIsMounted] = useState(false);
	const [cancelButtonText, setCancelButtonText] = useState("Cancel");
	const [downloadButtonText, setDownloadButtonText] = useState("Download");
	const [totalSelectedSize, setTotalSelectedSize] = useState("");
	const [totalSelectedSizeLabel, setTotalSelectedLabel] = useState("Total Download Size: ");

	// Update the progress text state type
	type ItemProgressText = {
		[itemTitle: string]: {
			poster?: string;
			backdrop?: string;
			seasonPoster?: string;
			specialSeasonPoster?: string;
			titleCard?: string;
			addToDB?: string;
		};
	};

	// Download Progress
	const [progressValues, setProgressValues] = useState<{
		progressValue: number;
		progressColor: string;
		progressText: ItemProgressText;
		warningMessages: string[];
	}>({
		progressValue: 0,
		progressColor: "",
		progressText: {},
		warningMessages: [],
	});

	const form = useForm<z.infer<typeof formSchema>>({
		resolver: zodResolver(formSchema),
		defaultValues: {
			selectedTypesByItem: (() => {
				const showSetTypes = boxset.show_sets.reduce(
					(acc, showSet) => {
						const types: string[] = [];
						if (showSet.show_poster.length > 0) types.push("poster");
						if (showSet.show_backdrop.length > 0) types.push("backdrop");
						if (showSet.season_posters.length > 0) types.push("seasonPoster");
						if (
							showSet.season_posters.some(
								(seasonPoster) => seasonPoster.season.season_number === 0
							)
						)
							types.push("specialSeasonPoster");
						if (showSet.titlecards.length > 0) types.push("titlecard");

						acc[showSet.MediaItem.RatingKey] = {
							types,
							autodownload: false,
							futureUpdatesOnly: false,
							addToDBOnly: false,
						};
						return acc;
					},
					{} as Record<
						string,
						{
							types: string[];
							autodownload: boolean;
							futureUpdatesOnly: boolean;
							addToDBOnly: boolean;
						}
					>
				);

				// Create a Set to track movies that exist in both sets
				const moviesInBothSets = new Set<string>();

				// Find movies that exist in both sets
				boxset.movie_sets.forEach((movieSet) => {
					const ratingKey = movieSet.MediaItem.RatingKey;
					const inCollection = boxset.collection_sets.some((collection) =>
						collection.movie_posters.some(
							(poster) => poster.movie.MediaItem.RatingKey === ratingKey
						)
					);
					if (inCollection) {
						moviesInBothSets.add(ratingKey);
					}
				});

				// Initialize movie sets (with duplicates checked by default)
				const movieSetTypes = boxset.movie_sets.reduce(
					(acc, movieSet) => {
						const types: string[] = [];
						const ratingKey = movieSet.MediaItem.RatingKey;

						if (movieSet.movie_poster.length > 0) types.push("poster");
						if (movieSet.movie_backdrop.length > 0) types.push("backdrop");

						// If this movie exists in both sets, include its types
						if (moviesInBothSets.has(ratingKey)) {
							acc[ratingKey] = {
								types,
								autodownload: false,
								futureUpdatesOnly: false,
								source: "movie",
								addToDBOnly: false,
							};
						} else {
							acc[ratingKey] = {
								types,
								autodownload: false,
								futureUpdatesOnly: false,
								source: "movie",
								addToDBOnly: false,
							};
						}
						return acc;
					},
					{} as Record<
						string,
						{
							types: string[];
							autodownload: boolean;
							futureUpdatesOnly: boolean;
							source: "movie" | "collection" | undefined;
							addToDBOnly: boolean;
						}
					>
				);

				// Initialize collection sets (with duplicates unchecked)
				const collectionSetTypes = boxset.collection_sets.reduce(
					(acc, collectionSet) => {
						collectionSet.movie_posters.forEach((poster) => {
							const movieKey = poster.movie.MediaItem.RatingKey;
							// If the movie exists in both sets, initialize with empty types
							if (moviesInBothSets.has(movieKey)) {
								acc[movieKey] = {
									types: [],
									autodownload: false,
									futureUpdatesOnly: false,
									source: "collection",
									addToDBOnly: false,
								};
							} else {
								const types: string[] = [];
								types.push("poster");
								if (
									collectionSet.movie_backdrops.find(
										(b) => b.movie.id === poster.movie.id
									)
								) {
									types.push("backdrop");
								}
								acc[movieKey] = {
									types,
									autodownload: false,
									futureUpdatesOnly: false,
									source: "collection",
									addToDBOnly: false,
								};
							}
						});
						return acc;
					},
					{} as Record<
						string,
						{
							types: string[];
							autodownload: boolean;
							futureUpdatesOnly: boolean;
							source: "movie" | "collection" | undefined;
							addToDBOnly: boolean;
						}
					>
				);

				return {
					...showSetTypes,
					...movieSetTypes,
					...collectionSetTypes,
				};
			})(),
		},
	});

	useEffect(() => {
		form.reset({
			selectedTypesByItem: (() => {
				// Initialize show sets
				const showSetTypes = boxset.show_sets.reduce(
					(acc, showSet) => {
						const types: string[] = [];
						if (showSet.show_poster.length > 0) types.push("poster");
						if (showSet.show_backdrop.length > 0) types.push("backdrop");
						if (showSet.season_posters.length > 0) types.push("seasonPoster");
						if (
							showSet.season_posters.some(
								(seasonPoster) => seasonPoster.season.season_number === 0
							)
						)
							types.push("specialSeasonPoster");
						if (showSet.titlecards.length > 0) types.push("titlecard");

						acc[showSet.MediaItem.RatingKey] = {
							types,
							autodownload: false,
							futureUpdatesOnly: false,
							addToDBOnly: false,
						};
						return acc;
					},
					{} as Record<
						string,
						{
							types: string[];
							autodownload: boolean;
							futureUpdatesOnly: boolean;
							addToDBOnly: boolean;
						}
					>
				);

				// Create a Set to track movies that exist in both sets
				const moviesInBothSets = new Set<string>();

				// Find movies that exist in both sets
				boxset.movie_sets.forEach((movieSet) => {
					const ratingKey = movieSet.MediaItem.RatingKey;
					const inCollection = boxset.collection_sets.some((collection) =>
						collection.movie_posters.some(
							(poster) => poster.movie.MediaItem.RatingKey === ratingKey
						)
					);
					if (inCollection) {
						moviesInBothSets.add(ratingKey);
					}
				});

				// Initialize movie sets (with duplicates checked by default)
				const movieSetTypes = boxset.movie_sets.reduce(
					(acc, movieSet) => {
						const types: string[] = [];
						const ratingKey = movieSet.MediaItem.RatingKey;

						if (movieSet.movie_poster.length > 0) types.push("poster");
						if (movieSet.movie_backdrop.length > 0) types.push("backdrop");

						// If this movie exists in both sets, initialize with its types
						acc[ratingKey] = {
							types,
							autodownload: false,
							futureUpdatesOnly: false,
							source: "movie",
							addToDBOnly: false,
						};
						return acc;
					},
					{} as Record<
						string,
						{
							types: string[];
							autodownload: boolean;
							futureUpdatesOnly: boolean;
							source: "movie" | "collection" | undefined;
							addToDBOnly: boolean;
						}
					>
				);

				// Initialize collection sets (with duplicates unchecked)
				const collectionSetTypes = boxset.collection_sets.reduce(
					(acc, collectionSet) => {
						collectionSet.movie_posters.forEach((poster) => {
							const movieKey = poster.movie.MediaItem.RatingKey;

							// Always initialize with empty types if movie exists in both sets
							if (moviesInBothSets.has(movieKey)) {
								acc[movieKey] = {
									types: [],
									autodownload: false,
									futureUpdatesOnly: false,
									source: "collection",
									addToDBOnly: false,
								};
							} else {
								const types: string[] = [];
								types.push("poster");
								if (
									collectionSet.movie_backdrops.find(
										(b) => b.movie.id === poster.movie.id
									)
								) {
									types.push("backdrop");
								}
								acc[movieKey] = {
									types,
									autodownload: false,
									futureUpdatesOnly: false,
									source: "collection",
									addToDBOnly: false,
								};
							}
						});
						return acc;
					},
					{} as Record<
						string,
						{
							types: string[];
							autodownload: boolean;
							futureUpdatesOnly: boolean;
							source: "movie" | "collection" | undefined;
							addToDBOnly: boolean;
						}
					>
				);

				return {
					...showSetTypes,
					...movieSetTypes, // Movie sets take precedence for duplicates
					...collectionSetTypes,
				};
			})(),
		});
	}, [boxset, form]);

	const watchSelectedTypesByItem = useWatch({
		control: form.control,
		name: "selectedTypesByItem",
	}) as Record<
		string,
		{
			types: string[];
			autodownload: boolean;
			futureUpdatesOnly: boolean;
			source?: "movie" | "collection"; // Optional field to track source
		}
	>;

	useEffect(() => {
		const totalSize = Object.entries(watchSelectedTypesByItem).reduce(
			(acc, [itemKey, itemData]) => {
				// Early return if no types selected
				if (!itemData?.types || !Array.isArray(itemData.types)) {
					return acc;
				}

				const selectedTypes = itemData.types;

				const movie = boxset.movie_sets.find((m) => m.MediaItem.RatingKey === itemKey);
				if (movie) {
					const moviePosterSize = movie.movie_poster.reduce(
						(total, poster) => total + Number(poster.filesize),
						0
					);
					const movieBackdropSize = movie.movie_backdrop.reduce(
						(total, backdrop) => total + Number(backdrop.filesize),
						0
					);
					if (selectedTypes.includes("poster")) {
						acc += moviePosterSize;
					}
					if (selectedTypes.includes("backdrop")) {
						acc += movieBackdropSize;
					}
					return acc;
				}

				const show = boxset.show_sets.find((s) => s.MediaItem.RatingKey === itemKey);
				if (show) {
					const showPosterSize = show.show_poster.reduce(
						(total, poster) => total + Number(poster.filesize),
						0
					);
					const showBackdropSize = show.show_backdrop.reduce(
						(total, backdrop) => total + Number(backdrop.filesize),
						0
					);
					const seasonPosterSize = show.season_posters.reduce(
						(total, seasonPoster) => total + Number(seasonPoster.filesize),
						0
					);
					const specialSeasonPosterSize = show.season_posters
						.filter((seasonPoster) => seasonPoster.season.season_number === 0)
						.reduce((total, seasonPoster) => total + Number(seasonPoster.filesize), 0);
					const titleCardSize = show.titlecards.reduce(
						(total, titleCard) => total + Number(titleCard.filesize),
						0
					);

					if (selectedTypes.includes("poster")) acc += showPosterSize;
					if (selectedTypes.includes("backdrop")) acc += showBackdropSize;
					if (selectedTypes.includes("seasonPoster")) acc += seasonPosterSize;
					if (selectedTypes.includes("specialSeasonPoster"))
						acc += specialSeasonPosterSize;
					if (selectedTypes.includes("titlecard")) acc += titleCardSize;
					return acc;
				}

				const movieFromCollection = boxset.collection_sets.some((collection) =>
					collection.movie_posters.some(
						(poster) => poster.movie.MediaItem.RatingKey === itemKey
					)
				);

				if (movieFromCollection) {
					const collection = boxset.collection_sets.find((collection) =>
						collection.movie_posters.some(
							(poster) => poster.movie.MediaItem.RatingKey === itemKey
						)
					);

					if (collection) {
						const poster = collection.movie_posters.find(
							(p) => p.movie.MediaItem.RatingKey === itemKey
						);
						const backdrop = collection.movie_backdrops.find(
							(b) => b.movie.MediaItem.RatingKey === itemKey
						);

						if (selectedTypes.includes("poster") && poster) {
							acc += Number(poster.filesize);
						}
						if (selectedTypes.includes("backdrop") && backdrop) {
							acc += Number(backdrop.filesize);
						}
					}
					return acc;
				}

				return acc;
			},
			0
		);

		setTotalSelectedSize(formatDownloadSize(totalSize));
	}, [watchSelectedTypesByItem, boxset]);

	const resetProgressValues = () => {
		setProgressValues({
			progressValue: 0,
			progressColor: "",
			progressText: {},
			warningMessages: [],
		});
	};

	function calculateShowTotalSteps(steps: ShowProgressSteps): number {
		return Object.values(steps).reduce((acc, val) => acc + val, 0);
	}

	function calculateMovieTotalSteps(steps: MovieProgressSteps): number {
		return Object.values(steps).reduce((acc, val) => acc + val, 0);
	}

	const handleClose = () => {
		setIsMounted(false);
		setCancelButtonText("Cancel");
		setDownloadButtonText("Download");
		setTotalSelectedSize("");
		setTotalSelectedLabel("Total Download Size: ");
		resetProgressValues();
		form.reset();
	};

	const downloadPosterFileAndUpdateMediaServer = async (
		posterFile: PosterFile,
		fileName: string,
		mediaItem: MediaItem
	) => {
		try {
			const resp = await patchDownloadPosterFileAndUpdateMediaServer(posterFile, mediaItem);
			if (resp.status !== "success") {
				throw new Error(`Failed to download ${fileName}`);
			} else {
				return resp.data;
			}
		} catch {
			setProgressValues((prev) => ({
				...prev,
				progressColor: "yellow",
				warningMessages: [...prev.warningMessages, fileName],
			}));
			return null;
		}
	};

	const onSubmit = async (data: z.infer<typeof formSchema>) => {
		if (isMounted) return;
		setIsMounted(true);

		setCancelButtonText("Close");
		setDownloadButtonText("Downloading...");
		resetProgressValues();

		try {
			const selectedItems = Object.entries(data.selectedTypesByItem);
			if (libraryType === "show") {
				// Get all selected show items
				const selectedShows = selectedItems.filter(([itemKey]) =>
					boxset.show_sets.some((show) => show.MediaItem.RatingKey === itemKey)
				);

				// Calculate steps for all selected shows
				const totalSteps = calculateShowSteps(
					boxset.show_sets,
					selectedShows.map(([key, itemData]) => [
						key,
						{
							types: itemData.types,
							futureUpdatesOnly: itemData.futureUpdatesOnly ?? false,
						},
					])
				);

				const totalFileCount = calculateShowTotalSteps(totalSteps);

				// Reserve 1% for start and 1% for completion
				const progressIncrement = totalFileCount > 0 ? 98 / totalFileCount : 0;

				setProgressValues((prev) => ({
					...prev,
					progressValue: 1,
				}));

				// Only process show sets for show libraries
				for (const [itemKey, itemData] of selectedItems) {
					const show = boxset.show_sets.find((s) => s.MediaItem.RatingKey === itemKey);

					if (!show) {
						continue; // Skip if no show found
					}
					if (
						!itemData.types ||
						itemData.types.length === 0 ||
						(itemData.futureUpdatesOnly && !show.MediaItem.RatingKey)
					) {
						continue; // Skip if no types selected or future updates only
					}

					if (show && itemData.types) {
						// Get the latest Media Item details from the server
						const latestMediaItemResp = await fetchMediaServerItemContent(
							show.MediaItem.RatingKey,
							show.MediaItem.LibraryTitle
						);
						if (!latestMediaItemResp) {
							throw new Error("No response from Plex API");
						}
						if (latestMediaItemResp.status !== "success") {
							throw new Error(latestMediaItemResp.message);
						}
						const latestMediaItem = latestMediaItemResp.data;
						if (!latestMediaItem) {
							throw new Error("No item found in response");
						}

						// Check to see if futureUpdatesOnly is true
						if (itemData.futureUpdatesOnly) {
							const SaveItem: DBSavedItem = {
								MediaItemID: latestMediaItem.RatingKey,
								MediaItem: latestMediaItem,
								PosterSetID: show.id,
								PosterSet: {
									ID: show.id,
									Title: show.set_title,
									Type: "show",
									User: {
										Name: show.user_created.username,
									},
									DateCreated: show.date_created,
									DateUpdated: show.date_updated,
									Poster: {
										ID: show.show_poster[0]?.id,
										Type: "poster",
										Modified: show.show_poster[0]?.modified_on,
										FileSize: Number(show.show_poster[0]?.filesize),
									},
									Backdrop: {
										ID: show.show_backdrop[0]?.id,
										Type: "backdrop",
										Modified: show.show_backdrop[0]?.modified_on,
										FileSize: Number(show.show_backdrop[0]?.filesize),
									},
									SeasonPosters: show.season_posters.map((seasonPoster) => ({
										ID: seasonPoster.id,
										Type: "seasonPoster",
										Modified: seasonPoster.modified_on,
										FileSize: Number(seasonPoster.filesize),
										Season: {
											Number: seasonPoster.season.season_number,
										},
									})),
									TitleCards: show.titlecards.map((titleCard) => ({
										ID: titleCard.id,
										Type: "titlecard",
										Modified: titleCard.modified_on,
										FileSize: Number(titleCard.filesize),
										Episode: {
											Title: titleCard.episode.episode_title,
											EpisodeNumber: titleCard.episode.episode_number,
											SeasonNumber: titleCard.episode.season_id.season_number,
										},
									})),
									Status: "",
								},
								LastDownloaded: new Date().toISOString(),
								SelectedTypes: itemData.types,
								AutoDownload: itemData.autodownload || false,
							};
							setProgressValues((prev) => ({
								...prev,
								progressText: {
									...prev.progressText,
									[show.MediaItem.Title]: {
										...prev.progressText[show.MediaItem.Title],
										addToDB: `Adding ${show.MediaItem.Title} to database`,
									},
								},
							}));
							const addToDBResp = await postAddItemToDB(SaveItem);
							if (addToDBResp.status !== "success") {
								setProgressValues((prev) => ({
									...prev,
									progressColor: "red",
									warningMessages: [
										`Failed to add ${show.MediaItem.Title} to database`,
									],
								}));
							} else {
								setProgressValues((prev) => ({
									...prev,
									progressValue: prev.progressValue + progressIncrement,
									progressText: {
										...prev.progressText,
										[show.MediaItem.Title]: {
											...prev.progressText[show.MediaItem.Title],
											addToDB: `Added ${show.MediaItem.Title} to database`,
										},
									},
								}));
							}
							continue;
						}

						// Sort the selected types to ensure consistent order
						// Poster, Backdrop, SeasonPoster, SpecialSeasonPoster, TitleCard
						const selectedTypes = itemData.types.sort((a, b) => {
							const order = [
								"poster",
								"backdrop",
								"seasonPoster",
								"specialSeasonPoster",
								"titlecard",
							];
							return order.indexOf(a) - order.indexOf(b);
						});

						for (const type of selectedTypes) {
							switch (type) {
								case "poster":
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											[show.MediaItem.Title]: {
												...prev.progressText[show.MediaItem.Title],
												poster: `Downloading Poster`,
											},
										},
									}));

									const posterFile: PosterFile = {
										ID: show.show_poster[0]?.id,
										Type: "poster",
										Modified: show.show_poster[0]?.modified_on,
										FileSize: Number(show.show_poster[0]?.filesize),
									};
									const posterResp = await downloadPosterFileAndUpdateMediaServer(
										posterFile,
										"Poster",
										latestMediaItem
									);
									if (posterResp === null) {
										setProgressValues((prev) => ({
											...prev,
											progressText: {
												...prev.progressText,
												[show.MediaItem.Title]: {
													...prev.progressText[show.MediaItem.Title],
													poster: `Failed to download Poster`,
												},
											},
										}));
									} else {
										setProgressValues((prev) => ({
											...prev,
											progressText: {
												...prev.progressText,
												[show.MediaItem.Title]: {
													...prev.progressText[show.MediaItem.Title],
													poster: `Finished Poster Download`,
												},
											},
										}));
									}
									setProgressValues((prev) => ({
										...prev,
										progressValue: prev.progressValue + progressIncrement,
									}));
									break;

								case "backdrop":
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											[show.MediaItem.Title]: {
												...prev.progressText[show.MediaItem.Title],
												backdrop: `Downloading Backdrop`,
											},
										},
									}));
									const backdropFile: PosterFile = {
										ID: show.show_backdrop[0]?.id,
										Type: "backdrop",
										Modified: show.show_backdrop[0]?.modified_on,
										FileSize: Number(show.show_backdrop[0]?.filesize),
									};
									const backdropResp =
										await downloadPosterFileAndUpdateMediaServer(
											backdropFile,
											"Backdrop",
											latestMediaItem
										);
									if (backdropResp === null) {
										setProgressValues((prev) => ({
											...prev,
											progressText: {
												...prev.progressText,
												[show.MediaItem.Title]: {
													...prev.progressText[show.MediaItem.Title],
													backdrop: `Failed to download Backdrop`,
												},
											},
										}));
									} else {
										setProgressValues((prev) => ({
											...prev,
											progressText: {
												...prev.progressText,
												[show.MediaItem.Title]: {
													...prev.progressText[show.MediaItem.Title],
													backdrop: `Finished Backdrop Download`,
												},
											},
										}));
									}
									setProgressValues((prev) => ({
										...prev,
										progressValue: prev.progressValue + progressIncrement,
									}));
									break;
								case "seasonPoster":
									// Sort season_posters by season_number ascending
									const sortedSeasonPosters = [
										...(show.season_posters || []),
									].sort(
										(a, b) =>
											(a.season.season_number ?? 0) -
											(b.season.season_number ?? 0)
									);
									for (const season of sortedSeasonPosters) {
										if (season.season.season_number === 0) {
											// Skip special season posters
											continue;
										}
										const seasonExists = latestMediaItem.Series?.Seasons?.some(
											(s) => s.SeasonNumber === season.season.season_number
										);
										if (!seasonExists) {
											continue;
										}
										setProgressValues((prev) => ({
											...prev,
											progressText: {
												...prev.progressText,
												[show.MediaItem.Title]: {
													...prev.progressText[show.MediaItem.Title],
													seasonPoster: `Downloading Season Poster ${season.season.season_number
														?.toString()
														.padStart(2, "0")}`,
												},
											},
										}));

										const seasonPosterFile: PosterFile = {
											ID: season.id,
											Type: "seasonPoster",
											Modified: season.modified_on,
											FileSize: Number(season.filesize),
											Season: {
												Number: season.season.season_number,
											},
										};
										const seasonPosterResp =
											await downloadPosterFileAndUpdateMediaServer(
												seasonPosterFile,
												`Season ${season.season.season_number} Poster`,
												latestMediaItem
											);
										if (seasonPosterResp === null) {
											setProgressValues((prev) => ({
												...prev,
												progressText: {
													...prev.progressText,
													[show.MediaItem.Title]: {
														...prev.progressText[show.MediaItem.Title],
														seasonPoster: `Failed to download Season Poster ${season.season.season_number
															?.toString()
															.padStart(2, "0")}`,
													},
												},
											}));
										} else {
											setProgressValues((prev) => ({
												...prev,
												progressText: {
													...prev.progressText,
													[show.MediaItem.Title]: {
														...prev.progressText[show.MediaItem.Title],
														seasonPoster: `Finished Season Poster ${season.season.season_number
															?.toString()
															.padStart(2, "0")} Download`,
													},
												},
											}));
										}
										setProgressValues((prev) => ({
											...prev,
											progressValue: prev.progressValue + progressIncrement,
										}));
									}
									break;

								case "specialSeasonPoster":
									for (const season of show.season_posters || []) {
										if (season.season.season_number !== 0) {
											// Skip regular season posters
											continue;
										}
										const seasonExists = latestMediaItem.Series?.Seasons?.some(
											(s) => s.SeasonNumber === 0
										);
										if (!seasonExists) {
											continue;
										}
										setProgressValues((prev) => ({
											...prev,
											progressText: {
												...prev.progressText,
												[show.MediaItem.Title]: {
													...prev.progressText[show.MediaItem.Title],
													specialSeasonPoster: `Downloading Special Season Poster`,
												},
											},
										}));

										const specialSeasonPosterFile: PosterFile = {
											ID: season.id,
											Type: "specialSeasonPoster",
											Modified: season.modified_on,
											FileSize: Number(season.filesize),
											Season: {
												Number: 0, // Special season
											},
										};
										const specialSeasonPosterResp =
											await downloadPosterFileAndUpdateMediaServer(
												specialSeasonPosterFile,
												`Special Season Poster`,
												latestMediaItem
											);
										if (specialSeasonPosterResp === null) {
											setProgressValues((prev) => ({
												...prev,
												progressText: {
													...prev.progressText,
													[show.MediaItem.Title]: {
														...prev.progressText[show.MediaItem.Title],
														specialSeasonPoster: `Failed to download Special Season Poster`,
													},
												},
											}));
										} else {
											setProgressValues((prev) => ({
												...prev,
												progressText: {
													...prev.progressText,
													[show.MediaItem.Title]: {
														...prev.progressText[show.MediaItem.Title],
														specialSeasonPoster: `Finished Special Season Poster Download`,
													},
												},
											}));
										}
										setProgressValues((prev) => ({
											...prev,
											progressValue: prev.progressValue + progressIncrement,
										}));
									}
									break;
								case "titlecard":
									// Sort titlecards on season number and episode number
									const sortedTitleCards = [...(show.titlecards || [])].sort(
										(a, b) => {
											const aSeason = a.episode.season_id.season_number || 0;
											const bSeason = b.episode.season_id.season_number || 0;
											const aEpisode = a.episode?.episode_number || 0;
											const bEpisode = b.episode?.episode_number || 0;
											if (aSeason !== bSeason) {
												return aSeason - bSeason;
											}
											return aEpisode - bEpisode;
										}
									);
									for (const titleCard of sortedTitleCards) {
										const episode = latestMediaItem.Series?.Seasons.flatMap(
											(s) => s.Episodes
										).find(
											(e) =>
												e.EpisodeNumber ===
													titleCard.episode?.episode_number &&
												e.SeasonNumber ===
													titleCard.episode.season_id.season_number
										);
										if (!episode) {
											// Skip titlecards for episodes that don't exist
											continue;
										}
										setProgressValues((prev) => ({
											...prev,
											progressText: {
												...prev.progressText,
												[show.MediaItem.Title]: {
													...prev.progressText[show.MediaItem.Title],
													titlecard: `Downloading Title Card for S${titleCard.episode.season_id.season_number
														?.toString()
														.padStart(
															2,
															"0"
														)}E${titleCard.episode?.episode_number
														?.toString()
														.padStart(2, "0")}`,
												},
											},
										}));
										const titleCardFile: PosterFile = {
											ID: titleCard.id,
											Type: "titlecard",
											Modified: titleCard.modified_on,
											FileSize: Number(titleCard.filesize),
											Episode: {
												Title: titleCard.episode.episode_title,
												EpisodeNumber: titleCard.episode?.episode_number,
												SeasonNumber:
													titleCard.episode.season_id.season_number,
											},
										};

										const titleCardResp =
											await downloadPosterFileAndUpdateMediaServer(
												titleCardFile,
												`Title Card S${titleCard.episode.season_id.season_number
													?.toString()
													.padStart(
														2,
														"0"
													)}E${titleCard.episode?.episode_number
													?.toString()
													.padStart(2, "0")}`,
												latestMediaItem
											);
										if (titleCardResp === null) {
											setProgressValues((prev) => ({
												...prev,
												progressText: {
													...prev.progressText,
													[show.MediaItem.Title]: {
														...prev.progressText[show.MediaItem.Title],
														titlecard: `Failed to download Title Card for S${titleCard.episode.season_id.season_number
															?.toString()
															.padStart(
																2,
																"0"
															)}E${titleCard.episode?.episode_number
															?.toString()
															.padStart(2, "0")}`,
													},
												},
											}));
										} else {
											setProgressValues((prev) => ({
												...prev,
												progressText: {
													...prev.progressText,
													[show.MediaItem.Title]: {
														...prev.progressText[show.MediaItem.Title],
														titlecard: `Finished Title Card for S${titleCard.episode.season_id.season_number
															?.toString()
															.padStart(
																2,
																"0"
															)}E${titleCard.episode?.episode_number
															?.toString()
															.padStart(2, "0")} Download`,
													},
												},
											}));
										}
										setProgressValues((prev) => ({
											...prev,
											progressValue: prev.progressValue + progressIncrement,
										}));
									}

									break;
								default:
									break;
							}
						}

						const SaveItem: DBSavedItem = {
							MediaItemID: latestMediaItem.RatingKey,
							MediaItem: latestMediaItem,
							PosterSetID: show.id,
							PosterSet: {
								ID: show.id,
								Title: show.set_title,
								Type: "show",
								User: {
									Name: show.user_created.username,
								},
								DateCreated: show.date_created,
								DateUpdated: show.date_updated,
								Poster: {
									ID: show.show_poster[0]?.id,
									Type: "poster",
									Modified: show.show_poster[0]?.modified_on,
									FileSize: Number(show.show_poster[0]?.filesize),
								},
								Backdrop: {
									ID: show.show_backdrop[0]?.id,
									Type: "backdrop",
									Modified: show.show_backdrop[0]?.modified_on,
									FileSize: Number(show.show_backdrop[0]?.filesize),
								},
								SeasonPosters: show.season_posters.map((seasonPoster) => ({
									ID: seasonPoster.id,
									Type: "seasonPoster",
									Modified: seasonPoster.modified_on,
									FileSize: Number(seasonPoster.filesize),
									Season: {
										Number: seasonPoster.season.season_number,
									},
								})),
								TitleCards: show.titlecards.map((titleCard) => ({
									ID: titleCard.id,
									Type: "titlecard",
									Modified: titleCard.modified_on,
									FileSize: Number(titleCard.filesize),
									Episode: {
										Title: titleCard.episode.episode_title,
										EpisodeNumber: titleCard.episode?.episode_number,
										SeasonNumber: titleCard.episode.season_id.season_number,
									},
								})),
								Status: "",
							},
							LastDownloaded: new Date().toISOString(),
							SelectedTypes: itemData.types,
							AutoDownload: itemData.autodownload || false,
						};
						setProgressValues((prev) => ({
							...prev,
							progressText: {
								...prev.progressText,
								[show.MediaItem.Title]: {
									...prev.progressText[show.MediaItem.Title],
									addToDB: `Adding ${show.MediaItem.Title} to database`,
								},
							},
						}));
						const addToDBResp = await postAddItemToDB(SaveItem);
						if (addToDBResp.status !== "success") {
							setProgressValues((prev) => ({
								...prev,
								progressColor: "red",
								warningMessages: [
									`Failed to add ${show.MediaItem.Title} to database`,
								],
							}));
						} else {
							setProgressValues((prev) => ({
								...prev,
								progressValue: prev.progressValue + progressIncrement,
								progressText: {
									...prev.progressText,
									[show.MediaItem.Title]: {
										...prev.progressText[show.MediaItem.Title],
										addToDB: `Added ${show.MediaItem.Title} to database`,
									},
								},
							}));
						}

						// If the warning messages are empty, we can set the progress color to green
						if (
							(setProgressValues((prev) => ({
								...prev,
								progressColor: prev.warningMessages.length ? "yellow" : "green",
							})),
							!progressValues.warningMessages.length)
						) {
							setProgressValues((prev) => ({
								...prev,
								progressColor: "green",
							}));
						}
					}
				}
				setProgressValues((prev) => ({
					...prev,
					progressValue: 100,
				}));
			} else if (libraryType === "movie") {
				// Get all selected movie items
				const selectedMovies = selectedItems.filter(([itemKey]) =>
					boxset.movie_sets.some((movie) => movie.MediaItem.RatingKey === itemKey)
				);
				const selectedCollections = selectedItems.filter(([itemKey]) =>
					boxset.collection_sets.some((collection) =>
						collection.movie_posters.some(
							(poster) => poster.movie.MediaItem.RatingKey === itemKey
						)
					)
				);

				// Calculate steps for all selected movies and collections
				const totalSteps = calculateMovieSteps(selectedItems);
				const totalFileCount = calculateMovieTotalSteps(totalSteps);

				// Reserve 1% for start and 1% for completion
				const progressIncrement = totalFileCount > 0 ? 98 / totalFileCount : 0;
				setProgressValues((prev) => ({
					...prev,
					progressValue: 1,
				}));

				// Process selected movies first
				for (const [itemKey, itemData] of selectedMovies) {
					if (itemData.source !== "movie") {
						// Skip processing movies from collections here
						continue;
					}

					const movie = boxset.movie_sets.find((m) => m.MediaItem.RatingKey === itemKey);

					if (!movie || !itemData.types) {
						continue;
					}

					const selectedTypes = itemData.types.sort((a, b) => {
						const order = ["poster", "backdrop"];
						return order.indexOf(a) - order.indexOf(b);
					});
					if (selectedTypes.length === 0 && !itemData.addToDBOnly) {
						// No types selected, skip this movie
						continue;
					}
					for (const type of selectedTypes) {
						switch (type) {
							case "poster":
								if (itemData.addToDBOnly) continue;

								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										[movie.MediaItem.Title]: {
											...prev.progressText[movie.MediaItem.Title],
											poster: `Downloading Poster`,
										},
									},
								}));

								const posterFile: PosterFile = {
									ID: movie.movie_poster[0]?.id,
									Type: "poster",
									Modified: movie.movie_poster[0]?.modified_on,
									FileSize: Number(movie.movie_poster[0]?.filesize),
									Movie: {
										ID: movie.movie_id.id,
										Title: movie.MediaItem.Title,
										Status: movie.movie_id.status,
										Tagline: movie.movie_id.tagline,
										Slug: movie.movie_id.slug,
										DateUpdated: movie.movie_id.date_updated,
										TVbdID: movie.movie_id.tvdb_id,
										ImdbID: movie.movie_id.imdb_id,
										TraktID: movie.movie_id.trakt_id,
										ReleaseDate: movie.movie_id.release_date,
										RatingKey: movie.MediaItem.RatingKey,
										LibrarySection: movie.MediaItem.LibraryTitle,
									},
								};

								const posterResp = await downloadPosterFileAndUpdateMediaServer(
									posterFile,
									"Poster",
									movie.MediaItem
								);
								if (posterResp === null) {
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											[movie.MediaItem.Title]: {
												...prev.progressText[movie.MediaItem.Title],
												poster: `Failed to download Poster`,
											},
										},
									}));
								} else {
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											[movie.MediaItem.Title]: {
												...prev.progressText[movie.MediaItem.Title],
												poster: `Finished Poster Download`,
											},
										},
									}));
								}
								setProgressValues((prev) => ({
									...prev,
									progressValue: prev.progressValue + progressIncrement,
								}));
								break;
							case "backdrop":
								if (itemData.addToDBOnly) continue;
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										[movie.MediaItem.Title]: {
											...prev.progressText[movie.MediaItem.Title],
											backdrop: `Downloading Backdrop`,
										},
									},
								}));
								const backdropFile: PosterFile = {
									ID: movie.movie_backdrop[0]?.id,
									Type: "backdrop",
									Modified: movie.movie_backdrop[0]?.modified_on,
									FileSize: Number(movie.movie_backdrop[0]?.filesize),
									Movie: {
										ID: movie.movie_id.id,
										Title: movie.MediaItem.Title,
										Status: movie.movie_id.status,
										Tagline: movie.movie_id.tagline,
										Slug: movie.movie_id.slug,
										DateUpdated: movie.movie_id.date_updated,
										TVbdID: movie.movie_id.tvdb_id,
										ImdbID: movie.movie_id.imdb_id,
										TraktID: movie.movie_id.trakt_id,
										ReleaseDate: movie.movie_id.release_date,
										RatingKey: movie.MediaItem.RatingKey,
										LibrarySection: movie.MediaItem.LibraryTitle,
									},
								};
								const backdropResp = await downloadPosterFileAndUpdateMediaServer(
									backdropFile,
									"Backdrop",
									movie.MediaItem
								);
								if (backdropResp === null) {
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											[movie.MediaItem.Title]: {
												...prev.progressText[movie.MediaItem.Title],
												backdrop: `Failed to download Backdrop`,
											},
										},
									}));
								} else {
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											[movie.MediaItem.Title]: {
												...prev.progressText[movie.MediaItem.Title],
												backdrop: `Finished Backdrop Download`,
											},
										},
									}));
								}
								setProgressValues((prev) => ({
									...prev,
									progressValue: prev.progressValue + progressIncrement,
								}));
								break;
						}
					}

					const SaveItem: DBSavedItem = {
						MediaItemID: movie.MediaItem.RatingKey,
						MediaItem: movie.MediaItem,
						PosterSetID: movie.id,
						PosterSet: {
							ID: movie.id,
							Title: movie.set_title,
							Type: "movie",
							User: {
								Name: movie.user_created.username,
							},
							DateCreated: movie.date_created,
							DateUpdated: movie.date_updated,
							Poster: {
								ID: movie.movie_poster[0]?.id,
								Type: "poster",
								Modified: movie.movie_poster[0]?.modified_on,
								FileSize: Number(movie.movie_poster[0]?.filesize),
							},
							Backdrop: {
								ID: movie.movie_backdrop[0]?.id,
								Type: "backdrop",
								Modified: movie.movie_backdrop[0]?.modified_on,
								FileSize: Number(movie.movie_backdrop[0]?.filesize),
							},
							Status: "",
						},
						LastDownloaded: new Date().toISOString(),
						SelectedTypes: itemData.types,
						AutoDownload: itemData.autodownload || false,
					};
					setProgressValues((prev) => ({
						...prev,
						progressText: {
							...prev.progressText,
							[movie.MediaItem.Title]: {
								...prev.progressText[movie.MediaItem.Title],
								addToDB: `Adding to database`,
							},
						},
					}));
					const addToDBResp = await postAddItemToDB(SaveItem);
					if (addToDBResp.status !== "success") {
						setProgressValues((prev) => ({
							...prev,
							progressColor: "red",
							warningMessages: [`Failed to add to database`],
						}));
					} else {
						setProgressValues((prev) => ({
							...prev,
							progressValue: prev.progressValue + progressIncrement,
							progressText: {
								...prev.progressText,
								[movie.MediaItem.Title]: {
									...prev.progressText[movie.MediaItem.Title],
									addToDB: `Added to database`,
								},
							},
						}));
					}

					// If the warning messages are empty, we can set the progress color to green
					if (
						(setProgressValues((prev) => ({
							...prev,
							progressColor: prev.warningMessages.length ? "yellow" : "green",
						})),
						!progressValues.warningMessages.length)
					) {
						setProgressValues((prev) => ({
							...prev,
							progressColor: "green",
						}));
					}
				}

				// Process selected collections
				for (const [itemKey, itemData] of selectedCollections) {
					if (itemData.source !== "collection") {
						// Skip processing movies that are not from collections
						continue;
					}

					const collection = boxset.collection_sets.find((c) =>
						c.movie_posters.some((p) => p.movie.MediaItem.RatingKey === itemKey)
					);
					if (!collection || !itemData.types) {
						continue; // Skip if no collection found
					}

					const selectedTypes = itemData.types.sort((a, b) => {
						const order = ["poster", "backdrop"];
						return order.indexOf(a) - order.indexOf(b);
					});
					if (selectedTypes.length === 0 && !itemData.addToDBOnly) {
						// No types selected, skip this collection
						continue;
					}

					const collectionPoster = collection.movie_posters.find(
						(p) => p.movie.MediaItem.RatingKey === itemKey
					);
					const collectionBackdrop = collection.movie_backdrops.find(
						(p) => p.movie.MediaItem.RatingKey === itemKey
					);

					if (!collectionPoster && !collectionBackdrop) {
						continue; // Skip if no poster and no backdrop found
					}

					// Get the media item based on the poster or backdrop
					const thisMediaItem =
						collectionPoster?.movie.MediaItem || collectionBackdrop?.movie.MediaItem;
					if (!thisMediaItem) {
						continue; // Skip if no media item found
					}

					for (const type of selectedTypes) {
						switch (type) {
							case "poster":
								if (!collectionPoster || itemData.addToDBOnly) {
									continue; // Skip if no poster found
								}
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										[thisMediaItem.Title]: {
											...prev.progressText[thisMediaItem.Title],
											poster: `Downloading Poster`,
										},
									},
								}));
								const posterFile: PosterFile = {
									ID: collection.movie_posters[0]?.id,
									Type: "poster",
									Modified: collectionPoster.modified_on,
									FileSize: Number(collectionPoster.filesize),
									Movie: {
										ID: collectionPoster.movie.id,
										Title: collectionPoster.movie.MediaItem.Title,
										Status: collectionPoster.movie.status,
										Tagline: collectionPoster.movie.tagline,
										Slug: collectionPoster.movie.slug,
										DateUpdated: collectionPoster.movie.date_updated,
										TVbdID: collectionPoster.movie.tvdb_id,
										ImdbID: collectionPoster.movie.imdb_id,
										TraktID: collectionPoster.movie.trakt_id,
										ReleaseDate: collectionPoster.movie.release_date,
										RatingKey: collectionPoster.movie.MediaItem.RatingKey,
										LibrarySection:
											collectionPoster.movie.MediaItem.LibraryTitle,
									},
								};
								const posterResp = await downloadPosterFileAndUpdateMediaServer(
									posterFile,
									"Poster",
									collectionPoster.movie.MediaItem
								);
								if (posterResp === null) {
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											[thisMediaItem.Title]: {
												...prev.progressText[thisMediaItem.Title],
												poster: `Failed to download Poster`,
											},
										},
									}));
								} else {
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											[thisMediaItem.Title]: {
												...prev.progressText[thisMediaItem.Title],
												poster: `Finished Poster Download`,
											},
										},
									}));
								}
								setProgressValues((prev) => ({
									...prev,
									progressValue: prev.progressValue + progressIncrement,
								}));
								break;
							case "backdrop":
								if (!collectionBackdrop || itemData.addToDBOnly) {
									continue; // Skip if no backdrop found
								}
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										[thisMediaItem.Title]: {
											...prev.progressText[thisMediaItem.Title],
											backdrop: `Downloading Backdrop`,
										},
									},
								}));
								const backdropFile: PosterFile = {
									ID: collection.movie_backdrops[0]?.id,
									Type: "backdrop",
									Modified: collectionBackdrop.modified_on,
									FileSize: Number(collectionBackdrop.filesize),
									Movie: {
										ID: collectionBackdrop.movie.id,
										Title: collectionBackdrop.movie.MediaItem.Title,
										Status: collectionBackdrop.movie.status,
										Tagline: collectionBackdrop.movie.tagline,
										Slug: collectionBackdrop.movie.slug,
										DateUpdated: collectionBackdrop.movie.date_updated,
										TVbdID: collectionBackdrop.movie.tvdb_id,
										ImdbID: collectionBackdrop.movie.imdb_id,
										TraktID: collectionBackdrop.movie.trakt_id,
										ReleaseDate: collectionBackdrop.movie.release_date,
										RatingKey: collectionBackdrop.movie.MediaItem.RatingKey,
										LibrarySection:
											collectionBackdrop.movie.MediaItem.LibraryTitle,
									},
								};
								const backdropResp = await downloadPosterFileAndUpdateMediaServer(
									backdropFile,
									"Backdrop",
									collectionBackdrop.movie.MediaItem
								);
								if (backdropResp === null) {
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											[thisMediaItem.Title]: {
												...prev.progressText[thisMediaItem.Title],
												backdrop: `Failed to download Backdrop`,
											},
										},
									}));
								} else {
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											[thisMediaItem.Title]: {
												...prev.progressText[thisMediaItem.Title],
												backdrop: `Finished Backdrop Download`,
											},
										},
									}));
								}
								setProgressValues((prev) => ({
									...prev,
									progressValue: prev.progressValue + progressIncrement,
								}));
								break;
						}
					}

					if (
						!collectionPoster?.movie.MediaItem &&
						!collectionBackdrop?.movie.MediaItem &&
						!collectionPoster?.movie.MediaItem.RatingKey &&
						!collectionBackdrop?.movie.MediaItem.RatingKey
					) {
						continue; // Skip if no movie found in poster or backdrop
					}

					const SaveItem: DBSavedItem = {
						MediaItemID: thisMediaItem.RatingKey,
						MediaItem: thisMediaItem,
						PosterSetID: collection.id,
						PosterSet: {
							ID: collection.id,
							Title: collection.set_title,
							Type: "collection",
							User: {
								Name: collection.user_created.username,
							},
							DateCreated: collection.date_created,
							DateUpdated: collection.date_updated,
							Poster: collectionPoster
								? {
										ID: collectionPoster.id,
										Type: "poster",
										Modified: collectionPoster.modified_on,
										FileSize: Number(collectionPoster.filesize),
										Movie: {
											ID: collectionPoster.movie.id,
											Title: collectionPoster.movie.MediaItem.Title,
											Status: collectionPoster.movie.status,
											Tagline: collectionPoster.movie.tagline,
											Slug: collectionPoster.movie.slug,
											DateUpdated: collectionPoster.movie.date_updated,
											TVbdID: collectionPoster.movie.tvdb_id,
											ImdbID: collectionPoster.movie.imdb_id,
											TraktID: collectionPoster.movie.trakt_id,
											ReleaseDate: collectionPoster.movie.release_date,
											RatingKey: collectionPoster.movie.MediaItem.RatingKey,
											LibrarySection:
												collectionPoster.movie.MediaItem.LibraryTitle,
										},
									}
								: undefined,
							Backdrop: collectionBackdrop
								? {
										ID: collectionBackdrop.id,
										Type: "backdrop",
										Modified: collectionBackdrop.modified_on,
										FileSize: Number(collectionBackdrop.filesize),
										Movie: {
											ID: collectionBackdrop.movie.id,
											Title: collectionBackdrop.movie.MediaItem.Title,
											Status: collectionBackdrop.movie.status,
											Tagline: collectionBackdrop.movie.tagline,
											Slug: collectionBackdrop.movie.slug,
											DateUpdated: collectionBackdrop.movie.date_updated,
											TVbdID: collectionBackdrop.movie.tvdb_id,
											ImdbID: collectionBackdrop.movie.imdb_id,
											TraktID: collectionBackdrop.movie.trakt_id,
											ReleaseDate: collectionBackdrop.movie.release_date,
											RatingKey: collectionBackdrop.movie.MediaItem.RatingKey,
											LibrarySection:
												collectionBackdrop.movie.MediaItem.LibraryTitle,
										},
									}
								: undefined,
							Status: "",
						},
						LastDownloaded: new Date().toISOString(),
						SelectedTypes: itemData.types,
						AutoDownload: itemData.autodownload || false,
					};
					setProgressValues((prev) => ({
						...prev,
						progressText: {
							...prev.progressText,
							[thisMediaItem.Title]: {
								...prev.progressText[thisMediaItem.Title],
								addToDB: `Adding to database`,
							},
						},
					}));
					const addToDBResp = await postAddItemToDB(SaveItem);
					if (addToDBResp.status !== "success") {
						setProgressValues((prev) => ({
							...prev,
							progressColor: "red",
							warningMessages: [`Failed to add to database`],
						}));
					} else {
						setProgressValues((prev) => ({
							...prev,
							progressValue: prev.progressValue + progressIncrement,
							progressText: {
								...prev.progressText,
								[thisMediaItem.Title]: {
									...prev.progressText[thisMediaItem.Title],
									addToDB: `Added to database`,
								},
							},
						}));
					}

					// If the warning messages are empty, we can set the progress color to green
					if (
						(setProgressValues((prev) => ({
							...prev,
							progressColor: prev.warningMessages.length ? "yellow" : "green",
						})),
						!progressValues.warningMessages.length)
					) {
						setProgressValues((prev) => ({
							...prev,
							progressColor: "green",
						}));
					}
				}
				setProgressValues((prev) => ({
					...prev,
					progressValue: 100,
				}));
			}
		} catch (error) {
			log("Poster Set Modal - Download Error", error);
			setProgressValues((prev) => ({
				...prev,
				progressColor: "red",
				warningMessages: ["An error occurred while downloading the files."],
			}));
		} finally {
			setIsMounted(false);
			setCancelButtonText("Close");
			setDownloadButtonText("Download Again");
		}
	};

	const renderShowFields = (show: MediuxUserShowSet) => (
		<FormField
			key={show.MediaItem.RatingKey}
			control={form.control}
			name={`selectedTypesByItem.${show.MediaItem.RatingKey}`}
			render={({ field }) => (
				<div className="rounded-md border p-4 rounded mb-4">
					<FormLabel className="text-md font-normal mb-4">
						{show.MediaItem.Title} ({show.MediaItem.Year})
					</FormLabel>
					<div className="space-y-2">
						{show.show_poster && show.show_poster.length > 0 && (
							<FormItem className="flex flex-row items-start space-x-2 space-y-0">
								<FormControl>
									<Checkbox
										checked={field.value?.types?.includes("poster")}
										onCheckedChange={(checked) => {
											const currentTypes = field.value?.types || [];
											field.onChange({
												...field.value,
												types: checked
													? [...currentTypes, "poster"]
													: currentTypes.filter(
															(type) => type !== "poster"
														),
											});
										}}
										className="h-5 w-5 sm:h-4 sm:w-4"
									/>
								</FormControl>
								<FormLabel className="text-md font-normal">Poster</FormLabel>
							</FormItem>
						)}

						{show.show_backdrop && show.show_backdrop.length > 0 && (
							<FormItem className="flex flex-row items-start space-x-2 space-y-0">
								<FormControl>
									<Checkbox
										checked={field.value?.types?.includes("backdrop")}
										onCheckedChange={(checked) => {
											const currentTypes = field.value?.types || [];
											field.onChange({
												...field.value,
												types: checked
													? [...currentTypes, "backdrop"]
													: currentTypes.filter(
															(type) => type !== "backdrop"
														),
											});
										}}
										className="h-5 w-5 sm:h-4 sm:w-4"
									/>
								</FormControl>
								<FormLabel className="text-md font-normal">Backdrop</FormLabel>
							</FormItem>
						)}
						{show.season_posters && show.season_posters.length > 0 && (
							<FormItem className="flex flex-row items-start space-x-2 space-y-0">
								<FormControl>
									<Checkbox
										checked={field.value?.types?.includes("seasonPoster")}
										onCheckedChange={(checked) => {
											const currentTypes = field.value?.types || [];
											field.onChange({
												...field.value,
												types: checked
													? [...currentTypes, "seasonPoster"]
													: currentTypes.filter(
															(type) => type !== "seasonPoster"
														),
											});
										}}
										className="h-5 w-5 sm:h-4 sm:w-4"
									/>
								</FormControl>
								<FormLabel className="text-md font-normal">
									Season Posters
								</FormLabel>
							</FormItem>
						)}
						{show.season_posters.some(
							(seasonPoster) => seasonPoster.season.season_number === 0
						) && (
							<FormItem className="flex flex-row items-start space-x-2 space-y-0">
								<FormControl>
									<Checkbox
										checked={field.value?.types?.includes(
											"specialSeasonPoster"
										)}
										onCheckedChange={(checked) => {
											const currentTypes = field.value?.types || [];
											field.onChange({
												...field.value,
												types: checked
													? [...currentTypes, "specialSeasonPoster"]
													: currentTypes.filter(
															(type) => type !== "specialSeasonPoster"
														),
											});
										}}
										className="h-5 w-5 sm:h-4 sm:w-4"
									/>
								</FormControl>
								<FormLabel className="text-md font-normal">
									Special Season Posters
								</FormLabel>
							</FormItem>
						)}

						{show.titlecards && show.titlecards.length > 0 && (
							<FormItem className="flex flex-row items-start space-x-2 space-y-0">
								<FormControl>
									<Checkbox
										checked={field.value?.types?.includes("titlecard")}
										onCheckedChange={(checked) => {
											const currentTypes = field.value?.types || [];
											field.onChange({
												...field.value,
												types: checked
													? [...currentTypes, "titlecard"]
													: currentTypes.filter(
															(type) => type !== "titlecard"
														),
											});
										}}
										className="h-5 w-5 sm:h-4 sm:w-4"
									/>
								</FormControl>
								<FormLabel className="text-md font-normal">Title Cards</FormLabel>
							</FormItem>
						)}

						<FormLabel className="text-md text-gray-500 dark:text-gray-400 mt-4">
							Download Options
						</FormLabel>

						{/* Add auto-download option */}
						<FormItem className="flex flex-row items-start space-x-2 space-y-0">
							<FormControl>
								<Checkbox
									// If futureUpdatesOnly is checked, autodownload should be checked
									checked={
										field.value?.futureUpdatesOnly ||
										field.value?.autodownload ||
										false
									}
									onCheckedChange={(checked) => {
										field.onChange({
											...field.value,
											autodownload: checked,
										});
									}}
									className="h-5 w-5 sm:h-4 sm:w-4"
									// Disable auto-download if future updates only is checked
									disabled={field.value?.futureUpdatesOnly || false}
								/>
							</FormControl>
							<FormLabel className="text-md font-normal">Auto Download</FormLabel>
							<div className="ml-auto">
								<Popover modal={true}>
									<PopoverTrigger className="cursor-pointer">
										<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
											?
										</span>
									</PopoverTrigger>
									<PopoverContent className="w-60">
										Auto Download will check periodically for new updates to
										this set. This is helpful if you want to download and apply
										titlecard updates from future updates to this set.
									</PopoverContent>
								</Popover>
							</div>
						</FormItem>

						{/* Add future updates only option */}
						<FormItem className="flex flex-row items-start space-x-2 space-y-0">
							<FormControl>
								<Checkbox
									checked={field.value?.futureUpdatesOnly || false}
									onCheckedChange={(checked) => {
										field.onChange({
											...field.value,
											futureUpdatesOnly: checked,
										});
										// If future updates only is checked, check autodownload
										if (checked) {
											form.setValue(
												`selectedTypesByItem.${show.MediaItem.RatingKey}.autodownload`,
												true
											);
										}
									}}
									className="h-5 w-5 sm:h-4 sm:w-4"
								/>
							</FormControl>
							<FormLabel className="text-md font-normal">
								Future Updates Only
							</FormLabel>
							<div className="ml-auto">
								<Popover modal={true}>
									<PopoverTrigger className="cursor-pointer">
										<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
											?
										</span>
									</PopoverTrigger>
									<PopoverContent className="w-60">
										Future Updates Only will not download anything right now.
										This is helpful if you have already downloaded the set and
										just want to future updates to be applied.
									</PopoverContent>
								</Popover>
							</div>
						</FormItem>
					</div>
				</div>
			)}
		/>
	);

	const isInBothSets = (ratingKey: string) => {
		const inMovieSet = boxset.movie_sets.some((m) => m.MediaItem.RatingKey === ratingKey);
		const inCollectionSet = boxset.collection_sets.some((collection) =>
			collection.movie_posters.some(
				(poster) => poster.movie.MediaItem.RatingKey === ratingKey
			)
		);
		return inMovieSet && inCollectionSet;
	};

	const isSelectedInMovieSet = (ratingKey: string) => {
		const movieSet = boxset.movie_sets.find((m) => m.MediaItem.RatingKey === ratingKey);
		if (!movieSet) return false;

		const item = watchSelectedTypesByItem[ratingKey];
		return item?.source === "movie" && (item?.types?.length ?? 0) > 0;
	};

	const isSelectedInCollectionSet = (ratingKey: string) => {
		const collectionMovie = boxset.collection_sets.some((collection) =>
			collection.movie_posters.some(
				(poster) => poster.movie.MediaItem.RatingKey === ratingKey
			)
		);
		if (!collectionMovie) return false;

		const item = watchSelectedTypesByItem[ratingKey];
		return item?.source === "collection" && (item?.types?.length ?? 0) > 0;
	};

	const renderMovieFields = (movie: MediuxUserMovieSet) => (
		<FormField
			key={movie.MediaItem.RatingKey}
			control={form.control}
			name={`selectedTypesByItem.${movie.MediaItem.RatingKey}`}
			render={({ field }) => (
				<div className="rounded-md border p-4 rounded mb-4">
					<FormLabel className="text-md font-normal mb-4">
						{movie.MediaItem.Title} ({movie.MediaItem.Year})
						{isInBothSets(movie.MediaItem.RatingKey) && (
							<Popover modal={true}>
								<PopoverTrigger>
									<TriangleAlert className="h-4 w-4 text-yellow-500 cursor-help" />
								</PopoverTrigger>
								<PopoverContent className="w-60">
									{isSelectedInCollectionSet(movie.MediaItem.RatingKey)
										? "This movie is currently selected in the Collection Set"
										: "This movie exists in both a Movie Set and a Collection Set"}
								</PopoverContent>
							</Popover>
						)}
					</FormLabel>
					<div className="space-y-2">
						{movie.movie_poster && movie.movie_poster.length > 0 && (
							<FormItem className="flex flex-row items-start space-x-2 space-y-0">
								<FormControl>
									<Checkbox
										checked={
											isSelectedInCollectionSet(movie.MediaItem.RatingKey)
												? false
												: field.value?.types?.includes("poster")
										}
										disabled={isSelectedInCollectionSet(
											movie.MediaItem.RatingKey
										)}
										onCheckedChange={(checked) => {
											const currentTypes = field.value?.types || [];
											// If checking this item, uncheck the collection version
											if (checked) {
												const collectionKey = movie.MediaItem.RatingKey;
												form.setValue(
													`selectedTypesByItem.${collectionKey}`,
													{
														types: [],
														autodownload: false,
														futureUpdatesOnly: false,
														source: "collection",
													}
												);
											}
											field.onChange({
												...field.value,
												types: checked
													? [...currentTypes, "poster"]
													: currentTypes.filter(
															(type) => type !== "poster"
														),
												autodownload: false,
												futureUpdatesOnly: false,
												source: "movie",
											});
										}}
										className="h-5 w-5 sm:h-4 sm:w-4"
									/>
								</FormControl>
								<FormLabel
									className={`text-md font-normal ${
										isSelectedInCollectionSet(movie.MediaItem.RatingKey)
											? "text-muted-foreground"
											: ""
									}`}
								>
									Poster
								</FormLabel>
							</FormItem>
						)}
						{movie.movie_backdrop && movie.movie_backdrop.length > 0 && (
							<FormItem className="flex flex-row items-start space-x-2 space-y-0">
								<FormControl>
									<Checkbox
										checked={
											isSelectedInCollectionSet(movie.MediaItem.RatingKey)
												? false
												: field.value?.types?.includes("backdrop")
										}
										disabled={isSelectedInCollectionSet(
											movie.MediaItem.RatingKey
										)}
										onCheckedChange={(checked) => {
											const currentTypes = field.value?.types || [];
											// If checking this item, uncheck the collection version
											if (checked) {
												const collectionKey = movie.MediaItem.RatingKey;
												form.setValue(
													`selectedTypesByItem.${collectionKey}`,
													{
														types: [],
														autodownload: false,
														futureUpdatesOnly: false,
														source: "collection",
													}
												);
											}
											field.onChange({
												...field.value,
												types: checked
													? [...currentTypes, "backdrop"]
													: currentTypes.filter(
															(type) => type !== "backdrop"
														),
												autodownload: false,
												futureUpdatesOnly: false,
												source: "movie",
											});
										}}
										className="h-5 w-5 sm:h-4 sm:w-4"
									/>
								</FormControl>
								<FormLabel
									className={`text-md font-normal ${
										isSelectedInCollectionSet(movie.MediaItem.RatingKey)
											? "text-muted-foreground"
											: ""
									}`}
								>
									Backdrop
								</FormLabel>
							</FormItem>
						)}

						{/* Add to Database Only Option */}
						<FormItem className="flex flex-row items-start space-x-2 space-y-0">
							<FormControl>
								<Checkbox
									checked={
										isSelectedInCollectionSet(movie.MediaItem.RatingKey)
											? false
											: field.value.addToDBOnly || false
									}
									disabled={isSelectedInCollectionSet(movie.MediaItem.RatingKey)}
									onCheckedChange={(checked) => {
										// If checking this item, uncheck the collection version
										if (checked) {
											const collectionKey = movie.MediaItem.RatingKey;
											form.setValue(`selectedTypesByItem.${collectionKey}`, {
												types: [],
												autodownload: false,
												futureUpdatesOnly: false,
												source: "collection",
												addToDBOnly: false,
											});
										}
										field.onChange({
											...field.value,
											addToDBOnly: checked,
											source: "movie",
										});
									}}
									className="h-5 w-5 sm:h-4 sm:w-4"
								/>
							</FormControl>
							<FormLabel
								className={`text-md font-normal ${
									isSelectedInCollectionSet(movie.MediaItem.RatingKey)
										? "text-muted-foreground"
										: ""
								}`}
							>
								Add to Database Only
							</FormLabel>
						</FormItem>
					</div>
				</div>
			)}
		/>
	);

	const renderCollectionFields = (collection: MediuxUserCollectionSet) => {
		const moviesMap = collection.movie_posters.reduce(
			(acc, poster) => {
				const movieKey = poster.movie.MediaItem.RatingKey;
				if (!acc[movieKey]) {
					acc[movieKey] = {
						title: poster.movie.title,
						year: new Date(poster.movie.release_date).getFullYear(),
						poster: poster,
						backdrop: collection.movie_backdrops.find(
							(b) => b.movie.id === poster.movie.id
						),
					};
				}
				return acc;
			},
			{} as Record<
				string,
				{
					title: string;
					year: number;
					poster: MediuxUserCollectionMovie;
					backdrop?: MediuxUserCollectionMovie;
				}
			>
		);

		return Object.entries(moviesMap).map(([movieKey, movie]) => (
			<FormField
				key={movieKey}
				control={form.control}
				name={`selectedTypesByItem.${movieKey}`}
				render={({ field }) => {
					const isMovieSetSelected = isSelectedInMovieSet(movieKey);

					return (
						<div className="rounded-md border p-4 rounded mb-4">
							<FormLabel className="text-md font-normal mb-4">
								{movie.title} ({movie.year}){" "}
								{isInBothSets(movieKey) && (
									<Popover modal={true}>
										<PopoverTrigger>
											<TriangleAlert className="h-4 w-4 text-yellow-500 cursor-help" />
										</PopoverTrigger>
										<PopoverContent className="w-60">
											{isMovieSetSelected
												? "This movie is currently selected in the Movie Set"
												: "This movie exists in both a Movie Set and a Collection Set"}
										</PopoverContent>
									</Popover>
								)}
							</FormLabel>
							<div className="space-y-2">
								<FormItem className="flex flex-row items-start space-x-2 space-y-0">
									<FormControl>
										<Checkbox
											checked={
												isSelectedInMovieSet(movieKey)
													? false
													: field.value?.types?.includes("poster")
											}
											disabled={isMovieSetSelected}
											onCheckedChange={(checked) => {
												const currentTypes = field.value?.types || [];
												// If checking this item, uncheck the movie version
												if (checked) {
													form.setValue(
														`selectedTypesByItem.${movieKey}`,
														{
															types: [],
															autodownload: false,
															futureUpdatesOnly: false,
															source: "movie",
														}
													);
												}
												field.onChange({
													...field.value,
													types: checked
														? [...currentTypes, "poster"]
														: currentTypes.filter(
																(type) => type !== "poster"
															),
													autodownload: false,
													futureUpdatesOnly: false,
													source: "collection",
												});
											}}
											className="h-5 w-5 sm:h-4 sm:w-4"
										/>
									</FormControl>
									<FormLabel
										className={`text-md font-normal ${
											isMovieSetSelected ? "text-muted-foreground" : ""
										}`}
									>
										Poster
									</FormLabel>
								</FormItem>

								{movie.backdrop && (
									<FormItem className="flex flex-row items-start space-x-2 space-y-0">
										<FormControl>
											<Checkbox
												checked={
													isSelectedInMovieSet(movieKey)
														? false
														: field.value?.types?.includes("backdrop")
												}
												disabled={isMovieSetSelected}
												onCheckedChange={(checked) => {
													const currentTypes = field.value?.types || [];
													// If checking this item, uncheck the movie version
													if (checked) {
														form.setValue(
															`selectedTypesByItem.${movieKey}`,
															{
																types: [],
																autodownload: false,
																futureUpdatesOnly: false,
																source: "movie",
															}
														);
													}
													field.onChange({
														...field.value,
														types: checked
															? [...currentTypes, "backdrop"]
															: currentTypes.filter(
																	(type) => type !== "backdrop"
																),
														autodownload: false,
														futureUpdatesOnly: false,
														source: "collection",
													});
												}}
												className="h-5 w-5 sm:h-4 sm:w-4"
											/>
										</FormControl>
										<FormLabel
											className={`text-md font-normal ${
												isMovieSetSelected ? "text-muted-foreground" : ""
											}`}
										>
											Backdrop
										</FormLabel>
									</FormItem>
								)}

								{/* Add to Database Only Option */}
								<FormItem className="flex flex-row items-start space-x-2 space-y-0">
									<FormControl>
										<Checkbox
											checked={
												isSelectedInMovieSet(movieKey)
													? false
													: field.value.addToDBOnly || false
											}
											disabled={isSelectedInMovieSet(movieKey)}
											onCheckedChange={(checked) => {
												// If checking this item, uncheck the movie version
												if (checked) {
													form.setValue(
														`selectedTypesByItem.${movieKey}`,
														{
															types: [],
															autodownload: false,
															futureUpdatesOnly: false,
															source: "movie",
															addToDBOnly: false,
														}
													);
												}
												field.onChange({
													...field.value,
													addToDBOnly: checked,
													source: "collection",
												});
											}}
											className="h-5 w-5 sm:h-4 sm:w-4"
										/>
									</FormControl>
									<FormLabel
										className={`text-md font-normal ${
											isMovieSetSelected ? "text-muted-foreground" : ""
										}`}
									>
										Add to Database Only
									</FormLabel>
								</FormItem>
							</div>
						</div>
					);
				}}
			/>
		));
	};

	return (
		<Dialog
			onOpenChange={(open) => {
				if (!open) {
					handleClose();
				}
			}}
		>
			<DialogTrigger asChild>
				<Download className="mr-2 h-5 w-5 sm:h-7 sm:w-7" />
			</DialogTrigger>
			<DialogPortal>
				<DialogOverlay />
				<DialogContent className="overflow-y-auto max-h-[80vh] sm:max-w-[425px] ">
					<DialogHeader>
						<DialogTitle>{boxset.boxset_title}</DialogTitle>
						<DialogDescription>
							BoxSet Author: {boxset.user_created.username}
						</DialogDescription>
						<DialogDescription>
							<Link
								href={`https://mediux.pro/boxsets/${boxset.id}`}
								className="hover:text-primary transition-colors text-sm text-muted-foreground"
								target="_blank"
								rel="noopener noreferrer"
							>
								BoxSet ID: {boxset.id}
							</Link>
						</DialogDescription>
					</DialogHeader>

					<Form {...form}>
						<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-2">
							{boxset.movie_sets && boxset.movie_sets.length > 0 && (
								<>
									<h3 className="text-lg font-semibold mb-2">Movie Sets</h3>
									{boxset.movie_sets.map((movie) => renderMovieFields(movie))}
								</>
							)}

							{boxset.collection_sets && boxset.collection_sets.length > 0 && (
								<>
									<h3 className="text-lg font-semibold mb-2">Collection Sets</h3>
									{boxset.collection_sets.map((collection) =>
										renderCollectionFields(collection)
									)}
								</>
							)}

							{boxset.show_sets && boxset.show_sets.length > 0 && (
								<>
									<h3 className="text-lg font-semibold mb-2">Show Sets</h3>
									{boxset.show_sets.map((show) => renderShowFields(show))}
								</>
							)}

							{/* Form Message for validation errors */}
							<FormMessage />

							{/* Total Size of Selected Types */}
							{totalSelectedSize && (
								<div className="text-sm text-muted-foreground">
									{totalSelectedSizeLabel} {totalSelectedSize}
								</div>
							)}

							{/* Progress Bar */}
							{progressValues.progressValue > 0 && (
								<div className="w-full">
									<div className="flex items-center justify-between">
										<Progress
											value={progressValues.progressValue}
											className={`flex-1 ${
												progressValues.progressColor
													? `[&>div]:bg-${progressValues.progressColor}-500`
													: ""
											}`}
										/>
										<span className="ml-2 text-sm text-muted-foreground">
											{Math.round(progressValues.progressValue)}%
										</span>
									</div>

									{/* Progress text display */}
									{Object.entries(progressValues.progressText).map(
										([itemTitle, statuses]) => (
											<div key={itemTitle} className="space-y-1">
												<div className="font-semibold text-sm">
													{itemTitle}
												</div>
												{(() => {
													const typeMap = {
														poster: "Poster",
														backdrop: "Backdrop",
														seasonPoster: "Season Poster",
														specialSeasonPoster:
															"Special Season Poster",
														titlecard: "Title Card",
														addToDB: "Added to DB",
													};
													return Object.entries(statuses).map(
														([type, status]) => (
															<div
																key={`${itemTitle}-${type}`}
																className="flex justify-between text-sm text-muted-foreground"
															>
																{status.startsWith("Finished") ||
																status.startsWith("Added") ? (
																	<div className="flex items-center">
																		<Check className="mr-1 h-4 w-4" />
																		<span>
																			{typeMap[
																				type as keyof typeof typeMap
																			] || type}
																		</span>
																	</div>
																) : status.startsWith("Failed") ? (
																	<div className="flex items-center text-destructive">
																		<X className="mr-1 h-4 w-4" />
																		<span>{status}</span>
																	</div>
																) : (
																	<div className="flex items-center">
																		<LoaderIcon className="mr-1 h-4 w-4 animate-spin" />
																		<span>{status}</span>
																	</div>
																)}
															</div>
														)
													);
												})()}
											</div>
										)
									)}
								</div>
							)}

							{/* Warning Messages */}
							{progressValues.warningMessages.length > 0 && (
								<div className="my-2">
									<Accordion type="single" collapsible>
										<AccordionItem value="warnings">
											<AccordionTrigger className="text-destructive">
												Failed Downloads (
												{progressValues.warningMessages.length})
											</AccordionTrigger>
											<AccordionContent>
												<div className="flex flex-col space-y-2">
													{progressValues.warningMessages.map(
														(message) => (
															<div
																key={message}
																className="flex items-center text-destructive"
															>
																<X className="mr-1 h-4 w-4" />
																<span>{message}</span>
															</div>
														)
													)}
												</div>
											</AccordionContent>
										</AccordionItem>
									</Accordion>
								</div>
							)}

							<DialogFooter>
								<div className="flex space-x-4">
									{/* Cancel button to close the modal */}
									<DialogClose asChild>
										<Button
											className=""
											variant="destructive"
											onClick={() => {
												handleClose();
											}}
										>
											{cancelButtonText}
										</Button>
									</DialogClose>

									{/* Download button to display download info */}
									<Button
										className=""
										disabled={
											!Object.values(watchSelectedTypesByItem).some(
												(item) =>
													item?.types &&
													Array.isArray(item.types) &&
													item.types.length > 0
											)
										}
										onClick={() => {
											onSubmit(form.getValues());
										}}
									>
										{downloadButtonText}
									</Button>
								</div>
							</DialogFooter>
						</form>
					</Form>
				</DialogContent>
			</DialogPortal>
		</Dialog>
	);
};

export default DownloadModalBoxset;
