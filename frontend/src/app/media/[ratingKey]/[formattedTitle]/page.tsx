"use client";

import { formatMediaItemUrl } from "@/helper/format-media-item-url";
import { getAdjacentMediaItem } from "@/helper/search-idb-for-tmdb-id";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { fetchMediaServerItemContent } from "@/services/mediaserver/api-mediaserver-fetch-item-content";
import { fetchMediuxSets } from "@/services/mediux/api-mediux-fetch-sets";
import { fetchMediuxUserFollowHides } from "@/services/mediux/api-mediux-fetch-user-follow-hide";
import {
	ArrowDown01,
	ArrowDown10,
	ArrowDownAZ,
	ArrowDownZA,
	ArrowLeftCircle,
	ArrowRightCircle,
	CalendarArrowDown,
	CalendarArrowUp,
} from "lucide-react";

import { useEffect, useMemo, useRef, useState } from "react";

import { useRouter } from "next/navigation";

import { DimmedBackground } from "@/components/shared/dimmed_backdrop";
import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { MediaCarousel } from "@/components/shared/media-carousel";
import { MediaItemDetails } from "@/components/shared/media_item_details";
import { SortControl } from "@/components/shared/select_sort";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";

import { cn } from "@/lib/cn";
import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";
import { useMediaStore } from "@/lib/stores/global-store-media-store";
import { useMediaPageStore } from "@/lib/stores/page-store-media";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterSet } from "@/types/media-and-posters/poster-sets";

