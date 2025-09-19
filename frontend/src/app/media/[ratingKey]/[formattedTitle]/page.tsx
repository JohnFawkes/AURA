"use client";

import { formatMediaItemUrl } from "@/helper/format-media-item-url";
import { getAdjacentMediaItem } from "@/helper/search-idb-for-tmdb-id";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { fetchMediaServerItemContent } from "@/services/mediaserver/api-mediaserver-fetch-item-content";
import { fetchMediuxSets } from "@/services/mediux/api-mediux-fetch-sets";
import { fetchMediuxUserFollowHides } from "@/services/mediux/api-mediux-fetch-user-follow-hide";
import {
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

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";
import { useMediaStore } from "@/lib/stores/global-store-media-store";
import { useMediaPageStore } from "@/lib/stores/page-store-media";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterSet } from "@/types/media-and-posters/poster-sets";

const MediaItemPage = () => {
	const router = useRouter();

	const hasFetchedInfo = useRef(false);

	const mediaStore = useMediaStore();
	const partialMediaItem = mediaStore.mediaItem;

	const [mediaItem, setMediaItem] = useState<MediaItem | null>(partialMediaItem);

	const [posterSets, setPosterSets] = useState<PosterSet[] | null>(null);
	const [filteredPosterSets, setFilteredPosterSets] = useState<PosterSet[] | null>(null);

	// State to track the selected sorting option
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

	// Loading State
	const [isLoading, setIsLoading] = useState(true);
	const [loadingMessage, setLoadingMessage] = useState("Loading...");

	// Error Handling
	const [hasError, setHasError] = useState(false);
	const [error, setError] = useState<APIResponse<unknown> | null>(null);

	const [userFollows, setUserFollows] = useState<{ ID: string; Username: string }[]>([]);
	const [userHides, setUserHides] = useState<{ ID: string; Username: string }[]>([]);

	const [existsInOtherSections, setExistsInOtherSections] = useState<MediaItem | null>(null);
	const [imageVersion, setImageVersion] = useState(Date.now());
	const sectionsMap = useLibrarySectionsStore((state) => state.sections);
	const [existsInDB, setExistsInDB] = useState<boolean>(mediaItem?.ExistInDatabase || false);

	const [adjacentItems, setAdjacentItems] = useState<{
		previous: MediaItem | null;
		next: MediaItem | null;
	}>({
		previous: null,
		next: null,
	});

	const handleShowSetsWithTitleCardsOnly = () => {
		setShowOnlyTitlecardSets(!showOnlyTitlecardSets);
	};

	const handleShowHiddenUsers = () => {
		setShowHiddenUsers(!showHiddenUsers);
	};

	useEffect(() => {
		if (hasFetchedInfo.current) return;
		hasFetchedInfo.current = true;

		const fetchUserFollowHides = async () => {
			try {
				setLoadingMessage("Loading User Follows/Hides");
				const response = await fetchMediuxUserFollowHides();

				if (response.status === "error") {
					setError(response);
					setHasError(true);
					return;
				}

				setUserFollows(response.data?.Follows || []);
				setUserHides(response.data?.Hides || []);
			} catch (error) {
				setError(ReturnErrorMessage<unknown>(error));
				setHasError(true);
				setUserFollows([]);
				setUserHides([]);
			}
		};

		const fetchPosterSets = async (responseItem: MediaItem) => {
			try {
				if (!responseItem.Guids?.length) {
					return;
				}

				const tmdbID = responseItem.Guids.find((guid) => guid.Provider === "tmdb")?.ID;
				if (!tmdbID) {
					setError(ReturnErrorMessage<unknown>(new Error("No TMDB ID found")));
					setHasError(true);
					return;
				}

				setLoadingMessage("Loading Poster Sets");
				const response = await fetchMediuxSets(
					tmdbID,
					responseItem.Type,
					responseItem.LibraryTitle,
					responseItem.RatingKey
				);

				if (response.status === "error") {
					if (response.error?.Message && response.error.Message.startsWith("No sets found")) {
						response.error.Message = `No Poster Sets found for ${responseItem.Title}`;
					}
					setError(response);
					setHasError(true);
					return;
				}

				setPosterSets(response.data || []);
			} catch (error) {
				setError(ReturnErrorMessage<unknown>(error));
				setHasError(true);
				setPosterSets([]);
			} finally {
				setIsLoading(false);
			}
		};

		const fetchAllInfo = async () => {
			try {
				const currentMediaItem = mediaItem || useMediaStore.getState().mediaItem;
				if (!currentMediaItem) {
					throw new Error("No media item found");
				}
				setMediaItem(currentMediaItem);
				setLoadingMessage(`Loading Details for ${currentMediaItem.Title}...`);

				const resp = await fetchMediaServerItemContent(
					currentMediaItem.RatingKey,
					currentMediaItem.LibraryTitle
				);

				if (resp.status === "error") {
					setError(resp);
					setHasError(true);
					return;
				}

				const responseItem = resp.data;
				if (!responseItem) {
					throw new Error("No media item found in response");
				}
				setMediaItem(responseItem);

				log("Media Item Page - Media Item fetched:", responseItem);
				log("Media Item Page - Cached Library Sections:", sectionsMap);
				const otherSections = Object.values(sectionsMap).filter(
					(s) => s.Type === responseItem.Type && s.Title !== responseItem.LibraryTitle
				);
				log("Cached Library Sections of same type:", otherSections);

				await fetchUserFollowHides();
				await fetchPosterSets(responseItem);

				// Check the other sections for the same media item by TMDB ID
				if (!responseItem.Guids?.length) return;
				const tmdbID = responseItem.Guids.find((guid) => guid.Provider === "tmdb")?.ID;
				if (!tmdbID) return;

				for (const section of otherSections) {
					if (!section.MediaItems || section.MediaItems.length === 0) continue;
					const otherMediaItem = section.MediaItems.find((item) =>
						item.Guids?.some((guid) => guid.Provider === "tmdb" && guid.ID === tmdbID)
					);
					log(`Checking section: ${section.Title}, found item: ${otherMediaItem ? "Yes" : "No"}`);
					if (otherMediaItem) {
						log(`Media Item ${responseItem.RatingKey} exists in section: ${section.Title}`);
						setExistsInOtherSections(otherMediaItem);
						break;
					}
				}
			} catch (error) {
				setHasError(true);
				setError(ReturnErrorMessage<unknown>(error));
				setIsLoading(false);
			}
		};

		fetchAllInfo();
	}, [partialMediaItem, mediaItem, sectionsMap]);

	// Compute hiddenCount based on posterSets and userHides
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

	useEffect(() => {
		if (posterSets) {
			log("Media Item Page - Poster Sets updated:", posterSets);
			// Filter out hidden users
			let filtered = posterSets.filter((set) => {
				if (showHiddenUsers) {
					return true; // Show all if the checkbox is checked
				}
				// Check if the user is in the hides list
				const isHidden = userHides.some((hide) => hide.Username === set.User.Name);
				return !isHidden; // Show only if not hidden
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
				// If it's a show and the checkbox is checked, filter for sets with title cards
				filtered = filtered.filter((set) => set.TitleCards && set.TitleCards.length > 0);
			}

			// Sort the filtered poster sets
			// Follows should always be at the top
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

				// For date sorting: newest to oldest unless sortOrder is "asc"
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

			setFilteredPosterSets(filtered);
		}
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
	]);

	useEffect(() => {
		if (!mediaItem?.RatingKey) return;
		setAdjacentItems({
			previous: getAdjacentMediaItem(mediaItem.RatingKey, "previous"),
			next: getAdjacentMediaItem(mediaItem.RatingKey, "next"),
		});
	}, [mediaItem?.RatingKey]);

	if (!mediaItem) {
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

	const handleMediaItemChange = (item: MediaItem) => {
		if (item.ExistInDatabase) {
			setExistsInDB(true);
		}
		setMediaItem(item);
		setImageVersion(Date.now());
	};

	if (hasError) {
		if (typeof window !== "undefined") {
			// Safe to use document here.
			document.title = "aura | Error";
		}
	} else {
		if (typeof window !== "undefined") {
			// Safe to use document here.
			document.title = `aura | ${mediaItem?.Title}`;
		}
	}

	return (
		<>
			{/* Backdrop Background */}
			<DimmedBackground
				backdropURL={`/api/mediaserver/image/${mediaItem.RatingKey}/backdrop?cb=${imageVersion}`}
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
						ratingKey={mediaItem.RatingKey}
						mediaItemType={mediaItem.Type}
						title={mediaItem.Title}
						summary={mediaItem.Summary || ""}
						year={mediaItem.Year}
						contentRating={mediaItem.ContentRating || ""}
						seasonCount={mediaItem.Type === "show" ? mediaItem.Series?.SeasonCount || 0 : 0}
						episodeCount={mediaItem.Type === "show" ? mediaItem.Series?.EpisodeCount || 0 : 0}
						moviePath={mediaItem.Movie?.File?.Path || "N/A"}
						movieSize={mediaItem.Movie?.File?.Size || 0}
						movieDuration={mediaItem.Movie?.File?.Duration || 0}
						guids={mediaItem.Guids || []}
						existsInDB={existsInDB}
						onExistsInDBChange={setExistsInDB}
						status={posterSets ? posterSets[0]?.Status : ""}
						libraryTitle={mediaItem.LibraryTitle || ""}
						otherMediaItem={existsInOtherSections}
					/>

					{isLoading && (
						<div className="flex justify-center">
							<Loader message={loadingMessage} />
						</div>
					)}
					{hasError && (
						<div className="flex justify-center">
							<ErrorMessage error={error} />
						</div>
					)}

					{/* Render filtered poster sets */}
					{posterSets && posterSets.length > 0 && (
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

								{mediaItem.Type === "show" &&
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

							{/* 
							If all poster sets are filtered out, show a message 
							This can happen if all users are hidden or the titlecard filter is applied
							*/}
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
									{mediaItem.Type === "show" && (
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
