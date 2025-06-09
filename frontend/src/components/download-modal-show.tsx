"use client";

import { formatDownloadSize } from "@/helper/formatDownloadSize";
import { postAddItemToDB } from "@/services/api.db";
import {
	fetchMediaServerItemContent,
	patchDownloadPosterFileAndUpdateMediaServer,
} from "@/services/api.mediaserver";
import { fetchShowSetByID } from "@/services/api.mediux";
import { zodResolver } from "@hookform/resolvers/zod";
import { Check, Download, LoaderIcon, TriangleAlert, X } from "lucide-react";
import { z } from "zod";

import { useEffect, useMemo, useState } from "react";
import { useForm, useWatch } from "react-hook-form";

import Link from "next/link";

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

import { log } from "@/lib/logger";

import { DBSavedItem } from "@/types/databaseSavedSet";
import { MediaItem } from "@/types/mediaItem";
import { PosterFile, PosterSet } from "@/types/posterSets";

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "./ui/accordion";
import { Button } from "./ui/button";
import { Checkbox } from "./ui/checkbox";
import { Popover, PopoverContent, PopoverTrigger } from "./ui/popover";
import { Progress } from "./ui/progress";

const formSchema = z.object({
	selectedTypes: z.array(z.string()).refine((value) => value.length > 0, {
		message: "You must select at least one asset type.",
	}),
});

