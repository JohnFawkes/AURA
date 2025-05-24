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
import {
	fetchMediuxSets,
	fetchMediuxUserFollowHides,
} from "@/services/api.mediux";
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
import { Checkbox } from "@/components/ui/checkbox";

const MediaItemPage = () => {
	const router = useRouter();

	const hasFetchedInfo = useRef(false);

	const partialMediaItem = useMediaStore((state) => state.mediaItem); // Retrieve partial mediaItem from Zustand

	const [mediaItem, setMediaItem] = React.useState<MediaItem | null>(
		partialMediaItem
	);

	const [posterSets, setPosterSets] = useState<PosterSet[] | null>(null);
	const [filteredPosterSets, setFilteredPosterSets] = useState<
		PosterSet[] | null
	>(null);

	// State to track the selected sorting option
	const [sortOption, setSortOption] = useState<string>("");
	const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");

	// Loading State
	const [isLoading, setIsLoading] = useState(true);

	// Error Handling
	const [hasError, setHasError] = useState(false);
	const [errorMessage, setErrorMessage] = useState("");

	const [userFollows, setUserFollows] = useState<
		{ ID: string; Username: string }[]
	>([]);
	const [userHides, setUserHides] = useState<
		{ ID: string; Username: string }[]
	>([]);
	const [showHiddenUsers, setShowHiddenUsers] = useState(false);

	const handleShowHiddenUsers = () => {
		setShowHiddenUsers((prev) => {
			// If the checkbox is checked, set it to false
			if (prev) {
				return false;
			}
			// If the checkbox is unchecked, set it to true
			return true;
		});
	};

	useEffect(() => {
		if (hasFetchedInfo.current) return;
		hasFetchedInfo.current = true;

		const fetchUserFollowHides = async () => {
			try {
				const resp = await fetchMediuxUserFollowHides();
				if (!resp) {
					throw new Error("No response from Mediux API");
				}
				const follows = resp.data?.Follows || [];
				const hides = resp.data?.Hides || [];
				log(
					"Media Item Page - Fetched user follows/hides:",
					"Follows:",
					follows,
					"Hides:",
					hides
				);
				setUserFollows(follows);
				setUserHides(hides);
			} catch (error) {
				log(
					"Media Item Page - Error fetching user follows/hides:",
					error
				);
				setHasError(true);
				if (error instanceof Error) {
					setErrorMessage(error.message);
				}
				// Fallback to empty follows/hides
				setUserFollows([]);
				setUserHides([]);
			} finally {
				setIsLoading(false);
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
					if (!sets) {
						throw new Error(
							`No Poster Sets found for ${responseItem.Title}`
						);
					}
					log(`Poster Sets for ${responseItem.Title}:`, sets);
					setPosterSets(sets);
					fetchUserFollowHides();
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
				log(
					`Media Item Page - Fetched item: ${responseItem.Title}`,
					responseItem
				);
				setMediaItem(responseItem);
				fetchPosterSets(responseItem);
			} catch (error) {
				setHasError(true);
				if (error instanceof Error) {
					setErrorMessage(error.message);
				} else {
					setErrorMessage("An unknown error occurred, check logs.");
				}
			}
		};

		fetchAllInfo();
	}, [partialMediaItem, mediaItem]);

	useEffect(() => {
		if (posterSets) {
			console.log("Poster Sets:", posterSets);

			// Filter out hidden users
			const filtered = posterSets.filter((set) => {
				if (showHiddenUsers) {
					return true; // Show all if the checkbox is checked
				}
				// Check if the user is in the hides list
				const isHidden = userHides.some(
					(hide) => hide.Username === set.User.Name
				);
				return !isHidden; // Show only if not hidden
			});

			// Sort the filtered poster sets
			// Follows should always be at the top
			filtered.sort((a, b) => {
				const isAFollow = userFollows.some(
					(follow) => follow.Username === a.User.Name
				);
				const isBFollow = userFollows.some(
					(follow) => follow.Username === b.User.Name
				);
				if (isAFollow && !isBFollow) return -1;
				if (!isAFollow && isBFollow) return 1;

				if (sortOption === "name") {
					return a.User.Name.localeCompare(b.User.Name);
				}

				// Default and "date" sort: newest to oldest unless sortOrder is "asc"
				const dateA = new Date(a.DateUpdated);
				const dateB = new Date(b.DateUpdated);
				if (sortOption === "date") {
					return sortOrder === "asc"
						? dateA.getTime() - dateB.getTime() // oldest to newest
						: dateB.getTime() - dateA.getTime(); // newest to oldest
				}
				// Default: newest to oldest
				return dateB.getTime() - dateA.getTime();
			});

			console.log("Filter & Sorted Poster Sets:", filtered);
			setFilteredPosterSets(filtered);
		}
	}, [
		posterSets,
		showHiddenUsers,
		userHides,
		userFollows,
		sortOption,
		sortOrder,
	]);

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

	if (hasError) {
		if (typeof window !== "undefined") {
			// Safe to use document here.
			document.title = "Aura | Error";
		}
	} else {
		if (typeof window !== "undefined") {
			// Safe to use document here.
			document.title = `AURA | ${mediaItem?.Title}`;
		}
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

					{isLoading && (
						<div className="flex justify-center">
							<Loader message="Loading Poster Sets..." />
						</div>
					)}
					{hasError && (
						<div className="flex justify-center">
							<ErrorMessage message={errorMessage} />
						</div>
					)}

					{posterSets && posterSets.length > 0 && (
						<>
							<div className="flex justify-end mb-6 pr-4 items-center gap-4">
								{userHides.length > 0 && (
									<label className="inline-flex items-center">
										<Checkbox
											checked={showHiddenUsers}
											onCheckedChange={
												handleShowHiddenUsers
											}
											className="mr-2 h-5 w-5 sm:h-4 sm:w-4" // Larger on mobile, normal on sm+
										/>
										<span className="text-sm">
											Show hidden users
										</span>
									</label>
								)}
								{sortOption !== "" && (
									<Button
										variant="ghost"
										onClick={() =>
											setSortOrder(
												sortOrder === "asc"
													? "desc"
													: "asc"
											)
										}
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
									onValueChange={(value) =>
										setSortOption(value)
									}
									defaultValue="date"
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
								{(filteredPosterSets ?? []).map((set) => (
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
