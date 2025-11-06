"use client";

import { ReturnErrorMessage } from "@/services/api-error-return";
import { fetchMediaServerItemContent } from "@/services/mediaserver/api-mediaserver-fetch-item-content";
import { ArrowDown01, ArrowDown10, ArrowDownAZ, ArrowDownZA, CalendarArrowDown, CalendarArrowUp } from "lucide-react";

import { useEffect, useMemo, useRef, useState } from "react";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { DimmedBackground } from "@/components/shared/dimmed_backdrop";
import { ErrorMessage } from "@/components/shared/error-message";
import { MediaItemFilter } from "@/components/shared/filter-media-item";
import Loader from "@/components/shared/loader";
import { MediaCarousel } from "@/components/shared/media-carousel";
import { MediaItemDetails } from "@/components/shared/media-item-details";
import { PopoverHelp } from "@/components/shared/popover-help";
import { SortControl } from "@/components/shared/select-sort";
import { Button } from "@/components/ui/button";

import { cn } from "@/lib/cn";
import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";
import { useMediaStore } from "@/lib/stores/global-store-media-store";
import { useUserPreferencesStore } from "@/lib/stores/global-user-preferences";
import { useHomePageStore } from "@/lib/stores/page-store-home";
import { useMediaPageStore } from "@/lib/stores/page-store-media";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterSet } from "@/types/media-and-posters/poster-sets";

