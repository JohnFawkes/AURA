"use client";

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
import { log } from "@/lib/logger";
import { fetchMediaServerItemContent } from "@/services/api.mediaserver";
import { fetchMediuxSets } from "@/services/api.mediux";
import { MediaItem } from "@/types/mediaItem";
import { PosterSet } from "@/types/posterSets";
import { useRouter } from "next/navigation";
import React, { useEffect, useRef, useState } from "react";
import {
	ArrowDownAZ,
	ArrowUpAZ,
	CalendarArrowDown,
	CalendarArrowUp,
} from "lucide-react";
import { useMediaStore } from "@/lib/mediaStore";
import { DimmedBackground } from "@/components/dimmed_backdrop";
import { MediaItemDetails } from "@/components/media_item_details";

const MediaItemPage = () => {
	const router = useRouter();

	const hasFetchedInfo = useRef(false);

	const partialMediaItem = useMediaStore((state) => state.mediaItem); // Retrieve partial mediaItem from Zustand

	const [mediaItem, setMediaItem] = React.useState<MediaItem | null>(
		partialMediaItem
	);

	const [posterSets, setPosterSets] = useState<PosterSet[] | null>(null);

	// State to track the selected sorting option
	const [sortOption, setSortOption] = useState<string>("");
	const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");

	// Loading State
	const [isLoading, setIsLoading] = React.useState(true);

	// Error Handling
	const [hasError, setHasError] = React.useState(false);
	const [errorMessage, setErrorMessage] = React.useState("");

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
						responseItem.Type,
						responseItem.LibraryTitle,
						responseItem.RatingKey
					);
					if (!resp) {
						throw new Error("No response from Mediux API");
					} else if (resp.status !== "success") {
						throw new Error(resp.message);
					}
					const sets = resp.data;
					// If no sets are returned, assign an empty array.
					setPosterSets(sets ? sets : []);
				}
			} catch (error) {
				log("Media Item Page - Error fetching poster sets:", error);
				setHasError(true);
				if (error instanceof Error) {
					setErrorMessage(error.message);
				}
				// Fallback to empty sets
				setPosterSets([]);
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

	if (posterSets?.length === 0) {
		return <ErrorMessage message="No poster sets found." />;
	}

	if (hasError) {
		if (typeof window !== "undefined") {
			// Safe to use document here.
			document.title = "Aura | Error";
		}
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
		if (typeof window !== "undefined") {
			// Safe to use document here.
			document.title = `AURA | ${mediaItem?.Title}`;
		}

		log("Media Item Page - Fetched media item:", mediaItem);
		log("Media Item Page - Fetched poster sets:", posterSets);
	}

	return (
		<>
			{/* Backdrop Background */}
			<DimmedBackground
				backdropURL={`/api/mediaserver/image/${mediaItem.RatingKey}/backdrop`}
			/>

			{/* Header */}
			<div className="p-4 lg:p-6">
				<div className="pb-6">
					<MediaItemDetails
						ratingKey={mediaItem.RatingKey}
						mediaItemType={mediaItem.Type}
						title={mediaItem.Title}
						summary={mediaItem.Summary || ""}
						year={mediaItem.Year}
						contentRating={mediaItem.ContentRating || ""}
						seasonCount={
							mediaItem.Type === "show"
								? mediaItem.Series?.SeasonCount || 0
								: 0
						}
						episodeCount={
							mediaItem.Type === "show"
								? mediaItem.Series?.EpisodeCount || 0
								: 0
						}
						moviePath={mediaItem.Movie?.File?.Path || "N/A"}
						movieSize={mediaItem.Movie?.File?.Size || 0}
						movieDuration={mediaItem.Movie?.File?.Duration || 0}
						guids={mediaItem.Guids || []}
						existsInDB={mediaItem.ExistInDatabase || false}
					/>

					{posterSets && posterSets.length > 0 && (
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
									posterSets &&
									posterSets.length > 0 &&
									[...posterSets]
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
											return (
												new Date(
													b.DateUpdated
												).getTime() -
												new Date(
													a.DateUpdated
												).getTime()
											);
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
