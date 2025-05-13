"use client";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "@/components/ui/accordion";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import ErrorMessage from "@/components/ui/error-message";
import Loader from "@/components/ui/loader";
import { MediaCarousel } from "@/components/ui/media-carousel";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { H1, Lead } from "@/components/ui/typography";
import { log } from "@/lib/logger";
import { cn } from "@/lib/utils";
import { fetchMediaServerItemContent } from "@/services/api.mediaserver";
import { fetchMediuxSets } from "@/services/api.mediux";
import { Guid, MediaItem } from "@/types/mediaItem";
import { PosterSets } from "@/types/posterSets";
import Image from "next/image";
import { useRouter } from "next/navigation";
import React, { useEffect, useState } from "react";
import { usePosterMediaStore } from "@/lib/setStore";

const MediaItemPage = () => {
	const router = useRouter();

	const [isMounted, setIsMounted] = useState(false);

	const partialMediaItem = usePosterMediaStore((state) => state.mediaItem); // Retrieve partial mediaItem from Zustand

	const [isBlurred, setIsBlurred] = useState(false);

	const [mediaItem, setMediaItem] = React.useState<MediaItem | null>(
		partialMediaItem
	);
	const [imdbLink, setImdbLink] = React.useState<string>("");
	const [tmdbLink, setTmdbLink] = React.useState<string>("");
	const [tvdbLink, setTvdbLink] = React.useState<string>("");

	const [posterSets, setPosterSets] = useState<PosterSets | null>(null);

	// State to track the selected sorting option
	const [sortOption, setSortOption] = useState<string>("");

	// Loading State
	const [isLoading, setIsLoading] = React.useState(true);

	// Error Handling
	const [hasError, setHasError] = React.useState(false);
	const [errorMessage, setErrorMessage] = React.useState("");

	// Handle scroll event to blur the background
	useEffect(() => {
		const handleScroll = () => {
			// Check if the user has scrolled down 300px (adjust as needed)
			if (window.scrollY > 300) {
				setIsBlurred(true);
			} else {
				setIsBlurred(false);
			}
		};

		// Add scroll event listener
		window.addEventListener("scroll", handleScroll);

		// Cleanup event listener on component unmount
		return () => {
			window.removeEventListener("scroll", handleScroll);
		};
	}, []);

	useEffect(() => {
		if (isMounted) return;
		setIsMounted(true);

		const fetchIMDBLink = (guids: Guid[]) => {
			if (!guids || guids.length === 0) {
				log("Media Item Page - No GUIDs found");
				return;
			}
			const imdbGuid = guids.find((guid) => guid.Provider === "imdb");
			const tmdbGuid = guids.find((guid) => guid.Provider === "tmdb");
			const tvdbGuid = guids.find((guid) => guid.Provider === "tvdb");
			if (imdbGuid) {
				setImdbLink(imdbGuid.ID);
			}
			if (tmdbGuid) {
				setTmdbLink(tmdbGuid.ID);
			}
			if (tvdbGuid) {
				setTvdbLink(tvdbGuid.ID);
			}
		};

		const fetchPosterSets = async (responseItem: MediaItem) => {
			// Check if the item has GUIDs
			try {
				if (!responseItem.Guids || responseItem.Guids.length === 0) {
					return;
				}
				const tmdbID = responseItem.Guids.find(
					(guid) => guid.Provider === "tmdb"
				)?.ID;
				if (tmdbID) {
					const resp = await fetchMediuxSets(
						tmdbID,
						responseItem.Type
					);
					if (!resp) {
						throw new Error("No response from Mediux API");
					} else if (resp.status !== "success") {
						throw new Error(resp.message);
					}
					const sets = resp.data;
					// If no sets are returned, assign an empty array.
					setPosterSets(sets ? sets : { Sets: [] });
				} else {
					log(
						"Media Item Page - No TMDB ID found in GUIDs, searching by external IDs (not implemented yet)"
					);
					// TODO: ADD THIS
				}
			} catch (error) {
				log("Media Item Page - Error fetching poster sets:", error);
				setHasError(true);
				if (error instanceof Error) {
					setErrorMessage(error.message);
				}
				// Fallback to empty sets
				setPosterSets({ Sets: [] });
			}
		};

		const fetchAllInfo = async () => {
			try {
				// Use local state, fallback to Zustand if needed.
				let currentMediaItem = mediaItem;
				if (!currentMediaItem) {
					const storedMediaItem =
						usePosterMediaStore.getState().mediaItem;
					if (storedMediaItem) {
						currentMediaItem = storedMediaItem;
						setMediaItem(storedMediaItem);
					} else {
						throw new Error("No media item found");
					}
				}
				// Now safely use currentMediaItem
				const resp = await fetchMediaServerItemContent(
					currentMediaItem.RatingKey,
					currentMediaItem.LibraryTitle
				);
				if (!resp) {
					throw new Error("No response from Plex API");
				}
				if (resp.status !== "success") {
					throw new Error(resp.message);
				}
				const responseItem = resp.data;
				if (!responseItem) {
					throw new Error("No item found in response");
				}
				setMediaItem(responseItem);
				fetchIMDBLink(responseItem.Guids);
				fetchPosterSets(responseItem);
			} catch (error) {
				setHasError(true);
				if (error instanceof Error) {
					setErrorMessage(error.message);
				} else {
					setErrorMessage("An unknown error occurred, check logs.");
				}
			} finally {
				setIsLoading(false);
				setIsMounted(false);
			}
		};

		fetchAllInfo();
	}, [partialMediaItem]);

	if (isLoading) {
		return <Loader message="Loading media item..." />;
	}

	if (!mediaItem) {
		return (
			<div className="flex flex-col items-center">
				<ErrorMessage message="No media item found." />
				<Button
					className="mt-4"
					variant="secondary"
					onClick={() => {
						router.push("/");
					}}
				>
					Go to Home
				</Button>
			</div>
		);
	}

	if (posterSets?.Sets?.length === 0) {
		return <ErrorMessage message="No poster sets found." />;
	}

	if (hasError) {
		document.title = "Poster-Setter - Error";
		return (
			<div className="flex flex-col items-center">
				<ErrorMessage message={errorMessage} />
				<Button
					className="mt-4"
					variant="secondary"
					onClick={() => {
						router.push("/");
					}}
				>
					Go to Home
				</Button>
				<Button
					className="mt-4"
					onClick={() => {
						router.push("/logs");
					}}
				>
					Go to Logs
				</Button>
			</div>
		);
	} else {
		document.title = `Poster-Setter - ${mediaItem?.Title}`;
		log("Media Item Page - Fetched media item:", mediaItem);
		log("Media Item Page - Fetched poster sets:", posterSets);
		log("Media Item Page - Fetched links:", {
			imdbLink,
			tmdbLink,
			tvdbLink,
		});
	}

	return (
		<>
			<div
				className={cn(
					"fixed inset-0 -z-20 overflow-hidden w-full h-full transition-all duration-1000",
					isBlurred && "blur-md"
				)}
			>
				<div className="absolute inset-0 bg-background">
					<div className="absolute inset-0 opacity-[0.015] mix-blend-overlay">
						<div className="absolute inset-0 bg-[url(/gradient.svg)]"></div>{" "}
					</div>

					<div
						className="absolute inset-0"
						style={
							{
								background: `
            radial-gradient(ellipse at 30% 30%, var(--dynamic-left) 0%, transparent 60%),
            radial-gradient(ellipse at bottom right, var(--dynamic-bottom) 0%, transparent 60%),
            radial-gradient(ellipse at center, var(--dynamic-dark-muted) 0%, transparent 80%),
            var(--background)
          `,
								opacity: 0.5,
							} as React.CSSProperties
						}
					/>
				</div>

				<div className="absolute top-0 right-0 w-full lg:w-[70vw] aspect-[16/9] z-50">
					<div className="relative w-full h-full">
						<Image
							src={`/api/mediaserver/image/${mediaItem.RatingKey}/backdrop`}
							alt={"Backdrop"}
							fill
							priority
							unoptimized
							className="object-cover object-right-top"
							style={{
								maskImage: `url(/gradient.svg)`,
								WebkitMaskImage: `url(/gradient.svg)`,
								maskSize: "100% 100%",
								WebkitMaskSize: "100% 100%",
								maskRepeat: "no-repeat",
								WebkitMaskRepeat: "no-repeat",
								maskPosition: "center",
								WebkitMaskPosition: "center",
							}}
						/>
					</div>
				</div>
			</div>

			<div className="p-4 lg:p-6">
				<div className="pb-6">
					{/* Title and Summary */}
					<div className="flex flex-col pt-40 justify-end items-center text-center lg:items-start lg:text-left">
						<H1 className="mb-1">{mediaItem?.Title}</H1>
						{/* Hide summary on mobile */}
						<Lead className="text-primary-dynamic max-w-4xl hidden md:block">
							{mediaItem?.Summary}
						</Lead>
					</div>

					{/* Year, Content Rating, IMDb Rating, and TV Show Information */}
					<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-4">
						{/* Year */}
						{mediaItem?.Year && (
							<Badge className="flex items-center text-sm">
								{mediaItem?.Year}
							</Badge>
						)}

						{/* Content Rating */}
						{mediaItem?.ContentRating && (
							<Badge className="flex items-center text-sm">
								{mediaItem?.ContentRating}
							</Badge>
						)}

						{/* IMDb Rating */}
						{imdbLink && (
							<div
								className="flex items-center text-md cursor-pointer"
								onClick={() => {
									window.open(
										`https://www.imdb.com/title/${imdbLink}/`,
										"_blank"
									);
								}}
							>
								<div className="relative w-[30px] h-[30px] mr-2">
									<Image
										src={"/imdb_logo.png"}
										alt="Rating"
										fill
										sizes="30px"
										className="object-contain"
									/>
								</div>
								<Lead className="text-primary-dynamic">
									{mediaItem?.AudienceRating}
								</Lead>
							</div>
						)}

						{/* Show Information for TV Shows */}
						{mediaItem?.Type === "show" &&
							mediaItem.Series?.SeasonCount &&
							mediaItem.Series.EpisodeCount > 0 && (
								<Lead className="flex items-center text-md text-primary-dynamic">
									{mediaItem.Series?.SeasonCount}{" "}
									{mediaItem.Series?.SeasonCount > 1
										? "Seasons"
										: "Season"}{" "}
									| {mediaItem.Series?.EpisodeCount}{" "}
									{mediaItem.Series?.EpisodeCount > 1
										? "Episodes"
										: "Episode"}
								</Lead>
							)}
					</div>

					<div className="lg:flex items-center text-white gap-8 tracking-wide mt-4">
						{mediaItem?.Type === "movie" && (
							<div className="flex flex-col w-full items-center">
								<Accordion
									type="single"
									collapsible
									className="w-full"
								>
									<AccordionItem value="movie-details">
										<AccordionTrigger className="text-primary font-semibold">
											Movie Details
										</AccordionTrigger>
										<AccordionContent className="mt-2 text-sm text-muted-foreground space-y-2">
											{/* Show the Movie File Path */}
											{mediaItem?.Movie?.File?.Path && (
												<p>
													<span className="font-semibold">
														File Path:
													</span>{" "}
													{
														mediaItem?.Movie?.File
															?.Path
													}
												</p>
											)}

											{/* Show the Movie File Size */}
											{mediaItem?.Movie?.File?.Size && (
												<p>
													<span className="font-semibold">
														File Size:
													</span>{" "}
													{mediaItem.Movie.File
														.Size >=
													1024 * 1024 * 1024
														? `${(
																mediaItem.Movie
																	.File.Size /
																(1024 *
																	1024 *
																	1024)
														  ).toFixed(2)} GB`
														: `${(
																mediaItem.Movie
																	.File.Size /
																(1024 * 1024)
														  ).toFixed(2)} MB`}
												</p>
											)}

											{/* Show the Movie Duration */}
											{mediaItem?.Movie?.File
												?.Duration && (
												<p>
													<span className="font-semibold">
														Duration:
													</span>{" "}
													{mediaItem.Movie.File
														.Duration < 3600000
														? `${Math.floor(
																mediaItem.Movie
																	.File
																	.Duration /
																	60000
														  )} minutes`
														: `${Math.floor(
																mediaItem.Movie
																	.File
																	.Duration /
																	3600000
														  )}hr ${Math.floor(
																(mediaItem.Movie
																	.File
																	.Duration %
																	3600000) /
																	60000
														  )}min`}
												</p>
											)}
										</AccordionContent>
									</AccordionItem>
								</Accordion>
							</div>
						)}
					</div>

					{posterSets?.Sets && posterSets.Sets.length > 0 && (
						<>
							<div className="flex justify-end mb-6 pr-4">
								<Select
									onValueChange={(value) => {
										setSortOption(value);
									}}
									defaultValue=""
								>
									<SelectTrigger className="w-[180px]">
										<SelectValue placeholder="Sort By" />
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="date">
											Date
										</SelectItem>
										<SelectItem value="name">
											User Name
										</SelectItem>
									</SelectContent>
								</Select>
							</div>
							<div className="divide-y divide-primary-dynamic/20 space-y-6">
								{mediaItem &&
									posterSets.Sets &&
									posterSets.Sets.length > 0 &&
									[...posterSets.Sets]
										.sort((a, b) => {
											// Inline sorting logic based on the selected sort option
											if (sortOption === "date") {
												return (
													new Date(
														b.DateUpdated
													).getTime() -
													new Date(
														a.DateUpdated
													).getTime()
												);
											} else if (sortOption === "name") {
												return a.User.Name.localeCompare(
													b.User.Name
												);
											}
											return 0; // Default case (no sorting)
										})
										.map((set) => (
											<div key={set.ID} className="pb-6">
												<MediaCarousel
													set={set}
													mediaItem={mediaItem}
												/>
											</div>
										))}
							</div>
						</>
					)}
				</div>
			</div>
		</>
	);
};

export default MediaItemPage;