const MediaItemPage = () => {
	const router = useRouter();
	const isMounted = useRef(false);

	// Partial Media Item States
	const partialMediaItem = useMediaStore((state) => state.mediaItem);

	// Library Sections States (From Library Section Store)
	const librarySectionsMap = useLibrarySectionsStore((state) => state.sections);
	const librarySectionsHasHydrated = useLibrarySectionsStore((state) => state.hasHydrated);

	// Response Loading State
	const [responseLoading, setResponseLoading] = useState<boolean>(true);

	// Main Media Item States
	const [mediaItem, setMediaItem] = useState<MediaItem | null>(null);
	const [existsInOtherSections, setExistsInOtherSections] = useState<MediaItem | null>(null);
	const [existsInDB, setExistsInDB] = useState<boolean>(mediaItem?.ExistInDatabase || false);
	const [serverType, setServerType] = useState<string>("Media Server");

	// User Follows/Hides States
	const [userFollows, setUserFollows] = useState<{ ID: string; Username: string }[]>([]);
	const [userHides, setUserHides] = useState<{ ID: string; Username: string }[]>([]);

	// Poster Sets States
	const [posterSets, setPosterSets] = useState<PosterSet[] | null>(null);
	const [filteredPosterSets, setFilteredPosterSets] = useState<PosterSet[] | null>(null);

	// UI States from Media Page Store
	const {
		sortStates,
		setSortOption,
		setSortOrder,
		showOnlyTitlecardSets,
		setShowOnlyTitlecardSets,
		showHiddenUsers,
		setShowHiddenUsers,
	} = useMediaPageStore();
	const sortType = partialMediaItem?.Type as "movie" | "show";
	const sortOption = sortStates[sortType]?.sortOption ?? "date";
	const sortOrder = sortStates[sortType]?.sortOrder ?? "desc";

	// Download Defaults from User Preferences Store
	const downloadDefaultsTypes = useUserPreferencesStore((state) => state.downloadDefaults);
	const showOnlyDownloadDefaults = useUserPreferencesStore((state) => state.showOnlyDownloadDefaults);

	// Loading States
	const [loadingMessage, setLoadingMessage] = useState("Loading...");
	const isLoading = useMemo(() => {
		return responseLoading;
	}, [responseLoading]);

	// Error States
	const [hasError, setHasError] = useState(false);
	const [error, setError] = useState<APIResponse<unknown> | null>(null);

	// Image Version State (for forcing image reloads)
	const [imageVersion, setImageVersion] = useState(Date.now());

	// Get Adjacent Items from Home Page Store
	const { setNextMediaItem, setPreviousMediaItem, getAdjacentMediaItem } = useHomePageStore();

	// Update the sortOption and sortOrder based on type
	useEffect(() => {
		if (!sortType) return;
		// If the current sortOption or sortOrder is not set, initialize them
		if (!sortStates[sortType]) {
			setSortOption(sortType, "date");
			setSortOrder(sortType, "desc");
		}
	}, [sortType, sortStates, setSortOption, setSortOrder]);

	// Set the document title
	useEffect(() => {
		document.title = `aura | ${mediaItem?.Title || partialMediaItem?.Title || "Media Item"}`;
	}, [mediaItem?.Title, partialMediaItem?.Title]);

	// When the Media Item changes, set ExistsInDB
	useEffect(() => {
		if (mediaItem && mediaItem.ExistInDatabase) {
			setExistsInDB(true);
		} else {
			setExistsInDB(false);
		}
	}, [mediaItem]);

	// 1. If no partial media item, show error and stop further effects
	useEffect(() => {
		if (!isMounted.current) {
			isMounted.current = true;
			return;
		}
		if (!partialMediaItem) {
			setHasError(true);
			setError(ReturnErrorMessage("No media item selected. Please go back and select a media item."));
			setResponseLoading(false);
			return;
		}
		// If we have a partialMediaItem, reset state for new load
		setMediaItem(null);
		setResponseLoading(true);
		setHasError(false);
		setError(null);
	}, [partialMediaItem]);

	// 2. Fetch full media item from server when partialMediaItem and librarySectionsMap are ready
	useEffect(() => {
		if (!librarySectionsHasHydrated) return;
		if (!partialMediaItem || Object.keys(librarySectionsMap).length === 0) return;
		if (mediaItem && mediaItem.RatingKey === partialMediaItem.RatingKey) return;
		if (hasError) return;
		if (responseLoading === true && mediaItem !== null) return;

		setError(null);

		const fetchMediaItem = async () => {
			try {
				setResponseLoading(true);
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
					setResponseLoading(false);
					return;
				}

				const mediaItemPageResponse = resp.data;
				const errorResponse = resp.data?.error;

				if (!mediaItemPageResponse) {
					setError(ReturnErrorMessage("No data found in response from server."));
					setHasError(true);
					setResponseLoading(false);
					return;
				}

				const serverTypeResponse = mediaItemPageResponse.serverType;
				const mediaItemResponse = mediaItemPageResponse.mediaItem;
				const posterSetsResponse = mediaItemPageResponse.posterSets;
				const userFollowHideResponse = mediaItemPageResponse.userFollowHide;

				log("INFO", "Media Item Page", "Fetch", `Server Type: ${serverTypeResponse}`, { serverTypeResponse });
				log("INFO", "Media Item Page", "Fetch", `Full Media Item Response`, { mediaItemResponse });
				log("INFO", "Media Item Page", "Fetch", `Poster Sets Response`, { posterSetsResponse });
				log("INFO", "Media Item Page", "Fetch", `User Follow/Hide Response`, { userFollowHideResponse });
				log("INFO", "Media Item Page", "Fetch", `Error Response`, { errorResponse });

				// Check to see if the serverTypeResponse is valid
				// Valid types are "Plex", "Emby", "Jellyfin"
				if (!["Plex", "Emby", "Jellyfin"].includes(serverTypeResponse)) {
					setServerType("Media Server");
				} else {
					setServerType(serverTypeResponse);
				}

				// Check to see if mediaItemResponse is valid
				if (
					!mediaItemResponse ||
					!mediaItemResponse.RatingKey ||
					!mediaItemResponse.Title ||
					!mediaItemResponse.LibraryTitle
				) {
					setError(ReturnErrorMessage("Invalid media item data found in response from server."));
					log("ERROR", "Media Item Page", "Fetch", "Invalid media item data", { mediaItemResponse });
					setHasError(true);
					setResponseLoading(false);
					return;
				} else {
					setMediaItem(mediaItemResponse);
				}

				// Find if this item exists in other sections
				const otherSections = Object.values(librarySectionsMap).filter(
					(s) => s.Type === mediaItemResponse.Type && s.Title !== mediaItemResponse.LibraryTitle
				);

				if (otherSections && otherSections.length > 0) {
					log(
						"INFO",
						"Media Item Page",
						"Fetch",
						`Found other sections of type ${mediaItemResponse.Type}`,
						otherSections
					);
					let foundOther: MediaItem | null = null;
					if (mediaItemResponse.Guids?.length) {
						const tmdbID = mediaItemResponse.Guids.find((guid) => guid.Provider === "tmdb")?.ID;
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

				// Check Poster Sets
				if (posterSetsResponse && Array.isArray(posterSetsResponse)) {
					setPosterSets(posterSetsResponse);
				} else {
					setPosterSets([]);
					setResponseLoading(false);
					setHasError(true);
					setError({
						status: "error",
						error: {
							message: errorResponse?.message || `No poster sets found for '${mediaItemResponse.Title}'`,
							help: errorResponse?.help || "",
							detail: errorResponse?.detail ?? undefined,
							function: errorResponse?.function || "Unknown",
							line_number: errorResponse?.line_number || 0,
						},
					});
				}

				// Check User Follows/Hides
				if (userFollowHideResponse) {
					setUserFollows(userFollowHideResponse.Follows || []);
					setUserHides(userFollowHideResponse.Hides || []);
				} else {
					setUserFollows([]);
					setUserHides([]);
				}
			} catch (err) {
				log("ERROR", "Media Item Page", "Fetch", "Exception while fetching media item", err);
				setError(ReturnErrorMessage<unknown>(err));
				setHasError(true);
				setResponseLoading(false);
			} finally {
				setResponseLoading(false);
			}
		};

		fetchMediaItem();
	}, [partialMediaItem, librarySectionsMap, librarySectionsHasHydrated, mediaItem, hasError, responseLoading]);

	// 3. Filtering logic for poster sets
	useEffect(() => {
		if (hasError) return; // Stop if there's already an error
		if (responseLoading) return; // Wait for response to finish loading
		if (!mediaItem) return; // If no mediaItem, do nothing
		if (!posterSets || posterSets.length === 0) return; // If no poster sets, do nothing

		log("INFO", "Media Item Page", "Filters Sets", "Applying filters to poster sets", {
			posterSets,
			showHiddenUsers,
			userHides,
			userFollows,
			sortStates,
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

		// If showOnlyDownloadDefaults is true, check sets to see if they have at least one of the download default types
		if (showOnlyDownloadDefaults && downloadDefaultsTypes && downloadDefaultsTypes.length > 0) {
			filtered = filtered.filter((set) => {
				for (const imageType of downloadDefaultsTypes) {
					if (imageType === "poster" && set.Poster) return true;
					if (imageType === "poster" && set.OtherPosters && set.OtherPosters.length > 0) return true;
					if (imageType === "backdrop" && set.Backdrop) return true;
					if (imageType === "backdrop" && set.OtherBackdrops && set.OtherBackdrops.length > 0) return true;
					if (imageType === "seasonPoster" && set.SeasonPosters && set.SeasonPosters.length > 0) return true;
					if (
						imageType === "specialSeasonPoster" &&
						set.SeasonPosters &&
						set.SeasonPosters.length > 0 &&
						set.SeasonPosters.some((sp) => sp.Season?.Number === 0)
					)
						return true;
					if (imageType === "titlecard" && set.TitleCards && set.TitleCards.length > 0) return true;
				}
				return false;
			});
		}

		const downloadedSetIDs = new Set(mediaItem.DBSavedSets?.map((s) => s.PosterSetID));
		// Sort the filtered poster sets
		filtered.sort((a, b) => {
			const aDownloaded = downloadedSetIDs.has(a.ID);
			const bDownloaded = downloadedSetIDs.has(b.ID);

			if (aDownloaded && !bDownloaded) return -1;
			if (!aDownloaded && bDownloaded) return 1;

			const isAFollow = userFollows.some((follow) => follow.Username === a.User.Name);
			const isBFollow = userFollows.some((follow) => follow.Username === b.User.Name);
			if (isAFollow && !isBFollow) return -1;
			if (!isAFollow && isBFollow) return 1;

			if (sortOption === "user") {
				// If users are the same, sort by date updated
				if (a.User.Name === b.User.Name) {
					const dateA = new Date(a.DateUpdated);
					const dateB = new Date(b.DateUpdated);
					return dateB.getTime() - dateA.getTime();
				}
				// Otherwise, sort by user name
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
				if (seasonsA === seasonsB) {
					// If number of seasons are equal, sort by number of titlecards
					const titlecardsA = a.TitleCards ? a.TitleCards.length : 0;
					const titlecardsB = b.TitleCards ? b.TitleCards.length : 0;

					if (titlecardsA === titlecardsB) {
						// If number of titlecards are also equal, sort by date
						return dateB.getTime() - dateA.getTime();
					}
					return sortOrder === "asc" ? titlecardsA - titlecardsB : titlecardsB - titlecardsA;
				}
				return sortOrder === "asc" ? seasonsA - seasonsB : seasonsB - seasonsA;
			}

			if (mediaItem?.Type === "show" && sortOption === "numberOfTitlecards") {
				const titlecardsA = a.TitleCards ? a.TitleCards.length : 0;
				const titlecardsB = b.TitleCards ? b.TitleCards.length : 0;
				if (titlecardsA === titlecardsB) {
					// If number of titlecards are equal, sort by date
					return dateB.getTime() - dateA.getTime();
				}
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
					// If max counts are equal, sort by sum of counts
					if (countASum === countBSum) {
						// If sum of counts are also equal, sort by date
						return dateB.getTime() - dateA.getTime();
					}
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
		sortStates,
		mediaItem,
		showOnlyTitlecardSets,
		setShowOnlyTitlecardSets,
		downloadDefaultsTypes,
		showOnlyDownloadDefaults,
		hasError,
		responseLoading,
		sortOption,
		sortOrder,
	]);

	// 4. Compute hiddenCount based on posterSets and userHides
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

	// 5. Compute adjacent items when mediaItem changes
	useEffect(() => {
		if (!mediaItem) return;
		if (!mediaItem?.RatingKey) return;
		setNextMediaItem(getAdjacentMediaItem(mediaItem.TMDB_ID, "next"));
		setPreviousMediaItem(getAdjacentMediaItem(mediaItem.TMDB_ID, "previous"));
	}, [getAdjacentMediaItem, mediaItem, setNextMediaItem, setPreviousMediaItem]);

	const handleShowSetsWithTitleCardsOnly = () => {
		setShowOnlyTitlecardSets(!showOnlyTitlecardSets);
	};

	const handleShowHiddenUsers = () => {
		setShowHiddenUsers(!showHiddenUsers);
	};

	const handleMediaItemChange = (item: MediaItem) => {
		setImageVersion(Date.now());
		if (item.ExistInDatabase && item.TMDB_ID === mediaItem?.TMDB_ID) {
			setExistsInDB(true);
		}
	};

	// Calculate number of active filters
	const numberOfActiveFilters = useMemo(() => {
		let count = 0;
		if (!showHiddenUsers) count++;
		if (showOnlyTitlecardSets) count++;
		if (showOnlyDownloadDefaults) count++;
		return count;
	}, [showHiddenUsers, showOnlyTitlecardSets, showOnlyDownloadDefaults]);

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

	if (responseLoading) {
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
				backdropURL={`/api/mediaserver/image?ratingKey=${mediaItem?.RatingKey}&imageType=backdrop&cb=${imageVersion}`}
			/>

			<div className="p-4 lg:p-6">
				<div className="pb-6">
					{/* Header */}
					<MediaItemDetails
						mediaItem={mediaItem || undefined}
						existsInDB={existsInDB}
						onExistsInDBChange={setExistsInDB}
						status={posterSets ? posterSets[0]?.Status : ""}
						otherMediaItem={existsInOtherSections}
						serverType={serverType}
						posterImageKeys={
							[
								mediaItem?.RatingKey,
								...(mediaItem?.Series?.Seasons?.map((season) => season.RatingKey) || []),
							].filter(Boolean) as string[]
						}
					/>

					{/* Loading and Error States */}
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
								className="flex flex-col w-full mb-4 gap-4 justify-center items-center sm:justify-between sm:items-center sm:flex-row"
								style={{
									background: "oklch(0.16 0.0202 282.55)",
									opacity: "0.95",
									padding: "0.5rem",
								}}
							>
								{/* Left column: Filters */}
								<MediaItemFilter
									numberOfActiveFilters={numberOfActiveFilters}
									hiddenCount={hiddenCount}
									showHiddenUsers={showHiddenUsers}
									handleShowHiddenUsers={handleShowHiddenUsers}
									hasTitleCards={
										mediaItem?.Type === "show"
											? posterSets.some(
													(set) => Array.isArray(set.TitleCards) && set.TitleCards.length > 0
												)
											: false
									}
									showOnlyTitlecardSets={showOnlyTitlecardSets}
									handleShowSetsWithTitleCardsOnly={handleShowSetsWithTitleCardsOnly}
									showOnlyDownloadDefaults={showOnlyDownloadDefaults}
								/>

								{/* Right column: sort options */}
								<div className="flex items-center sm:justify-end sm:ml-4">
									<SortControl
										options={[
											{
												value: "date",
												label: "Date Updated",
												ascIcon: <CalendarArrowUp />,
												descIcon: <CalendarArrowDown />,
												type: "date",
											},
											{
												value: "user",
												label: "User Name",
												ascIcon: <ArrowDownAZ />,
												descIcon: <ArrowDownZA />,
												type: "string",
											},
											...(mediaItem?.Type === "movie"
												? [
														{
															value: "numberOfItemsInCollection",
															label: "Number in Collection",
															ascIcon: <ArrowDown01 />,
															descIcon: <ArrowDown10 />,
															type: "number" as const,
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
															type: "number" as const,
														},
														{
															value: "numberOfTitlecards",
															label: "Number of Titlecards",
															ascIcon: <ArrowDown01 />,
															descIcon: <ArrowDown10 />,
															type: "number" as const,
														},
													]
												: []),
										]}
										sortOption={sortOption}
										sortOrder={sortOrder}
										setSortOption={(option) => setSortOption(sortType, option)}
										setSortOrder={(order) => setSortOrder(sortType, order)}
										showLabel={false}
									/>
								</div>
							</div>

							<div className="text-center mb-4">
								{filteredPosterSets && filteredPosterSets.length !== posterSets.length ? (
									<div className="flex items-center justify-center gap-2 text-sm text-muted-foreground">
										<span>
											Showing {filteredPosterSets.length} of {posterSets.length} Poster Set
											{posterSets.length > 1 ? "s" : ""}
										</span>
										<PopoverHelp ariaLabel="help-filters">
											<p className="mb-2">
												Some of your sets are being hidden by{" "}
												{`${numberOfActiveFilters ? `${numberOfActiveFilters} active filter${numberOfActiveFilters > 1 ? "s" : ""}` : "no filters"}`}
												.
											</p>
											<ul className="list-disc list-inside mb-2">
												{hiddenCount > 0 && (
													<li>
														You have {hiddenCount} hidden user
														{hiddenCount > 1 ? "s" : ""}.{" "}
													</li>
												)}
												{mediaItem?.Type === "show" &&
													showOnlyTitlecardSets &&
													posterSets.some(
														(set) =>
															Array.isArray(set.TitleCards) && set.TitleCards.length > 0
													) && <li>You are filtering to show only titlecard sets.</li>}
												{showOnlyDownloadDefaults &&
													downloadDefaultsTypes &&
													downloadDefaultsTypes.length > 0 && (
														<li>
															You are filtering to show only sets with your selected
															download default types.
														</li>
													)}
											</ul>
											<p>
												You can adjust your filters using the checkboxes on this page. You can
												also adjust your default download image types in{" "}
												<Link href="/settings#preferences-section" className="underline">
													User Preferences
												</Link>
												.
											</p>
										</PopoverHelp>
									</div>
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
