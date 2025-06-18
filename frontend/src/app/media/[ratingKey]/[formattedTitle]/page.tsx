"use client";

import { formatMediaItemUrl } from "@/helper/formatMediaItemURL";
import { getAdjacentMediaItemFromIDB, getAllLibrarySectionsFromIDB } from "@/helper/searchIDBForTMDBID";
import { fetchMediaServerItemContent } from "@/services/api.mediaserver";
import { fetchMediuxSets, fetchMediuxUserFollowHides } from "@/services/api.mediux";
import { ReturnErrorMessage } from "@/services/api.shared";
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
import { SortControl } from "@/components/shared/sort-control";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";

import { log } from "@/lib/logger";
import { useMediaStore } from "@/lib/mediaStore";
import { storage } from "@/lib/storage";

import { APIResponse } from "@/types/apiResponse";
import { MediaItem } from "@/types/mediaItem";
import { PosterSet } from "@/types/posterSets";

const MediaItemPage = () => {
	const router = useRouter();

	const hasFetchedInfo = useRef(false);

	const partialMediaItem = useMediaStore((state) => state.mediaItem); // Retrieve partial mediaItem from Zustand

	const [mediaItem, setMediaItem] = useState<MediaItem | null>(partialMediaItem);

	const [posterSets, setPosterSets] = useState<PosterSet[] | null>(null);
	const [filteredPosterSets, setFilteredPosterSets] = useState<PosterSet[] | null>(null);

	// State to track the selected sorting option
	const [sortOption, setSortOption] = useState<string>("date");
	const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");

	// Loading State
	const [isLoading, setIsLoading] = useState(true);
	const [loadingMessage, setLoadingMessage] = useState("Loading...");

	// Error Handling
	const [hasError, setHasError] = useState(false);
	const [error, setError] = useState<APIResponse<unknown> | null>(null);

	const [userFollows, setUserFollows] = useState<{ ID: string; Username: string }[]>([]);
	const [userHides, setUserHides] = useState<{ ID: string; Username: string }[]>([]);
	const [showHiddenUsers, setShowHiddenUsers] = useState(false);
	const [showSetsWithTitleCardsOnly, setShowSetsWithTitleCardsOnly] = useState(false);

	const [existsInOtherSections, setExistsInOtherSections] = useState<MediaItem | null>(null);

	const [adjacentItems, setAdjacentItems] = useState<{
		previous: MediaItem | null;
		next: MediaItem | null;
	}>({
		previous: null,
		next: null,
	});

	const handleShowSetsWithTitleCardsOnly = () => {
		setShowSetsWithTitleCardsOnly((prev) => {
			// If the checkbox is checked, set it to false
			if (prev) {
				return false;
			}
			return true;
		});
	};

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

				// Fetch all sections in parallel, then check for existence in other sections
				const sections = await getAllLibrarySectionsFromIDB();
				log("Media Item Page - Media Item fetched:", responseItem);
				log("Library Sections:", sections);

				if (sections.length > 0) {
					const otherSections = sections.filter(
						(s) => s.type === responseItem.Type && s.title !== responseItem.LibraryTitle
					);

					// Fetch all section data in parallel
					const sectionDataArr = await Promise.all(
						otherSections.map((section) =>
							storage
								.getItem<{
									data: {
										MediaItems: MediaItem[];
									};
								}>(section.title)
								.then((data) => ({ section, data }))
						)
					);

					await fetchUserFollowHides();
					await fetchPosterSets(responseItem);

					const tmdbId = Array.isArray(responseItem.Guids)
						? responseItem.Guids?.find?.((guid) => guid.Provider === "tmdb")?.ID
						: currentMediaItem.Guids?.find?.((guid) => guid.Provider === "tmdb")?.ID;
					if (!tmdbId) {
						return;
					}
					for (const { section, data } of sectionDataArr) {
						log("SECTION:", section);
						log("Library Section Data:", data);

						if (data && data.data && data.data.MediaItems) {
							const otherMediaItem = data.data.MediaItems?.find?.(
								(item) => item.Guids?.find?.((guid) => guid.Provider === "tmdb")?.ID === tmdbId
							);
							log(`Checking section: ${section.title}, found item: ${otherMediaItem ? "Yes" : "No"}`);
							if (otherMediaItem) {
								log(`Media Item ${responseItem.RatingKey} exists in section: ${section.title}`);
								setExistsInOtherSections(otherMediaItem);
								break;
							}
						}
					}
				}
			} catch (error) {
				setHasError(true);
				setError(ReturnErrorMessage<unknown>(error));
				setIsLoading(false);
			}
		};

		fetchAllInfo();
	}, [partialMediaItem, mediaItem]);

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

	// Check if all sets are hidden
	const allSetsHidden = useMemo(() => {
		if (!posterSets) return false;
		return (
			posterSets.length > 0 &&
			posterSets.every((set) => userHides.some((hide) => hide.Username === set.User.Name))
		);
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

			if (mediaItem && mediaItem.Type === "show" && showSetsWithTitleCardsOnly) {
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
		showSetsWithTitleCardsOnly,
	]);

	useEffect(() => {
		const fetchAdjacentItems = async () => {
			if (!mediaItem?.LibraryTitle || !mediaItem?.RatingKey) return;

			const [previousItem, nextItem] = await Promise.all([
				getAdjacentMediaItemFromIDB(mediaItem.LibraryTitle, mediaItem.RatingKey, "previous"),
				getAdjacentMediaItemFromIDB(mediaItem.LibraryTitle, mediaItem.RatingKey, "next"),
			]);

			setAdjacentItems({
				previous: previousItem,
				next: nextItem,
			});
		};

		fetchAdjacentItems();
	}, [mediaItem?.LibraryTitle, mediaItem?.RatingKey]);

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
			<DimmedBackground backdropURL={`/api/mediaserver/image/${mediaItem.RatingKey}/backdrop`} />

			{/* Navigation Buttons */}
			<div className="flex justify-between mt-2 mx-2">
				<div>
					{adjacentItems.previous && adjacentItems.previous.RatingKey && (
						<ArrowLeftCircle
							className="h-8 w-8 text-primary hover:text-primary/80 transition-colors"
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
							className="h-8 w-8 text-primary hover:text-primary/80 transition-colors"
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
						existsInDB={mediaItem.ExistInDatabase || false}
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

					{/* Check if all poster sets are hidden */}
					{posterSets && !showHiddenUsers && allSetsHidden && (
						<div className="flex flex-col items-center">
							<ErrorMessage error={ReturnErrorMessage<string>("All poster sets are hidden.")} />
							<Button className="mt-4" variant="secondary" onClick={handleShowHiddenUsers}>
								Show Hidden Users
							</Button>
						</div>
					)}

					{/* Render filtered poster sets */}
					{filteredPosterSets && filteredPosterSets.length > 0 && (
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
											className="h-5 w-5 sm:h-4 sm:w-4 flex-shrink-0 rounded-xs ml-2 sm:ml-0"
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
									filteredPosterSets.length > 0 &&
									filteredPosterSets.some(
										(set) => Array.isArray(set.TitleCards) && set.TitleCards.length > 0
									) && (
										<div className="flex items-center space-x-2">
											<Checkbox
												checked={showSetsWithTitleCardsOnly}
												onCheckedChange={handleShowSetsWithTitleCardsOnly}
												className="h-5 w-5 sm:h-4 sm:w-4 flex-shrink-0 rounded-xs ml-2 sm:ml-0"
											/>
											{showSetsWithTitleCardsOnly ? (
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

							{filteredPosterSets.length > 2 && (
								<div className="text-center mb-5">{filteredPosterSets.length} Poster Sets</div>
							)}

							<div className="divide-y divide-primary-dynamic/20 space-y-6">
								{(filteredPosterSets ?? []).map((set) => (
									<div key={set.ID} className="pb-6">
										<MediaCarousel set={set} mediaItem={mediaItem} />
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