const DownloadModalShow: React.FC<{
	posterSet: PosterSet;
	mediaItem: MediaItem;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	autoDownloadDefault?: boolean;
	forceSetRefresh?: boolean; // Optional prop to force a refresh of the set
}> = ({ posterSet, mediaItem, open, onOpenChange, autoDownloadDefault, forceSetRefresh }) => {
	const [isMounted, setIsMounted] = useState(false);
	const [cancelButtonText, setCancelButtonText] = useState("Cancel");
	const [downloadButtonText, setDownloadButtonText] = useState("Download");
	const [autoDownload, setAutoDownload] = useState(autoDownloadDefault || false);
	const [futureUpdatesOnly, setFutureUpdatesOnly] = useState(false);

	// Tracking selected checkboxes for what to download
	const [totalSelectedSize, setTotalSelectedSize] = useState("");
	const [totalSelectedSizeLabel, setTotalSelectedSizeLabel] = useState("Total Download Size: ");
	const [currentPosterSet, setCurrentPosterSet] = useState<PosterSet>(posterSet);

	const handleAutoDownloadChange = () => {
		if (futureUpdatesOnly) {
			// Auto Download must remain true if Future Updates Only is selected
			setAutoDownload(true);
		} else {
			setAutoDownload((prev) => !prev);
		}
	};
	const handleFutureUpdatesOnlyChange = () => {
		setFutureUpdatesOnly((prev) => {
			const newValue = !prev;
			if (newValue) {
				setAutoDownload(true);
			}
			return newValue;
		});
	};

	// Download Progress
	const [progressValues, setProgressValues] = useState<{
		progressValue: number;
		progressColor: string;
		progressText: {
			poster: string;
			backdrop: string;
			seasonPoster: string;
			specialSeasonPoster: string;
			titleCard: string;
			addToDB: string;
		};
		warningMessages: string[];
	}>({
		progressValue: 0,
		progressColor: "",
		progressText: {
			poster: "",
			backdrop: "",
			seasonPoster: "",
			specialSeasonPoster: "",
			titleCard: "",
			addToDB: "",
		},
		warningMessages: [],
	});

	useEffect(() => {
		// If forceSetRefresh is true, we need to get the latest poster set from the server
		if (forceSetRefresh) {
			log("Poster Set Modal - Force Refresh Triggered");
			const getShowSetByID = async () => {
				const resp = await fetchShowSetByID(posterSet.ID);
				if (resp.status !== "success") {
					return;
				}

				// Update the posterSet state with the latest data
				if (resp.data) {
					setCurrentPosterSet(resp.data);
					log(
						"Poster Set Modal - Poster Set Updated:",
						"Before:",
						posterSet,
						"After:",
						resp.data
					);
				}
			};
			getShowSetByID();
		} else {
			// If not forcing a refresh, use the passed posterSet directly
			log("Poster Set Modal - Using passed posterSet data");
			setCurrentPosterSet(posterSet);
		}
	}, [forceSetRefresh, posterSet]);

	// Compute asset options
	const assetTypes = useMemo(() => {
		if (!currentPosterSet) {
			return [];
		}
		const setHasPoster = currentPosterSet.Poster;
		const setHasBackdrop = currentPosterSet.Backdrop;
		const setHasSeasonPosters = (currentPosterSet.SeasonPosters?.length ?? 0) > 0;
		const setHasSpecialSeasonPosters = currentPosterSet.SeasonPosters?.some(
			(season) => season.Type === "specialSeasonPoster"
		);
		const setHasTitleCards = (currentPosterSet.TitleCards?.length ?? 0) > 0;

		return [
			setHasPoster ? { id: "poster", label: "Poster" } : null,
			setHasBackdrop ? { id: "backdrop", label: "Backdrop" } : null,
			setHasSeasonPosters ? { id: "seasonPoster", label: "Season Poster" } : null,
			setHasSpecialSeasonPosters
				? { id: "specialSeasonPoster", label: "Special Poster" }
				: null,
			setHasTitleCards ? { id: "titlecard", label: "Titlecard" } : null,
		].filter(Boolean);
	}, [currentPosterSet]);

	const form = useForm<z.infer<typeof formSchema>>({
		resolver: zodResolver(formSchema),
		defaultValues: {
			selectedTypes: assetTypes
				.filter((item): item is { id: string; label: string } => item !== null)
				.map((item) => item.id),
		},
	});

	const watchSelectedTypes = useWatch({
		control: form.control,
		name: "selectedTypes",
	});

	// Calculate the total size of selected types
	useEffect(() => {
		if (!currentPosterSet) {
			return;
		}
		const selectedTypes = watchSelectedTypes || [];
		log("Poster Set Modal - Selected Types:", selectedTypes);
		const totalSize = selectedTypes.reduce((acc, type) => {
			let size = 0;
			switch (type) {
				case "poster":
					size = currentPosterSet.Poster?.FileSize || 0;
					break;
				case "backdrop":
					size = currentPosterSet.Backdrop?.FileSize || 0;
					break;
				case "seasonPoster":
					size =
						currentPosterSet.SeasonPosters?.reduce(
							(s, sp) => s + (sp.FileSize || 0),
							0
						) || 0;
					break;
				case "specialSeasonPoster":
					size =
						currentPosterSet.SeasonPosters?.filter(
							(season) => season.Type === "specialSeasonPoster"
						).reduce((s, sp) => s + (sp.FileSize || 0), 0) || 0;
					break;
				case "titlecard":
					size =
						currentPosterSet.TitleCards?.reduce((s, tc) => s + (tc.FileSize || 0), 0) ||
						0;
					break;
				default:
					size = 0;
			}
			return acc + size;
		}, 0);
		setTotalSelectedSize(formatDownloadSize(totalSize));
	}, [watchSelectedTypes, currentPosterSet]);

	if (!currentPosterSet) {
		return (
			<Dialog
				open={open}
				onOpenChange={(isOpen) => {
					if (typeof onOpenChange === "function") {
						onOpenChange(isOpen);
					}
					if (!isOpen) {
						handleClose();
					}
				}}
			>
				{!open && (
					<DialogTrigger asChild>
						<Download className="mr-2 h-5 w-5 sm:h-7 sm:w-7" />
					</DialogTrigger>
				)}
				<DialogPortal>
					<DialogOverlay />
					<DialogContent className="max-w-md">
						<DialogHeader>
							<DialogTitle>Loading Poster Set...</DialogTitle>
							<DialogDescription>This may take a moment.</DialogDescription>
						</DialogHeader>
						<div className="flex justify-center items-center">
							<LoaderIcon className="h-8 w-8 animate-spin" />
						</div>
					</DialogContent>
				</DialogPortal>
			</Dialog>
		);
	}

	const resetProgressValues = () => {
		setProgressValues({
			progressValue: 0,
			progressColor: "",
			progressText: {
				poster: "",
				backdrop: "",
				seasonPoster: "",
				specialSeasonPoster: "",
				titleCard: "",
				addToDB: "",
			},
			warningMessages: [],
		});
	};

	const handleClose = () => {
		setCancelButtonText("Close");
		setDownloadButtonText("Download");
		resetProgressValues();
		setIsMounted(false);
		setTotalSelectedSizeLabel("Total Download Size: ");
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

	// Function to handle form submission
	const onSubmit = async (data: z.infer<typeof formSchema>) => {
		if (isMounted) return;
		if (!currentPosterSet) return;

		setIsMounted(true);
		setCancelButtonText("Cancel");
		setDownloadButtonText("Downloading...");
		resetProgressValues();

		try {
			// Using the Selected Types, download the files and update the media server
			const selectedTypes = data.selectedTypes || [];

			// Get the latest Media Item details from the server
			const latestMediaItemResp = await fetchMediaServerItemContent(
				mediaItem.RatingKey,
				mediaItem.LibraryTitle
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

			if (futureUpdatesOnly) {
				const SaveItem: DBSavedItem = {
					MediaItemID: latestMediaItem.RatingKey,
					MediaItem: latestMediaItem,
					PosterSetID: currentPosterSet.ID,
					PosterSet: currentPosterSet,
					LastDownloaded: new Date().toISOString(),
					SelectedTypes: data.selectedTypes,
					AutoDownload: autoDownload,
				};
				setProgressValues((prev) => ({
					...prev,
					progressValue: 50,
					progressText: {
						...prev.progressText,
						addToDB: "Adding item to database",
					},
				}));
				const addToDBResp = await postAddItemToDB(SaveItem);
				if (addToDBResp.status !== "success") {
					setProgressValues((prev) => ({
						...prev,
						progressColor: "red",
						progressText: {
							...prev.progressText,
							addToDB: "Failed to add item to database",
						},
					}));
				} else {
					setProgressValues((prev) => ({
						...prev,
						progressText: {
							...prev.progressText,
							addToDB: "Finished adding item to database",
						},
					}));
				}
				setProgressValues((prev) => ({
					...prev,
					progressValue: 100,
				}));
				setCancelButtonText("Close");
				setDownloadButtonText("Download Again");
				return;
			} else {
				// Calculate the number of files to download based on selected types
				// This will be used to update the progress bar in increments
				const totalFilesToDownload = selectedTypes.reduce((acc, type) => {
					switch (type) {
						case "poster":
							return acc + 1;
						case "backdrop":
							return acc + 1;
						case "seasonPoster":
						case "specialSeasonPoster":
							return acc + (currentPosterSet.SeasonPosters?.length || 0);
						case "titlecard":
							return acc + (currentPosterSet.TitleCards?.length || 0);
						default:
							return acc;
					}
				}, 0);
				const progressIncrement = 95 / totalFilesToDownload;

				// Sort Selected Types to ensure the order of download
				// Poster, Backdrop, SeasonPoster, TitleCard
				// This is to ensure that the progress bar updates in the correct order
				selectedTypes.sort((a, b) => {
					const order = [
						"poster",
						"backdrop",
						"seasonPoster",
						"specialSeasonPoster",
						"titlecard",
					];
					return order.indexOf(a) - order.indexOf(b);
				});

				setProgressValues((prev) => ({
					...prev,
					progressValue: 1,
				}));

				for (const type of selectedTypes) {
					switch (type) {
						case "poster":
							if (!currentPosterSet.Poster) {
								throw new Error("Poster file is missing");
							}
							setProgressValues((prev) => ({
								...prev,
								progressText: {
									...prev.progressText,
									poster: "Downloading Poster",
								},
							}));
							const posterResp = await downloadPosterFileAndUpdateMediaServer(
								currentPosterSet.Poster,
								"Poster",
								latestMediaItem
							);
							if (posterResp === null) {
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										poster: "Failed to download Poster",
									},
								}));
							} else {
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										poster: "Finished Poster Download",
									},
								}));
							}
							setProgressValues((prev) => ({
								...prev,
								progressValue: prev.progressValue + progressIncrement,
							}));
							break;
						case "backdrop":
							if (!currentPosterSet.Backdrop) {
								throw new Error("Backdrop file is missing");
							}
							setProgressValues((prev) => ({
								...prev,
								progressText: {
									...prev.progressText,
									backdrop: "Downloading Backdrop",
								},
							}));
							const backdropResp = await downloadPosterFileAndUpdateMediaServer(
								currentPosterSet.Backdrop,
								"Backdrop",
								latestMediaItem
							);
							if (backdropResp === null) {
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										backdrop: "Failed to download Backdrop",
									},
								}));
							} else {
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										backdrop: "Finished Backdrop Download",
									},
								}));
							}
							setProgressValues((prev) => ({
								...prev,
								progressValue: prev.progressValue + progressIncrement,
							}));
							break;
						case "seasonPoster":
							let seasonErrorCount = 0;
							for (const season of currentPosterSet.SeasonPosters || []) {
								if (season.Season?.Number === 0) {
									// Skip season 0
									continue;
								}

								// Check to see if the season is present in the MediaItem
								// Use the season.Season.Number
								// Check MediaItem.Series.Season.Number
								const seasonExists = latestMediaItem.Series?.Seasons?.some(
									(seasonItem) =>
										seasonItem.SeasonNumber === season.Season?.Number
								);
								if (!seasonExists) {
									continue;
								}
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										seasonPoster: `Downloading Season Poster ${season.Season?.Number?.toString().padStart(
											2,
											"0"
										)}`,
									},
								}));

								const seasonResp = await downloadPosterFileAndUpdateMediaServer(
									season,
									`S${String(season.Season?.Number).padStart(2, "0")} Poster`,
									latestMediaItem
								);
								if (seasonResp === null) {
									seasonErrorCount++;
								}
								setProgressValues((prev) => ({
									...prev,
									progressValue: prev.progressValue + progressIncrement,
								}));
							}
							if (seasonErrorCount > 0) {
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										seasonPoster: `Failed to download ${seasonErrorCount} Season Poster${
											seasonErrorCount > 1 ? "s" : ""
										}`,
									},
								}));
							} else {
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										seasonPoster: `Finished Season Poster${
											(currentPosterSet.SeasonPosters?.length ?? 0) > 1
												? "s"
												: ""
										} Download`,
									},
								}));
							}
							break;
						case "specialSeasonPoster":
							for (const season of currentPosterSet.SeasonPosters || []) {
								if (season.Season?.Number !== 0) {
									// Skip season 0
									continue;
								}

								// Check to see if the season is present in the MediaItem
								// Use the season.Season.Number
								// Check MediaItem.Series.Season.Number
								const seasonExists = latestMediaItem.Series?.Seasons?.some(
									(seasonItem) =>
										seasonItem.SeasonNumber === season.Season?.Number
								);
								if (!seasonExists) {
									continue;
								}
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										specialSeasonPoster: `Downloading Special Season Poster ${season.Season?.Number?.toString().padStart(
											2,
											"0"
										)}`,
									},
								}));
								const seasonResp = await downloadPosterFileAndUpdateMediaServer(
									season,
									`S${String(season.Season?.Number).padStart(2, "0")} Poster`,
									latestMediaItem
								);
								if (seasonResp === null) {
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											specialSeasonPoster: `Failed to download Special Season Poster`,
										},
									}));
								} else {
									setProgressValues((prev) => ({
										...prev,
										progressText: {
											...prev.progressText,
											specialSeasonPoster: `Finished Special Season Poster Download`,
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
							let titleCardErrorCount = 0;
							for (const titleCard of currentPosterSet.TitleCards || []) {
								// Check to see if the episode is present in the MediaItem
								// Use the titlecard.Episode.SeasonNumber and titlecard.Episode.EpisodeNumber
								const episode = latestMediaItem.Series?.Seasons?.flatMap(
									(season) => season.Episodes
								).find(
									(episode) =>
										episode.SeasonNumber === titleCard.Episode?.SeasonNumber &&
										episode.EpisodeNumber === titleCard.Episode?.EpisodeNumber
								);
								if (!episode) {
									log(
										"Poster Set Modal - Title Card Skipped: Episode not found in MediaItem"
									);
									continue;
								}
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										titleCard: `Downloading S${String(
											titleCard.Episode?.SeasonNumber
										).padStart(2, "0")}E${String(
											titleCard.Episode?.EpisodeNumber
										).padStart(2, "0")} Titlecard`,
									},
								}));
								const titlecardResp = await downloadPosterFileAndUpdateMediaServer(
									titleCard,
									`S${String(titleCard.Episode?.SeasonNumber).padStart(
										2,
										"0"
									)}E${String(titleCard.Episode?.EpisodeNumber).padStart(
										2,
										"0"
									)} Titlecard`,
									latestMediaItem
								);
								if (titlecardResp === null) {
									titleCardErrorCount++;
								}
								setProgressValues((prev) => ({
									...prev,
									progressValue: prev.progressValue + progressIncrement,
								}));
							}
							if (titleCardErrorCount > 0) {
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										titleCard: `Failed to download ${titleCardErrorCount} Title Card${
											titleCardErrorCount > 1 ? "s" : ""
										}`,
									},
								}));
							} else {
								setProgressValues((prev) => ({
									...prev,
									progressText: {
										...prev.progressText,
										titleCard: `Finished Title Card${
											(currentPosterSet.TitleCards?.length ?? 0) > 1
												? "s"
												: ""
										} Download`,
									},
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
					PosterSetID: currentPosterSet.ID,
					PosterSet: currentPosterSet,
					LastDownloaded: new Date().toISOString(),
					SelectedTypes: data.selectedTypes,
					AutoDownload: autoDownload,
				};
				setProgressValues((prev) => ({
					...prev,
					progressValue: 100,
					progressText: {
						...prev.progressText,
						addToDB: "Adding item to database",
					},
				}));
				const addToDBResp = await postAddItemToDB(SaveItem);
				if (addToDBResp.status !== "success") {
					setProgressValues((prev) => ({
						...prev,
						progressColor: "red",
						progressText: {
							...prev.progressText,
							addToDB: "Failed to add item to database",
						},
					}));
				} else {
					setProgressValues((prev) => ({
						...prev,
						progressText: {
							...prev.progressText,
							addToDB: "Finished adding item to database",
						},
					}));
					mediaItem.ExistInDatabase = true;
				}
			}

			if (progressValues.warningMessages.length > 0) {
				setProgressValues((prev) => ({
					...prev,
					progressValue: 100,
					progressColor: "yellow",
				}));
			} else {
				setProgressValues((prev) => ({
					...prev,
					progressValue: 100,
					progressColor: "green",
				}));
			}
		} catch (error) {
			log("Poster Set Modal - Download Error:", error);
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

	return (
		<Dialog
			open={open}
			onOpenChange={(isOpen) => {
				if (typeof onOpenChange === "function") {
					onOpenChange(isOpen);
				}
				if (!isOpen) {
					handleClose();
				}
			}}
		>
			{!open && (
				<DialogTrigger asChild>
					<Download className="mr-2 h-5 w-5 sm:h-7 sm:w-7" />
				</DialogTrigger>
			)}
			<DialogPortal>
				<DialogOverlay />
				<DialogContent className="overflow-y-auto max-h-[80vh] sm:max-w-[425px] ">
					<DialogHeader>
						<DialogTitle>{currentPosterSet.Title}</DialogTitle>
						<DialogDescription>
							Set Author: {currentPosterSet.User.Name}
						</DialogDescription>
						<DialogDescription>
							<Link
								href={`https://mediux.pro/sets/${currentPosterSet.ID}`}
								className="hover:text-primary transition-colors text-sm text-muted-foreground"
								target="_blank"
								rel="noopener noreferrer"
							>
								Set ID: {currentPosterSet.ID}
							</Link>
						</DialogDescription>
					</DialogHeader>

					<Form {...form}>
						<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-2">
							<FormField
								control={form.control}
								name="selectedTypes"
								render={() => (
									<FormItem className="rounded-md border p-4 shadow">
										<FormLabel className="text-md font-normal">
											Select Assets to Download
										</FormLabel>
										{assetTypes
											.filter(
												(
													item
												): item is {
													id: string;
													label: string;
												} => item !== null
											)
											.map((item) => (
												<FormField
													key={item.id}
													control={form.control}
													name="selectedTypes"
													render={({ field }) => (
														<FormItem className="flex flex-row items-start space-x-3 space-y-0">
															<FormControl>
																<Checkbox
																	checked={field.value?.includes(
																		item.id
																	)}
																	onCheckedChange={(checked) => {
																		return checked
																			? field.onChange([
																					...field.value,
																					item.id,
																				])
																			: field.onChange(
																					field.value?.filter(
																						(v) =>
																							v !==
																							item.id
																					)
																				);
																	}}
																	className="h-5 w-5 sm:h-4 sm:w-4" // Larger on mobile, normal on sm+
																/>
															</FormControl>
															<FormLabel className="text-md font-normal">
																{item.label}
															</FormLabel>
														</FormItem>
													)}
												/>
											))}
										<FormMessage />
									</FormItem>
								)}
							/>

							{/* Auto Download & Future Updates Only Check Boxes */}
							{totalSelectedSize && (
								<FormItem className="rounded-md border p-4 shadow">
									<FormLabel className="text-md font-normal">
										Download Options
									</FormLabel>
									<div className="flex flex-col space-y-2">
										<FormItem className="flex flex-row items-start space-x-3 space-y-0">
											<FormControl>
												<Checkbox
													checked={autoDownload}
													onCheckedChange={handleAutoDownloadChange}
													disabled={futureUpdatesOnly}
													className="h-5 w-5 sm:h-4 sm:w-4" // Larger on mobile, normal on sm+
												/>
											</FormControl>
											<FormLabel className="text-md font-normal">
												Auto Download
											</FormLabel>
											<div className="ml-auto">
												<Popover modal={true}>
													<PopoverTrigger className="cursor-pointer">
														<span className="text-gray-500 dark:text-gray-400 cursor-pointer">
															?
														</span>
													</PopoverTrigger>
													<PopoverContent className="w-60">
														Auto Download will check periodically for
														new updates to this set. This is helpful if
														you want to download and apply titlecard
														updates from future updates to this set.
													</PopoverContent>
												</Popover>
											</div>
										</FormItem>
										<FormItem className="flex flex-row items-start space-x-3 space-y-0">
											<FormControl>
												<Checkbox
													checked={futureUpdatesOnly}
													onCheckedChange={handleFutureUpdatesOnlyChange}
													className="h-5 w-5 sm:h-4 sm:w-4" // Larger on mobile, normal on sm+
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
														Future Updates Only will not download
														anything right now. This is helpful if you
														have already downloaded the set and just
														want to future updates to be applied.
													</PopoverContent>
												</Popover>
											</div>
										</FormItem>
									</div>
									<FormMessage />
								</FormItem>
							)}

							{/* Total Size of Selected Types */}
							{totalSelectedSize && (
								<div className="text-sm text-muted-foreground">
									{totalSelectedSizeLabel} {totalSelectedSize}
								</div>
							)}

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

									{progressValues.progressText.poster && (
										<div className="flex justify-between text-sm text-muted-foreground ">
											{progressValues.progressText.poster.startsWith(
												"Finished"
											) ? (
												<div className="flex items-center">
													<Check className="mr-1 h-4 w-4" />
													<span>Poster</span>
												</div>
											) : progressValues.progressText.poster.startsWith(
													"Failed"
											  ) ? (
												<div className="flex items-center text-destructive">
													<X className="mr-1 h-4 w-4" />
													<span>
														{progressValues.progressText.poster}
													</span>
												</div>
											) : (
												<div className="flex items-center">
													<LoaderIcon className="mr-1 h-4 w-4 animate-spin" />
													<span>
														{progressValues.progressText.poster}
													</span>
												</div>
											)}
										</div>
									)}
									{progressValues.progressText.backdrop && (
										<div className="flex justify-between text-sm text-muted-foreground ">
											{progressValues.progressText.backdrop.startsWith(
												"Finished"
											) ? (
												<div className="flex items-center">
													<Check className="mr-1 h-4 w-4" />
													<span>Backdrop</span>
												</div>
											) : progressValues.progressText.backdrop.startsWith(
													"Failed"
											  ) ? (
												<div className="flex items-center text-destructive">
													<X className="mr-1 h-4 w-4" />
													<span>
														{progressValues.progressText.backdrop}
													</span>
												</div>
											) : (
												<div className="flex items-center">
													<LoaderIcon className="mr-1 h-4 w-4 animate-spin" />
													<span>
														{progressValues.progressText.backdrop}
													</span>
												</div>
											)}
										</div>
									)}
									{progressValues.progressText.seasonPoster && (
										<div className="flex justify-between text-sm text-muted-foreground ">
											<span>
												{progressValues.progressText.seasonPoster.startsWith(
													"Finished"
												) ? (
													<div className="flex items-center">
														<Check className="mr-1 h-4 w-4" />
														<span>Season Poster</span>
													</div>
												) : progressValues.progressText.seasonPoster.startsWith(
														"Failed"
												  ) ? (
													<div className="flex items-center text-destructive">
														{(() => {
															const total =
																currentPosterSet.SeasonPosters
																	?.length ?? 0;
															const text =
																progressValues.progressText.seasonPoster
																	.trim()
																	.toLowerCase();
															const single = `${total} season poster`;
															const plural = `${total} season posters`;
															return text.endsWith(single) ||
																text.endsWith(plural) ? (
																<X className="mr-1 h-4 w-4" />
															) : (
																<TriangleAlert className="mr-1 h-4 w-4" />
															);
														})()}
														<span>
															{
																progressValues.progressText
																	.seasonPoster
															}
														</span>
													</div>
												) : (
													<div className="flex items-center">
														<LoaderIcon className="mr-1 h-4 w-4 animate-spin" />
														<span>
															{
																progressValues.progressText
																	.seasonPoster
															}
														</span>
													</div>
												)}
											</span>
										</div>
									)}

									{progressValues.progressText.specialSeasonPoster && (
										<div className="flex justify-between text-sm text-muted-foreground ">
											{progressValues.progressText.specialSeasonPoster.startsWith(
												"Finished"
											) ? (
												<div className="flex items-center">
													<Check className="mr-1 h-4 w-4" />
													<span>Special Season Poster</span>
												</div>
											) : progressValues.progressText.specialSeasonPoster.startsWith(
													"Failed"
											  ) ? (
												<div className="flex items-center text-destructive">
													<X className="mr-1 h-4 w-4" />
													<span>
														{
															progressValues.progressText
																.specialSeasonPoster
														}
													</span>
												</div>
											) : (
												<div className="flex items-center">
													<LoaderIcon className="mr-1 h-4 w-4 animate-spin" />
													<span>
														{
															progressValues.progressText
																.specialSeasonPoster
														}
													</span>
												</div>
											)}
										</div>
									)}

									{progressValues.progressText.titleCard && (
										<div className="flex justify-between text-sm text-muted-foreground ">
											<span>
												{progressValues.progressText.titleCard.startsWith(
													"Finished"
												) ? (
													<div className="flex items-center">
														<Check className="mr-1 h-4 w-4" />
														<span>Title Cards</span>
													</div>
												) : progressValues.progressText.titleCard.startsWith(
														"Failed"
												  ) ? (
													<div className="flex items-center text-destructive">
														{(() => {
															const total =
																currentPosterSet.TitleCards
																	?.length ?? 0;
															const text =
																progressValues.progressText.titleCard
																	.trim()
																	.toLowerCase();
															const single = `${total} titlecard`;
															const plural = `${total} titlecards`;
															return text.endsWith(single) ||
																text.endsWith(plural) ? (
																<X className="mr-1 h-4 w-4" />
															) : (
																<TriangleAlert className="mr-1 h-4 w-4" />
															);
														})()}
														<span>
															{progressValues.progressText.titleCard}
														</span>
													</div>
												) : (
													<div className="flex items-center">
														<LoaderIcon className="mr-1 h-4 w-4 animate-spin" />
														<span>
															{progressValues.progressText.titleCard}
														</span>
													</div>
												)}
											</span>
										</div>
									)}

									{progressValues.progressText.addToDB && (
										<div className="flex justify-between text-sm text-muted-foreground ">
											{progressValues.progressText.addToDB.startsWith(
												"Finished"
											) ? (
												<div className="flex items-center">
													<Check className="mr-1 h-4 w-4" />
													<span>Added to DB</span>
												</div>
											) : progressValues.progressText.addToDB.startsWith(
													"Failed"
											  ) ? (
												<div className="flex items-center text-destructive">
													<X className="mr-1 h-4 w-4" />
													<span>
														{progressValues.progressText.addToDB}
													</span>
												</div>
											) : (
												<div className="flex items-center">
													<LoaderIcon className="mr-1 h-4 w-4 animate-spin" />
													<span>
														{progressValues.progressText.addToDB}
													</span>
												</div>
											)}
										</div>
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
										disabled={watchSelectedTypes.length === 0}
										className=""
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

export default DownloadModalShow;
