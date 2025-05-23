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
import { MediaItem } from "@/types/mediaItem";
import { PosterFile, PosterSet } from "@/types/posterSets";
import { Check, Download, LoaderIcon, X } from "lucide-react";
import Link from "next/link";
import { Button } from "./ui/button";
import { useEffect, useMemo, useState } from "react";
import { Checkbox } from "./ui/checkbox";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { z } from "zod";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { log } from "@/lib/logger";
import { patchDownloadPosterFileAndUpdateMediaServer } from "@/services/api.mediaserver";
import { Progress } from "./ui/progress";
import { formatDownloadSize } from "@/helper/formatDownloadSize";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "./ui/accordion";
import { DBSavedItem } from "@/types/databaseSavedSet";
import { openDB } from "idb";
import { CACHE_DB_NAME, CACHE_STORE_NAME } from "@/constants/cache";
import { postAddItemToDB } from "@/services/api.db";

const formSchema = z
	.object({
		selectedTypesByMovie: z.record(z.array(z.string())),
	})
	.refine(
		(data) =>
			Object.values(data.selectedTypesByMovie).some(
				(arr) => Array.isArray(arr) && arr.length > 0
			),
		{
			message:
				"Please select at least one poster or backdrop to download.",
			path: ["selectedTypesByMovie"],
		}
	);

