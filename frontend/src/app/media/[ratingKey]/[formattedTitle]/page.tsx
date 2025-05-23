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
import React, { useEffect, useRef, useState } from "react";
import {
	ArrowDownAZ,
	ArrowUpAZ,
	CalendarArrowDown,
	CalendarArrowUp,
} from "lucide-react";
import { useMediaStore } from "@/lib/mediaStore";

interface ProviderInfo {
	id: string;
	rating: string;
	logoUrl: string;
	linkUrl: string;
}

const providerLogoMap: {
	[key: string]: { logoUrl: string; urlPrefix: string };
} = {
	imdb: {
		logoUrl: "/imdb-icon.png",
		urlPrefix: "https://www.imdb.com/title/",
	},
	tmdb: {
		logoUrl: "/tmdb-icon.svg",
		urlPrefix: "https://www.themoviedb.org/movie/",
	},
	tvdb: {
		logoUrl: "/tvdb-icon.svg",
		urlPrefix: "https://thetvdb.com/dereferrer/",
	},
	rottentomatoes: {
		logoUrl: "/rottentomatoes-icon.png",
		urlPrefix: "https://www.rottentomatoes.com/",
	},
	community: {
		logoUrl: "",
		urlPrefix: "",
	},
};

const MediaItemPage = () => {
	const router = useRouter();

	const hasFetchedInfo = useRef(false);

	const partialMediaItem = useMediaStore((state) => state.mediaItem); // Retrieve partial mediaItem from Zustand

	const [isBlurred, setIsBlurred] = useState(false);

	const [mediaItem, setMediaItem] = React.useState<MediaItem | null>(
		partialMediaItem
	);

	const [posterSets, setPosterSets] = useState<PosterSets | null>(null);

	// State to track the selected sorting option
	const [sortOption, setSortOption] = useState<string>("");
	const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");

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
		if (hasFetchedInfo.current) return;
		hasFetchedInfo.current = true;

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
					const storedMediaItem = useMediaStore.getState().mediaItem;
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
			}
		};

		fetchAllInfo();
	}, [partialMediaItem, mediaItem]);

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
		document.title = "AURA | Error";
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
		document.title = `AURA | ${mediaItem?.Title}`;
		log("Media Item Page - Fetched media item:", mediaItem);
		log("Media Item Page - Fetched poster sets:", posterSets);
	}

	const guidMap: { [provider: string]: ProviderInfo } = {};
	mediaItem.Guids.forEach((guid: Guid) => {
		if (guid.Provider) {
			const providerInfo = providerLogoMap[guid.Provider];
			if (providerInfo) {
				guidMap[guid.Provider] = {
					id: guid.ID || "",
					rating:
						mediaItem.Guids.find(
							(g) => g.Provider === guid.Provider
						)?.Rating || "",
					logoUrl: providerInfo.logoUrl,
					linkUrl:
						guid.Provider === "tvdb"
							? `https://www.thetvdb.com/dereferrer/${
									mediaItem.Type === "show"
										? "series"
										: "movies"
							  }/${guid.ID}`
							: guid.Provider === "tmdb"
							? mediaItem.Type === "show"
								? `https://www.themoviedb.org/tv/${guid.ID}`
								: `https://www.themoviedb.org/movie/${guid.ID}`
							: `${providerInfo.urlPrefix}${guid.ID}`,
				};
			}
		}
	});

	console.log("Media Item Page - GUID Map:", guidMap);

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
							unoptimized
							loading="lazy"
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

						{/* Ratings and External Links */}
						<div className="flex gap-4">
							{Object.entries(guidMap).map(([provider, info]) => (
								<div
									key={provider}
									className="flex items-center gap-2"
								>
									{provider === "community" ? (
										<>
											{/* Display a star icon with the rating */}
											<span className="text-sm flex items-center gap-1">
												<svg
													xmlns="http://www.w3.org/2000/svg"
													width="16"
													height="16"
													fill="currentColor"
													viewBox="0 0 16 16"
												>
													<path d="M3.612 15.443c-.396.198-.86-.106-.746-.592l.83-4.73L.173 6.765c-.329-.32-.158-.888.283-.95l4.898-.696 2.189-4.327c.197-.39.73-.39.927 0l2.189 4.327 4.898.696c.441.062.612.63.282.95l-3.522 3.356.83 4.73c.114.486-.35.79-.746.592L8 13.187l-4.389 2.256z" />
												</svg>
												{info.rating}
											</span>
										</>
									) : (
										<>
											<a
												href={info.linkUrl!}
												target="_blank"
												rel="noopener noreferrer"
											>
												<div className="relative ml-1 w-[40px] h-[40px]">
													<Image
														src={info.logoUrl}
														alt={`${provider} Logo`}
														fill
														className="object-contain"
													/>
												</div>
											</a>
											{/* Only display rating if it exists */}
											{info.rating && (
												<span className="text-sm">
													{info.rating}
												</span>
											)}
										</>
									)}
								</div>
							))}
						</div>

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
								{sortOption !== "" && (
									<Button
										variant="ghost"
										onClick={() => {
											setSortOrder(
												sortOrder === "asc"
													? "desc"
													: "asc"
											);
										}}
									>
										{sortOption === "name" &&
											sortOrder === "desc" && (
												<ArrowUpAZ />
											)}
										{sortOption === "name" &&
											sortOrder === "asc" && (
												<ArrowDownAZ />
											)}
										{sortOption === "date" &&
											sortOrder === "desc" && (
												<CalendarArrowDown />
											)}
										{sortOption === "date" &&
											sortOrder === "asc" && (
												<CalendarArrowUp />
											)}
									</Button>
								)}
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
											// Sorting logic based on the sort option and sort order
											if (sortOption === "date") {
												return sortOrder === "asc"
													? new Date(
															a.DateUpdated
													  ).getTime() -
															new Date(
																b.DateUpdated
															).getTime()
													: new Date(
															b.DateUpdated
													  ).getTime() -
															new Date(
																a.DateUpdated
															).getTime();
											} else if (sortOption === "name") {
												return sortOrder === "asc"
													? a.User.Name.localeCompare(
															b.User.Name
													  )
													: b.User.Name.localeCompare(
															a.User.Name
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