const MediaItemPage = () => {
	const router = useRouter();
	const isMounted = useRef(false);

	// Partial Media Item States
	const mediaStore = useMediaStore();
	const partialMediaItem = mediaStore.mediaItem;

	// Library Sections States (From Library Section Store)
	const librarySectionsMap = useLibrarySectionsStore((state) => state.sections);
	const librarySectionsHasHydrated = useLibrarySectionsStore((state) => state.hasHydrated);

	// Main Media Item States
	const [mediaItem, setMediaItem] = useState<MediaItem | null>(null);
	const [mediaItemLoading, setMediaItemLoading] = useState<boolean>(true);
	const [existsInOtherSections, setExistsInOtherSections] = useState<MediaItem | null>(null);
	const [existsInDB, setExistsInDB] = useState<boolean>(mediaItem?.ExistInDatabase || false);

	// User Follows/Hides States
	const [userFollows, setUserFollows] = useState<{ ID: string; Username: string }[]>([]);
	const [userHides, setUserHides] = useState<{ ID: string; Username: string }[]>([]);
	const [userFollowsHidesLoading, setUserFollowsHidesLoading] = useState<boolean>(true);

	// Poster Sets States
	const [posterSets, setPosterSets] = useState<PosterSet[] | null>(null);
	const [filteredPosterSets, setFilteredPosterSets] = useState<PosterSet[] | null>(null);
	const [posterSetsLoading, setPosterSetsLoading] = useState<boolean>(true);

	// UI States from Media Page Store
	const {
		sortOption,
		setSortOption,
		sortOrder,
		setSortOrder,
		showOnlyTitlecardSets,
		setShowOnlyTitlecardSets,
		showHiddenUsers,
		setShowHiddenUsers,
	} = useMediaPageStore();

	// Loading States
	const [loadingMessage, setLoadingMessage] = useState("Loading...");
	const isLoading = useMemo(() => {
		return mediaItemLoading || userFollowsHidesLoading || posterSetsLoading;
	}, [mediaItemLoading, userFollowsHidesLoading, posterSetsLoading]);

	// Error States
	const [hasError, setHasError] = useState(false);
	const [error, setError] = useState<APIResponse<unknown> | null>(null);

	// Image Version State (for forcing image reloads)
	const [imageVersion, setImageVersion] = useState(Date.now());

	// Adjacent Items States
	const [adjacentItems, setAdjacentItems] = useState<{
		previous: MediaItem | null;
		next: MediaItem | null;
	}>({
		previous: null,
		next: null,
	});

	useEffect(() => {
		const validSortOptions =
			mediaItem?.Type === "show"
				? ["user", "date", "numberOfSeasons", "numberOfTitlecards"]
				: mediaItem?.Type === "movie"
					? ["user", "date", "numberOfItemsInCollection"]
					: ["user", "date"];

		if (!validSortOptions.includes(sortOption)) {
			setSortOption("date");
			setSortOrder("desc");
		}
	}, [sortOption, setSortOption, setSortOrder, mediaItem?.Type]);

	useEffect(() => {
		document.title = `aura | ${mediaItem?.Title || partialMediaItem?.Title || "Media Item"}`;
	}, [mediaItem?.Title, partialMediaItem?.Title]);

	// 1. If no partial media item, show error and stop further effects
	useEffect(() => {
		if (!isMounted.current) {
			isMounted.current = true;
			return;
		}
		if (!partialMediaItem) {
			setHasError(true);
			setError(ReturnErrorMessage("No media item selected. Please go back and select a media item."));
			setMediaItemLoading(false);
			setUserFollowsHidesLoading(false);
			setPosterSetsLoading(false);
			return;
		}
	}, [partialMediaItem]);

	// 2. Fetch full media item from server when partialMediaItem and librarySectionsMap are ready
	useEffect(() => {
		if (!librarySectionsHasHydrated) return;
		if (!partialMediaItem || Object.keys(librarySectionsMap).length === 0) return;

		setError(null);

		const fetchMediaItem = async () => {
			try {
				setMediaItemLoading(true);
				setLoadingMessage(`Loading Details for ${partialMediaItem.Title}...`);
				log(
					"INFO",
					"Media Item Page",
					"Fetch",
					`Getting full media item for ${partialMediaItem.Title} (${partialMediaItem.RatingKey})`
				);
				const resp = await fetchMediaServerItemContent(
					partialMediaItem.RatingKey,
					partialMediaItem.LibraryTitle
				);
				if (resp.status === "error") {
					setError(resp);
					setHasError(true);
					setMediaItemLoading(false);
					return;
				}

				const responseItem = resp.data;
				if (!responseItem) {
					setError(ReturnErrorMessage("No media item found in response from server."));
					setHasError(true);
					setMediaItemLoading(false);
					return;
				}
				log("INFO", "Media Item Page", "Fetch", `Fetched full media item`, responseItem);
				setMediaItem(responseItem);

				// Find if this item exists in other sections
				const otherSections = Object.values(librarySectionsMap).filter(
					(s) => s.Type === responseItem.Type && s.Title !== responseItem.LibraryTitle
				);

				if (otherSections && otherSections.length > 0) {
					log(
						"INFO",
						"Media Item Page",
						"Fetch",
						`Found other sections of type ${responseItem.Type}`,
						otherSections
					);
					let foundOther: MediaItem | null = null;
					if (responseItem.Guids?.length) {
						const tmdbID = responseItem.Guids.find((guid) => guid.Provider === "tmdb")?.ID;
						if (tmdbID) {
							for (const section of otherSections) {
								if (!section.MediaItems || section.MediaItems.length === 0) continue;
								const otherMediaItem = section.MediaItems.find((item) =>
									item.Guids?.some((guid) => guid.Provider === "tmdb" && guid.ID === tmdbID)
								);
								if (otherMediaItem) {
									foundOther = otherMediaItem;
									break;
								}
							}
						}
					}
					log("INFO", "Media Item Page", "Fetch", `Media Item - Exists in other sections?`, foundOther);
					setExistsInOtherSections(foundOther);
				}
			} catch (err) {
				log("ERROR", "Media Item Page", "Fetch", "Exception while fetching media item", err);
				setError(ReturnErrorMessage<unknown>(err));
				setHasError(true);
				setMediaItemLoading(false);
			} finally {
				setMediaItemLoading(false);
			}
		};

		fetchMediaItem();
	}, [partialMediaItem, librarySectionsMap, librarySectionsHasHydrated]);

	// 3. Fetch user follows/hides when mediaItem is loaded
	useEffect(() => {
		if (hasError) return; // Stop if there's already an error
		if (mediaItemLoading) return; // Wait for mediaItem to finish loading
		if (!mediaItem) return; // If no mediaItem, do nothing

		setError(null);

		const fetchUserFollowHidesAsync = async () => {
			try {
				setUserFollowsHidesLoading(true);
				setLoadingMessage("Loading User Follows/Hides");
				log("INFO", "Media Item Page", "User Follows/Hides", "Fetching user preferences from Mediux");
				const response = await fetchMediuxUserFollowHides();

				if (response.status === "error") {
					log("ERROR", "Media Item Page", "User Follows/Hides", "Error fetching user preferences", response);
					setError(response);
					setHasError(true);
					setUserFollowsHidesLoading(false);
					setUserFollows([]);
					setUserHides([]);
					return;
				}

				log("INFO", "Media Item Page", "User Follows/Hides", "Fetched user preferences", response.data);
				setUserFollows(response.data?.Follows || []);
				setUserHides(response.data?.Hides || []);
			} catch (err) {
				log("ERROR", "Media Item Page", "User Follows/Hides", "Exception while fetching user preferences", err);
				setError(ReturnErrorMessage<unknown>(err));
				setHasError(true);
				setUserFollowsHidesLoading(false);
				setUserFollows([]);
				setUserHides([]);
			} finally {
				setUserFollowsHidesLoading(false);
			}
		};

		fetchUserFollowHidesAsync();
	}, [hasError, mediaItem, mediaItemLoading]);

	// 4. Fetch poster sets when mediaItem and userHides are loaded
	useEffect(() => {
		if (hasError) return; // Stop if there's already an error
		if (mediaItemLoading) return; // Wait for mediaItem to finish loading
		if (userFollowsHidesLoading) return; // Wait for user follows/hides to finish loading
		if (!mediaItem) return; // If no mediaItem, do nothing

		setError(null);

		const fetchPosterSetsAsync = async () => {
			try {
				if (!mediaItem.Guids?.length) {
					log(
						"INFO",
						"Media Item Page",
						"Poster Sets",
						"No Guids found on mediaItem, skipping poster sets fetch"
					);
					return;
				}

				const tmdbID = mediaItem.Guids.find((guid) => guid.Provider === "tmdb")?.ID;
				if (!tmdbID) {
					log("ERROR", "Media Item Page", "Poster Sets", "No TMDB ID found, cannot fetch poster sets");
					setError(ReturnErrorMessage<string>("No TMDB ID found"));
					setHasError(true);
					setPosterSets([]);
					setPosterSetsLoading(false);
					return;
				}

				setPosterSetsLoading(true);
				setLoadingMessage("Loading Poster Sets");
				log("INFO", "Media Item Page", "Poster Sets", "Fetching poster sets", {
					tmdbID,
					type: mediaItem.Type,
					libraryTitle: mediaItem.LibraryTitle,
					ratingKey: mediaItem.RatingKey,
				});
				const response = await fetchMediuxSets(
					tmdbID,
					mediaItem.Type,
					mediaItem.LibraryTitle,
					mediaItem.RatingKey
				);

				if (response.status === "error") {
					log("ERROR", "Media Item Page", "Poster Sets", "Error fetching poster sets", response);
					if (response.error?.Message && response.error.Message.startsWith("No sets found")) {
						response.error.Message = `No Poster Sets found for ${mediaItem.Title}`;
					}
					setError(response);
					setHasError(true);
					setPosterSets([]);
					setPosterSetsLoading(false);
					return;
				}

				log("INFO", "Media Item Page", "Poster Sets", "Fetched poster sets", response.data);
				setPosterSets(response.data || []);
			} catch (err) {
				log("ERROR", "Media Item Page", "Poster Sets", "Exception while fetching poster sets", err);
				setError(ReturnErrorMessage<unknown>(err));
				setHasError(true);
				setPosterSets([]);
			} finally {
				setPosterSetsLoading(false);
			}
		};

		fetchPosterSetsAsync();
	}, [hasError, mediaItem, mediaItemLoading, userFollowsHidesLoading, userHides]);

	// 6. Filtering logic for poster sets
	useEffect(() => {
		if (hasError) return; // Stop if there's already an error
		if (mediaItemLoading) return; // Wait for mediaItem to finish loading
		if (userFollowsHidesLoading) return; // Wait for user follows/hides to finish loading
		if (posterSetsLoading) return; // Wait for poster sets to finish loading
		if (!mediaItem) return; // If no mediaItem, do nothing
		if (!posterSets || posterSets.length === 0) return; // If no poster sets, do nothing

		log("INFO", "Media Item Page", "Filters Sets", "Applying filters to poster sets", {
			posterSets,
			showHiddenUsers,
			userHides,
			userFollows,
			sortOption,
			sortOrder,
			mediaItem,
			showOnlyTitlecardSets,
		});

		let filtered = posterSets.filter((set) => {
			if (showHiddenUsers) return true;
			const isHidden = userHides.some((hide) => hide.Username === set.User.Name);
			return !isHidden;
		});

		// If there is no titlecard sets, then showOnlyTitlecardSets should be false
		if (mediaItem && mediaItem.Type === "show") {
			const hasTitlecardSets = posterSets.some(
				(set) => Array.isArray(set.TitleCards) && set.TitleCards.length > 0
			);
			if (!hasTitlecardSets) {
				setShowOnlyTitlecardSets(false);
			}
		}

		if (mediaItem && mediaItem.Type === "show" && showOnlyTitlecardSets) {
			filtered = filtered.filter((set) => set.TitleCards && set.TitleCards.length > 0);
		}

		// Sort the filtered poster sets
		filtered.sort((a, b) => {
			const isAFollow = userFollows.some((follow) => follow.Username === a.User.Name);
			const isBFollow = userFollows.some((follow) => follow.Username === b.User.Name);
			if (isAFollow && !isBFollow) return -1;
			if (!isAFollow && isBFollow) return 1;

			if (sortOption === "user") {
				return sortOrder === "asc"
					? a.User.Name.localeCompare(b.User.Name)
					: b.User.Name.localeCompare(a.User.Name);
			}

			const dateA = new Date(a.DateUpdated);
			const dateB = new Date(b.DateUpdated);
			if (sortOption === "date") {
				return sortOrder === "asc" ? dateA.getTime() - dateB.getTime() : dateB.getTime() - dateA.getTime();
			}

			if (mediaItem?.Type === "show" && sortOption === "numberOfSeasons") {
				const seasonsA = a.SeasonPosters ? a.SeasonPosters.length : 0;
				const seasonsB = b.SeasonPosters ? b.SeasonPosters.length : 0;
				return sortOrder === "asc" ? seasonsA - seasonsB : seasonsB - seasonsA;
			}

			if (mediaItem?.Type === "show" && sortOption === "numberOfTitlecards") {
				const titlecardsA = a.TitleCards ? a.TitleCards.length : 0;
				const titlecardsB = b.TitleCards ? b.TitleCards.length : 0;
				return sortOrder === "asc" ? titlecardsA - titlecardsB : titlecardsB - titlecardsA;
			}

			if (mediaItem?.Type === "movie" && sortOption === "numberOfItemsInCollection") {
				const countAPosters = a.OtherPosters?.length ?? 0;
				const countABackdrops = a.OtherBackdrops?.length ?? 0;
				const countAMax = Math.max(countAPosters, countABackdrops);
				const countASum = countAPosters + countABackdrops;

				const countBPosters = b.OtherPosters?.length ?? 0;
				const countBBackdrops = b.OtherBackdrops?.length ?? 0;
				const countBMax = Math.max(countBPosters, countBBackdrops);
				const countBSum = countBPosters + countBBackdrops;

				if (countAMax === countBMax) {
					return sortOrder === "asc" ? countASum - countBSum : countBSum - countASum;
				}

				return sortOrder === "asc" ? countAMax - countBMax : countBMax - countAMax;
			}

			return dateB.getTime() - dateA.getTime();
		});

		log("INFO", "Media Item Page", "Filters Sets", "Filtered and sorted poster sets", filtered);
		setFilteredPosterSets(filtered);
	}, [
		posterSets,
		showHiddenUsers,
		userHides,
		userFollows,
		sortOption,
		sortOrder,
		mediaItem,
		showOnlyTitlecardSets,
		setShowOnlyTitlecardSets,
		hasError,
		mediaItemLoading,
		userFollowsHidesLoading,
		posterSetsLoading,
	]);

	// 7. Compute hiddenCount based on posterSets and userHides
	const hiddenCount = useMemo(() => {
		if (!posterSets) return 0;
		if (!userHides || userHides.length === 0) return 0;
		const uniqueHiddenUsers = new Set<string>();
		posterSets.forEach((set) => {
			if (userHides.some((hide) => hide.Username === set.User.Name)) {
				uniqueHiddenUsers.add(set.User.Name);
			}
		});
		return uniqueHiddenUsers.size;
	}, [posterSets, userHides]);

	// 8. Compute adjacent items when mediaItem changes
	useEffect(() => {
		if (!mediaItem?.RatingKey) return;
		setAdjacentItems({
			previous: getAdjacentMediaItem(mediaItem.RatingKey, "previous"),
			next: getAdjacentMediaItem(mediaItem.RatingKey, "next"),
		});
	}, [mediaItem?.RatingKey]);

	const handleShowSetsWithTitleCardsOnly = () => {
		setShowOnlyTitlecardSets(!showOnlyTitlecardSets);
	};

	const handleShowHiddenUsers = () => {
		setShowHiddenUsers(!showHiddenUsers);
	};

	const handleMediaItemChange = (item: MediaItem) => {
		if (item.ExistInDatabase) {
			setExistsInDB(true);
		}
		setMediaItem(item);
		setImageVersion(Date.now());
	};

	if (!partialMediaItem && !mediaItem && hasError) {
		return (
			<div className="flex flex-col items-center">
				<ErrorMessage error={error} />
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

	if (!mediaItem && mediaItemLoading) {
		return (
			<div className={cn("mt-4 flex flex-col items-center", hasError ? "hidden" : "block")}>
				<Loader message={loadingMessage} />
			</div>
		);
	}

	if (!mediaItem && hasError) {
		return (
			<div className="flex flex-col items-center">
				<ErrorMessage error={error} />
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

	return (
		<>
			<DimmedBackground
				backdropURL={`/api/mediaserver/image/${mediaItem?.RatingKey}/backdrop?cb=${imageVersion}`}
			/>

			{/* Navigation Buttons */}
			<div className="flex justify-between mt-2 mx-2">
				<div>
					{adjacentItems.previous && adjacentItems.previous.RatingKey && (
						<ArrowLeftCircle
							className="h-8 w-8 text-primary hover:text-primary/80 transition-colors cursor-pointer"
							onClick={() => {
								useMediaStore.setState({
									mediaItem: adjacentItems.previous,
								});
								router.push(formatMediaItemUrl(adjacentItems.previous!));
							}}
						/>
					)}
				</div>
				<div>
					{adjacentItems.next && (
						<ArrowRightCircle
							className="h-8 w-8 text-primary hover:text-primary/80 transition-colors cursor-pointer"
							onClick={() => {
								useMediaStore.setState({
									mediaItem: adjacentItems.next,
								});
								router.push(formatMediaItemUrl(adjacentItems.next!));
							}}
						/>
					)}
				</div>
			</div>

			{/* Header */}
			<div className="p-4 lg:p-6">
				<div className="pb-6">
					<MediaItemDetails
						ratingKey={mediaItem?.RatingKey || ""}
						mediaItemType={mediaItem?.Type || ""}
						title={mediaItem?.Title || ""}
						summary={mediaItem?.Summary || ""}
						year={mediaItem?.Year || 0}
						contentRating={mediaItem?.ContentRating || ""}
						seasonCount={mediaItem?.Type === "show" ? mediaItem?.Series?.SeasonCount || 0 : 0}
						episodeCount={mediaItem?.Type === "show" ? mediaItem?.Series?.EpisodeCount || 0 : 0}
						moviePath={mediaItem?.Movie?.File?.Path || "N/A"}
						movieSize={mediaItem?.Movie?.File?.Size || 0}
						movieDuration={mediaItem?.Movie?.File?.Duration || 0}
						guids={mediaItem?.Guids || []}
						existsInDB={existsInDB}
						onExistsInDBChange={setExistsInDB}
						status={posterSets ? posterSets[0]?.Status : ""}
						libraryTitle={mediaItem?.LibraryTitle || ""}
						otherMediaItem={existsInOtherSections}
					/>

					{isLoading && (
						<div className={cn("mt-4 flex flex-col items-center", hasError ? "hidden" : "block")}>
							<Loader message={loadingMessage} />
						</div>
					)}
					{hasError && error && <ErrorMessage error={error} />}

					{/* Render filtered poster sets */}
					{posterSets && posterSets.length > 0 && mediaItem && (
						<>
							<div
								className="flex flex-col sm:flex-row sm:justify-end mb-6 pr-0 sm:pr-4 items-stretch sm:items-center gap-3 sm:gap-4 w-full"
								style={{
									background: "oklch(0.16 0.0202 282.55)",
									opacity: "0.95",
									padding: "0.5rem",
								}}
							>
								{hiddenCount === 0 ? (
									<span className="text-sm text-muted-foreground ml-2">No hidden users</span>
								) : (
									<div className="flex items-center space-x-2">
										<Checkbox
											checked={showHiddenUsers}
											onCheckedChange={handleShowHiddenUsers}
											disabled={hiddenCount === 0}
											className="h-5 w-5 sm:h-4 sm:w-4 flex-shrink-0 rounded-xs ml-2 sm:ml-0 cursor-pointer"
										/>
										{showHiddenUsers ? (
											<span className="text-sm ml-2">Showing all users</span>
										) : (
											<span className="text-sm ml-2">
												Show {hiddenCount} hidden user
												{hiddenCount > 1 ? "s" : ""}
											</span>
										)}
									</div>
								)}

								{mediaItem?.Type === "show" &&
									posterSets.some(
										(set) => Array.isArray(set.TitleCards) && set.TitleCards.length > 0
									) && (
										<div className="flex items-center space-x-2">
											<Checkbox
												checked={showOnlyTitlecardSets}
												onCheckedChange={handleShowSetsWithTitleCardsOnly}
												className="h-5 w-5 sm:h-4 sm:w-4 flex-shrink-0 rounded-xs ml-2 sm:ml-0 cursor-pointer"
											/>
											{showOnlyTitlecardSets ? (
												<span className="text-sm ml-2">Showing Titlecard Sets Only</span>
											) : (
												<span className="text-sm ml-2">Filter Titlecard Only Sets</span>
											)}
										</div>
									)}

								{/* Sorting controls */}
								<SortControl
									options={[
										{
											value: "date",
											label: "Date Updated",
											ascIcon: <CalendarArrowUp />,
											descIcon: <CalendarArrowDown />,
										},
										{
											value: "user",
											label: "User Name",
											ascIcon: <ArrowDownAZ />,
											descIcon: <ArrowDownZA />,
										},
										...(mediaItem?.Type === "movie"
											? [
													{
														value: "numberOfItemsInCollection",
														label: "Number in Collection",
														ascIcon: <ArrowDown01 />,
														descIcon: <ArrowDown10 />,
													},
												]
											: []),
										...(mediaItem?.Type === "show"
											? [
													{
														value: "numberOfSeasons",
														label: "Number of Seasons",
														ascIcon: <ArrowDown01 />,
														descIcon: <ArrowDown10 />,
													},
													{
														value: "numberOfTitlecards",
														label: "Number of Titlecards",
														ascIcon: <ArrowDown01 />,
														descIcon: <ArrowDown10 />,
													},
												]
											: []),
									]}
									sortOption={sortOption}
									sortOrder={sortOrder}
									setSortOption={(value) => {
										setSortOption(value as "user" | "date" | "");
										if (value === "user") setSortOrder("asc");
										else if (value === "date") setSortOrder("desc");
									}}
									setSortOrder={setSortOrder}
									showLabel={false}
								/>
							</div>

							<div className="text-center">
								{filteredPosterSets && filteredPosterSets.length !== posterSets.length ? (
									<p className="text-sm text-muted-foreground">
										Showing {filteredPosterSets.length} of {posterSets.length} Poster Set
										{posterSets.length > 1 ? "s" : ""}
									</p>
								) : (
									<p className="text-sm text-muted-foreground">
										{posterSets.length} Poster Set{posterSets.length > 1 ? "s" : ""}
									</p>
								)}
							</div>

							{filteredPosterSets && filteredPosterSets.length === 0 && posterSets.length > 0 && (
								<div className="flex flex-col items-center">
									<ErrorMessage
										error={ReturnErrorMessage<string>(
											"All sets are hidden. Check your filters or hidden users."
										)}
									/>
									{!showHiddenUsers && (
										<Button className="mt-4" variant="secondary" onClick={handleShowHiddenUsers}>
											Show Hidden Users
										</Button>
									)}
									{mediaItem?.Type === "show" && (
										<Button
											className="mt-4"
											variant="secondary"
											onClick={handleShowSetsWithTitleCardsOnly}
										>
											Show Non-Titlecard Sets
										</Button>
									)}
								</div>
							)}

							<div className="divide-y divide-primary-dynamic/20 space-y-6">
								{(filteredPosterSets ?? []).map((set) => (
									<div key={set.ID} className="pb-6">
										<MediaCarousel
											set={set}
											mediaItem={mediaItem}
											onMediaItemChange={handleMediaItemChange}
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