const DownloadModalMovie: React.FC<{
	posterSet: PosterSet;
	mediaItem: MediaItem;
}> = ({ posterSet, mediaItem }) => {
	const [isMounted, setIsMounted] = useState(false);
	const [cancelButtonText, setCancelButtonText] = useState("Close");
	const [downloadButtonText, setDownloadButtonText] = useState("Download");

	// Tracking selected checkboxes for what to download
	const [totalSelectedSize, setTotalSelectedSize] = useState("");
	const [totalSelectedSizeLabel, setTotalSelectedSizeLabel] = useState(
		"Total Download Size: "
	);

	// Download Progress
	const [progressValue, setProgressValue] = useState(0);
	const [progressColor, setProgressColor] = useState("");
	// Track download status for each movie and type (poster/backdrop)
	const [progressDownloads, setProgressDownloads] = useState<
		Record<
			string, // Movie name or key
			{
				poster?: "downloading" | "success" | "failed";
				backdrop?: "downloading" | "success" | "failed";
				addToDB?: "adding" | "success" | "failed";
			}
		>
	>({});
	const [progressWarningMessages, setProgressWarningMessages] = useState<
		string[]
	>([]);

	// Create a map of Movies within the Poster Set
	// This is used to display the movies in the set in the modal
	// and to download the files for each movie
	// Movie Interface
	interface MovieDisplay {
		MediaItemRatingKey: string;
		MediaItem: MediaItem;
		SetID: string;
		Poster?: PosterFile;
		Backdrop?: PosterFile;
		Title: string;
		Year: number;
		MainItem: boolean;
	}

	const moviesDisplay: MovieDisplay[] = useMemo(() => {
		const movies: MovieDisplay[] = [];
		if (posterSet.Poster) {
			const existingMovie = movies.find(
				(m) =>
					m.MediaItemRatingKey === posterSet.Poster?.Movie?.RatingKey
			);
			if (existingMovie) {
				existingMovie.Poster = posterSet.Poster;
			} else {
				const movie: MovieDisplay = {
					MediaItemRatingKey: mediaItem.RatingKey,
					MediaItem: {
						RatingKey: mediaItem.RatingKey,
						LibraryTitle: mediaItem.LibraryTitle,
						Type: mediaItem.Type,
						Title: "",
						Year: 0,
						ExistInDatabase: false,
						Guids: [],
					},
					SetID: posterSet.ID,
					Poster: posterSet.Poster,
					Backdrop: posterSet.Backdrop,
					Title: mediaItem.Title,
					Year: mediaItem.Year,
					MainItem: true,
				};
				movies.push(movie);
			}
		}
		if (posterSet.Backdrop) {
			const existingMovie = movies.find(
				(m) => m.MediaItemRatingKey === mediaItem.RatingKey
			);
			if (existingMovie) {
				existingMovie.Backdrop = posterSet.Backdrop;
			} else {
				const movie: MovieDisplay = {
					MediaItemRatingKey: mediaItem.RatingKey,
					MediaItem: {
						Type: mediaItem.Type,
						RatingKey: mediaItem.RatingKey,
						LibraryTitle: mediaItem.LibraryTitle,
						Title: "",
						Year: 0,
						ExistInDatabase: false,
						Guids: [],
					},
					SetID: posterSet.ID,
					Poster: posterSet.Poster,
					Backdrop: posterSet.Backdrop,
					Title: mediaItem.Title,
					Year: mediaItem.Year,
					MainItem: true,
				};
				movies.push(movie);
			}
		}
		if (posterSet.OtherPosters) {
			posterSet.OtherPosters.forEach((poster: PosterFile) => {
				if (poster.Movie && poster.Movie.RatingKey) {
					const existingMovie = movies.find(
						(m) =>
							poster.Movie &&
							m.MediaItemRatingKey === poster.Movie.RatingKey
					);
					if (existingMovie) {
						existingMovie.Poster = poster;
					} else {
						const movie: MovieDisplay = {
							MediaItemRatingKey: poster.Movie.RatingKey,
							MediaItem: {
								Type: mediaItem.Type,
								RatingKey: poster.Movie.RatingKey,
								LibraryTitle: mediaItem.LibraryTitle,
								Title: "",
								Year: 0,
								ExistInDatabase: false,
								Guids: [],
							},
							SetID: posterSet.ID,
							Poster: poster,
							Title: poster.Movie.Title,
							Year: poster.Movie.ReleaseDate
								? new Date(
										poster.Movie.ReleaseDate
								  ).getFullYear()
								: 0,
							MainItem: false,
						};
						movies.push(movie);
					}
				}
			});
		}
		if (posterSet.OtherBackdrops) {
			posterSet.OtherBackdrops.forEach((backdrop: PosterFile) => {
				if (backdrop.Movie && backdrop.Movie.RatingKey) {
					const existingMovie = movies.find(
						(m) =>
							backdrop.Movie &&
							m.MediaItemRatingKey === backdrop.Movie.RatingKey
					);
					if (existingMovie) {
						existingMovie.Backdrop = backdrop;
					} else {
						const movie: MovieDisplay = {
							MediaItemRatingKey: backdrop.Movie.RatingKey,
							MediaItem: {
								Type: mediaItem.Type,
								RatingKey: backdrop.Movie.RatingKey,
								LibraryTitle: mediaItem.LibraryTitle,
								Title: "",
								Year: 0,
								ExistInDatabase: false,
								Guids: [],
							},
							SetID: posterSet.ID,
							Backdrop: backdrop,
							Title: backdrop.Movie.Title,
							Year: backdrop.Movie.ReleaseDate
								? new Date(
										backdrop.Movie.ReleaseDate
								  ).getFullYear()
								: 0,
							MainItem: false,
						};
						movies.push(movie);
					}
				}
			});
		}
		return movies;
	}, [mediaItem, posterSet]);

	const form = useForm<z.infer<typeof formSchema>>({
		resolver: zodResolver(formSchema),
		defaultValues: {
			selectedTypesByMovie: moviesDisplay.reduce((acc, movie) => {
				const types: string[] = [];
				if (movie.Poster) types.push("poster");
				if (movie.Backdrop) types.push("backdrop");
				acc[movie.MediaItemRatingKey] = types;
				return acc;
			}, {} as Record<string, string[]>),
		},
	});

	const watchSelectionsByMovie = useWatch({
		control: form.control,
		name: "selectedTypesByMovie",
	});

	// Calculate the total size of selected types for all movies
	useEffect(() => {
		const totalSize = Object.entries(watchSelectionsByMovie).reduce(
			(acc, [movieKey, selectedTypes]) => {
				const movie = moviesDisplay.find(
					(m) => m.MediaItemRatingKey === movieKey
				);
				if (movie) {
					const selectedPoster = movie.Poster?.FileSize || 0;
					const selectedBackdrop = movie.Backdrop?.FileSize || 0;
					const selectedSize = selectedTypes.reduce((size, type) => {
						switch (type) {
							case "poster":
								return size + selectedPoster;
							case "backdrop":
								return size + selectedBackdrop;
							default:
								return size;
						}
					}, 0);
					return acc + selectedSize;
				}
				return acc;
			},
			0
		);
		setTotalSelectedSize(formatDownloadSize(totalSize));
	}, [watchSelectionsByMovie, moviesDisplay]);

	const handleClose = () => {
		setCancelButtonText("Close");
		setDownloadButtonText("Download");
		setProgressValue(0);
		setIsMounted(false);
		setProgressDownloads({});
		setProgressWarningMessages([]);
		setProgressColor("");
		setTotalSelectedSizeLabel("Total Download Size: ");
		form.reset();
	};

	const downloadFileAndUpdateMediaServer = async (
		posterFile: PosterFile,
		posterFileName: string,
		posterMediaItem: MediaItem
	) => {
		try {
			const resp = await patchDownloadPosterFileAndUpdateMediaServer(
				posterFile,
				posterMediaItem
			);
			if (resp.status !== "success") {
				throw new Error(`Failed to download ${posterFileName}`);
			} else {
				return resp.data;
			}
		} catch {
			setProgressWarningMessages((prev) => [...prev, posterFileName]);
			setProgressColor("yellow");
			return null;
		}
	};

	async function getMediaItemDetails(key: string) {
		const db = await openDB(CACHE_DB_NAME, 1, {
			upgrade(db) {
				if (!db.objectStoreNames.contains(CACHE_STORE_NAME)) {
					db.createObjectStore(CACHE_STORE_NAME);
				}
			},
		});

		// Get the library section details from the cache
		const librarySection = await db.get(
			CACHE_STORE_NAME,
			mediaItem.LibraryTitle
		);
		if (!librarySection) {
			return undefined;
		}
		// Get the media item details from the cache
		const mediaItems = librarySection.data.MediaItems;
		return mediaItems.find((item: MediaItem) => item.RatingKey === key);
	}

	// Function to handle form submission
	const onSubmit = async (data: z.infer<typeof formSchema>) => {
		if (isMounted) return;
		setIsMounted(true);
		setCancelButtonText("Cancel");
		setProgressValue(0);
		setProgressDownloads({});
		setProgressWarningMessages([]);
		setProgressColor("");

		try {
			// Calculate the number of files to download based on selected types
			// This will be used to update the progress bar in increments
			const totalFilesToDownload = Object.entries(
				data.selectedTypesByMovie
			).reduce((acc, [movieKey]) => {
				const movie = moviesDisplay.find(
					(m) => m.MediaItemRatingKey === movieKey
				);
				if (movie) {
					const selectedPoster = movie.Poster ? 1 : 0;
					const selectedBackdrop = movie.Backdrop ? 1 : 0;
					return acc + selectedPoster + selectedBackdrop;
				}
				return acc;
			}, 0);

			const progressIncrement = 90 / totalFilesToDownload;

			setProgressValue(1);

			// Sort the movies to display main item first
			const orderedMovies: MovieDisplay[] = [
				...moviesDisplay.filter((movie) => movie.MainItem),
				...moviesDisplay.filter((movie) => !movie.MainItem),
			];

			for (const movie of orderedMovies) {
				const selectedTypesByMovie =
					data.selectedTypesByMovie[movie.MediaItemRatingKey];
				if (
					!selectedTypesByMovie ||
					selectedTypesByMovie.length === 0
				) {
					continue; // Skip if no types are selected
				}
				if (movie && movie.MediaItem && movie.MediaItem.RatingKey) {
					if (movie && movie.MediaItem && movie.MediaItem.RatingKey) {
						if (
							selectedTypesByMovie.includes("poster") &&
							movie.Poster
						) {
							setProgressDownloads((prev) => ({
								...prev,
								[movie.Title]: {
									...prev[movie.Title],
									poster: "downloading",
								},
							}));
							const posterResp =
								await downloadFileAndUpdateMediaServer(
									movie.Poster,
									`Poster for ${movie.Title} (${movie.Year})`,
									movie.MediaItem
								);
							if (posterResp === null) {
								setProgressDownloads((prev) => ({
									...prev,
									[movie.Title]: {
										...prev[movie.Title],
										poster: "failed",
									},
								}));
							} else {
								setProgressDownloads((prev) => ({
									...prev,
									[movie.Title]: {
										...prev[movie.Title],
										poster: "success",
									},
								}));
								setProgressValue(
									(prev) => prev + progressIncrement
								);
							}
						}
						if (
							selectedTypesByMovie.includes("backdrop") &&
							movie.Backdrop
						) {
							setProgressDownloads((prev) => ({
								...prev,
								[movie.Title]: {
									...prev[movie.Title],
									backdrop: "downloading",
								},
							}));
							const backdropResp =
								await downloadFileAndUpdateMediaServer(
									movie.Backdrop,
									`Backdrop for ${movie.Title} (${movie.Year})`,
									movie.MediaItem
								);
							if (backdropResp === null) {
								setProgressDownloads((prev) => ({
									...prev,
									[movie.Title]: {
										...prev[movie.Title],
										backdrop: "failed",
									},
								}));
							} else {
								setProgressDownloads((prev) => ({
									...prev,
									[movie.Title]: {
										...prev[movie.Title],
										backdrop: "success",
									},
								}));
								setProgressValue(
									(prev) => prev + progressIncrement
								);
							}
						}
						// Get the media item details from the cache

						const mediaDetails = await getMediaItemDetails(
							movie.MediaItemRatingKey
						);

						const SaveItem: DBSavedItem = {
							MediaItemID: movie.MediaItemRatingKey ?? "",
							MediaItem: mediaDetails,
							PosterSetID: posterSet.ID,
							PosterSet: posterSet,
							LastDownloaded: new Date().toISOString(),
							SelectedTypes: selectedTypesByMovie,
							AutoDownload: false,
						};
						setProgressDownloads((prev) => ({
							...prev,
							[movie.Title]: {
								...prev[movie.Title],
								addToDB: "adding",
							},
						}));
						const addToDBResp = await postAddItemToDB(SaveItem);
						if (addToDBResp.status !== "success") {
							setProgressDownloads((prev) => ({
								...prev,
								[movie.Title]: {
									...prev[movie.Title],
									addToDB: "failed",
								},
							}));
						} else {
							setProgressDownloads((prev) => ({
								...prev,
								[movie.Title]: {
									...prev[movie.Title],
									addToDB: "success",
								},
							}));
							mediaItem.ExistInDatabase = true;
						}
					}
				}
			}

			if (progressWarningMessages.length > 0) {
				setProgressColor("yellow");
			} else {
				setProgressColor("green");
			}

			setProgressValue(100);
		} catch (error) {
			log("Poster Set Modal - Download Error:", error);
			setProgressColor("red");
			setProgressWarningMessages(() => [
				"An error occurred while downloading the files.",
			]);
		} finally {
			setIsMounted(false);
		}
	};

	// Inside your component:
	const renderMovieField = (movie: MovieDisplay) => (
		<FormField
			key={movie.MediaItemRatingKey}
			control={form.control}
			name={`selectedTypesByMovie.${movie.MediaItemRatingKey}`}
			render={({ field }) => (
				<div className="rounded-md border p-4 rounded mb-4">
					<FormLabel className="text-md font-normal mb-4">
						{movie.Title} ({movie.Year})
					</FormLabel>
					<div className="space-y-2">
						{movie.Poster && (
							<FormItem className="flex flex-row items-start space-x-2 space-y-0">
								<FormControl>
									<Checkbox
										checked={field.value.includes("poster")}
										onCheckedChange={(checked) => {
											if (checked) {
												field.onChange([
													...field.value,
													"poster",
												]);
											} else {
												field.onChange(
													field.value.filter(
														(type) =>
															type !== "poster"
													)
												);
											}
										}}
										className="h-5 w-5 sm:h-4 sm:w-4"
									/>
								</FormControl>
								<FormLabel className="text-md font-normal">
									Poster
								</FormLabel>
							</FormItem>
						)}
						{movie.Backdrop && (
							<FormItem className="flex flex-row items-start space-x-2 space-y-0">
								<FormControl>
									<Checkbox
										checked={field.value.includes(
											"backdrop"
										)}
										onCheckedChange={(checked) => {
											if (checked) {
												field.onChange([
													...field.value,
													"backdrop",
												]);
											} else {
												field.onChange(
													field.value.filter(
														(type) =>
															type !== "backdrop"
													)
												);
											}
										}}
										className="h-5 w-5 sm:h-4 sm:w-4"
									/>
								</FormControl>
								<FormLabel className="text-md font-normal">
									Backdrop
								</FormLabel>
							</FormItem>
						)}
					</div>
				</div>
			)}
		/>
	);

	const mainMovies = moviesDisplay.filter((movie) => movie.MainItem);
	const otherMovies = moviesDisplay.filter((movie) => !movie.MainItem);

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
						<DialogTitle>{posterSet.Title}</DialogTitle>
						<DialogDescription>
							Set Author: {posterSet.User.Name}
						</DialogDescription>
						<DialogDescription>
							<Link
								href={`https://mediux.pro/sets/${posterSet.ID}`}
								className="hover:text-primary transition-colors text-sm text-muted-foreground"
								target="_blank"
								rel="noopener noreferrer"
							>
								Set ID: {posterSet.ID}
							</Link>
						</DialogDescription>
					</DialogHeader>

					<Form {...form}>
						<form
							onSubmit={form.handleSubmit(onSubmit)}
							className="space-y-2"
						>
							{moviesDisplay.length > 1 && (
								<FormLabel className="text-md font-normal">
									Movies in Set & Server:
								</FormLabel>
							)}

							{mainMovies.map(renderMovieField)}
							{otherMovies.map(renderMovieField)}
							<FormMessage />

							{/* Total Size of Selected Types */}
							{totalSelectedSize && (
								<div className="text-sm text-muted-foreground">
									{totalSelectedSizeLabel} {totalSelectedSize}
								</div>
							)}

							{/* Progress Bar */}

							{progressValue > 0 && (
								<div className="w-full">
									<div className="flex items-center justify-between">
										<Progress
											value={progressValue}
											className={`flex-1 ${
												progressColor
													? `[&>div]:bg-${progressColor}-500`
													: ""
											}`}
										/>
										<span className="ml-2 text-sm text-muted-foreground">
											{Math.round(progressValue)}%
										</span>
									</div>
									{Object.entries(progressDownloads).map(
										([movieName, statuses]) => (
											<div
												key={movieName}
												className="my-2"
											>
												<div className="font-bold text-sm">
													{movieName}
												</div>
												{statuses.poster && (
													<div className="flex justify-between text-sm text-muted-foreground">
														{statuses.poster ===
														"success" ? (
															<div className="flex items-center">
																<Check className="mr-1 h-4 w-4" />
																<span>
																	Poster
																</span>
															</div>
														) : statuses.poster ===
														  "failed" ? (
															<div className="flex items-center text-destructive">
																<X className="mr-1 h-4 w-4" />
																<span>
																	Poster
																</span>
															</div>
														) : (
															<div className="flex items-center">
																<LoaderIcon className="mr-1 h-4 w-4 animate-spin" />
																<span>
																	Downloading
																	Poster
																</span>
															</div>
														)}
													</div>
												)}
												{statuses.backdrop && (
													<div className="flex justify-between text-sm text-muted-foreground">
														{statuses.backdrop ===
														"success" ? (
															<div className="flex items-center">
																<Check className="mr-1 h-4 w-4" />
																<span>
																	Backdrop
																</span>
															</div>
														) : statuses.backdrop ===
														  "failed" ? (
															<div className="flex items-center text-destructive">
																<X className="mr-1 h-4 w-4" />
																<span>
																	Backdrop
																</span>
															</div>
														) : (
															<div className="flex items-center">
																<LoaderIcon className="mr-1 h-4 w-4 animate-spin" />
																<span>
																	Downloading
																	Backdrop
																</span>
															</div>
														)}
													</div>
												)}
												{statuses.addToDB && (
													<div className="flex justify-between text-sm text-muted-foreground">
														{statuses.addToDB ===
														"success" ? (
															<div className="flex items-center">
																<Check className="mr-1 h-4 w-4" />
																<span>
																	Added to DB
																</span>
															</div>
														) : statuses.addToDB ===
														  "failed" ? (
															<div className="flex items-center text-destructive">
																<X className="mr-1 h-4 w-4" />
																<span>
																	Failing
																	adding to DB
																</span>
															</div>
														) : (
															<div className="flex items-center">
																<LoaderIcon className="mr-1 h-4 w-4 animate-spin" />
																<span>
																	Adding to DB
																</span>
															</div>
														)}
													</div>
												)}
											</div>
										)
									)}
								</div>
							)}

							{/* Warning Messages */}
							{progressWarningMessages.length > 0 && (
								<div className="my-2">
									<Accordion type="single" collapsible>
										<AccordionItem value="warnings">
											<AccordionTrigger className="text-destructive">
												Failed Downloads (
												{progressWarningMessages.length}
												)
											</AccordionTrigger>
											<AccordionContent>
												<div className="flex flex-col space-y-2">
													{progressWarningMessages.map(
														(message) => (
															<div
																key={message}
																className="flex items-center text-destructive"
															>
																<X className="mr-1 h-4 w-4" />
																<span>
																	{message}
																</span>
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
										onClick={() => {
											onSubmit(form.getValues());
										}}
										disabled={Object.values(
											watchSelectionsByMovie
										).every(
											(arr) => !arr || arr.length === 0
										)}
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

export default DownloadModalMovie;
