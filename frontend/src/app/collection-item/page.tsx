"use client";

import CollectionsDownloadModal from "@/app/collection-item/collection-download-modal";
import { CollectionItem } from "@/app/collections/page";
import { formatLastUpdatedDate } from "@/helper/format-date-last-updates";
import { ReturnErrorMessage } from "@/services/api-error-return";
import {
	CollectionSet,
	fetchCollectionChildrenAndPosters,
} from "@/services/mediaserver/api-mediaserver-fetch-collection-children";
import { ArrowDownAZ, ArrowDownZA, CalendarArrowDown, CalendarArrowUp, User } from "lucide-react";

import { useEffect, useMemo, useRef, useState } from "react";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { AssetImage } from "@/components/shared/asset-image";
import { CollectionItemDetails } from "@/components/shared/collection-item-details";
import { DimmedBackground } from "@/components/shared/dimmed_backdrop";
import { ErrorMessage } from "@/components/shared/error-message";
import { CollectionItemFilter } from "@/components/shared/filter-collection-item-sets";
import Loader from "@/components/shared/loader";
import { PopoverHelp } from "@/components/shared/popover-help";
import { SortControl } from "@/components/shared/select-sort";
import { Button } from "@/components/ui/button";
import { Carousel, CarouselContent, CarouselItem, CarouselNext, CarouselPrevious } from "@/components/ui/carousel";
import { Lead, P } from "@/components/ui/typography";

import { cn } from "@/lib/cn";
import { log } from "@/lib/logger";
import { useCollectionStore } from "@/lib/stores/global-store-collection-store";
import { useCollectionItemPageStore } from "@/lib/stores/page-store-collection-item";
import { useCollectionsPageStore } from "@/lib/stores/page-store-collections";

import { APIResponse } from "@/types/api/api-response";

export default function CollectionItemPage() {
	const router = useRouter();
	const isMounted = useRef(false);

	// Partial Collection Item from Store
	const partialCollectionItem = useCollectionStore((state) => state.collectionItem);

	// Main Collection Item State
	const [collectionItem, setCollectionItem] = useState<CollectionItem | null>(null);
	const [collectionItemSets, setCollectionItemSets] = useState<CollectionSet[]>([]);
	const [filteredCollectionItemSets, setFilteredCollectionItemSets] = useState<CollectionSet[]>([]);

	// User Follows/Hides States
	const [userFollows, setUserFollows] = useState<{ ID: string; Username: string }[]>([]);
	const [userHides, setUserHides] = useState<{ ID: string; Username: string }[]>([]);

	// Loading States
	const [responseLoading, setResponseLoading] = useState<boolean>(true);
	const [loadingMessage, setLoadingMessage] = useState("Loading...");
	const isLoading = useMemo(() => {
		return responseLoading;
	}, [responseLoading]);

	// Error States
	const [hasError, setHasError] = useState(false);
	const [error, setError] = useState<APIResponse<unknown> | null>(null);

	// Image Version State (for forcing image reloads)
	const imageVersion = useState(Date.now());

	// UI States from Store
	const { sortOrder, setSortOrder, sortOption, setSortOption, showHiddenUsers, setShowHiddenUsers } =
		useCollectionItemPageStore();

	const { setNextCollectionItem, setPreviousCollectionItem, getAdjacentCollectionItem } = useCollectionsPageStore();

	// Set document title
	useEffect(() => {
		const rawTitle = collectionItem?.Title || partialCollectionItem?.Title || "Collection Item";
		const title = rawTitle && !rawTitle.toLowerCase().includes("collection") ? `${rawTitle} Collection` : rawTitle;
		document.title = `aura | ${title}`;
	}, [collectionItem?.Title, partialCollectionItem?.Title]);

	// Set the default sort order and option on mount
	useEffect(() => {
		if (sortOption !== "dateUpdated" && sortOption !== "username") {
			setSortOption("dateUpdated");
			setSortOrder("desc");
		}
	}, [setSortOption, setSortOrder, sortOption]);

	// 1. If no partial collection item, show error and stop further effects
	useEffect(() => {
		if (!isMounted.current) {
			isMounted.current = true;
			return;
		}
		if (!partialCollectionItem) {
			setHasError(true);
			setError(ReturnErrorMessage("No media item selected. Please go back and select a media item."));
			setResponseLoading(false);
			return;
		}
		// If we have a partialCollectionItem, reset state for new load
		setCollectionItem(null);
		setResponseLoading(true);
		setHasError(false);
		setError(null);
	}, [partialCollectionItem]);

	// 2. Fetch full collection item details when partialCollectionItem is ready
	useEffect(() => {
		if (!partialCollectionItem) return;

		setError(null);
		const fetchFullCollectionItem = async () => {
			try {
				setResponseLoading(true);
				setLoadingMessage(`Loading Collection Item: ${partialCollectionItem.Title}`);
				log(
					"INFO",
					"Collection Item Page",
					"Fetch",
					`Fetching full collection item for: ${partialCollectionItem.Title} (${partialCollectionItem.RatingKey})`
				);

				const resp = await fetchCollectionChildrenAndPosters(partialCollectionItem);
				if (resp.status === "error") {
					setError(resp);
					setHasError(true);
					setResponseLoading(false);
					return;
				}

				if (!resp.data) {
					setError(ReturnErrorMessage("No collection item data returned from server."));
					setHasError(true);
					setResponseLoading(false);
					return;
				}

				const errorResponse = resp.data?.error;
				const collectionItem = resp.data.collection_item || null;
				const collectionItemSets = resp.data.collection_sets || [];
				const userFollowHide = resp.data.user_follow_hide || null;

				log("INFO", "Collection Item Page", "Fetch", "Collection Item Response", { collectionItem });
				log("INFO", "Collection Item Page", "Fetch", "Collection Sets Response", { collectionItemSets });
				log("INFO", "Collection Item Page", "Fetch", "User Follow/Hide Response", { userFollowHide });
				log("INFO", "Collection Item Page", "Fetch", `Error Response`, { errorResponse });

				setCollectionItem(collectionItem);

				// Check if collectionItemSets is an array
				if (collectionItemSets && Array.isArray(collectionItemSets) && collectionItemSets.length > 0) {
					setCollectionItemSets(collectionItemSets);
				} else {
					setCollectionItemSets([]);
					setResponseLoading(false);
					setHasError(true);
					setError({
						status: "error",
						error: {
							message:
								errorResponse?.message ||
								`No collection sets found for '${partialCollectionItem.Title}'`,
							help: errorResponse?.help || "",
							detail: errorResponse?.detail ?? undefined,
							function: errorResponse?.function || "Unknown",
							line_number: errorResponse?.line_number || 0,
						},
					});
				}

				if (userFollowHide) {
					setUserFollows(userFollowHide.Follows || []);
					setUserHides(userFollowHide.Hides || []);
				} else {
					setUserFollows([]);
					setUserHides([]);
				}

				setResponseLoading(false);
			} catch (error) {
				log("ERROR", "Collection Item Page", "Fetch", "Exception while fetching collection item", error);
				setError(ReturnErrorMessage<unknown>(error));
				setHasError(true);
				setResponseLoading(false);
			} finally {
				setResponseLoading(false);
			}
		};

		fetchFullCollectionItem();
	}, [partialCollectionItem]);

	// 3. Filtering Logic
	useEffect(() => {
		if (hasError) return; // Stop if there is an error
		if (responseLoading) return; // Stop if still loading
		if (!collectionItem) return; // Stop if no collection item
		if (collectionItemSets.length === 0) return; // Stop if no sets

		log(
			"INFO",
			"Collection Item Page",
			"Filter",
			`Applying filters: sortOption=${sortOption}, sortOrder=${sortOrder}, showHiddenUsers=${showHiddenUsers}`
		);

		let filtered = collectionItemSets.filter((set) => {
			if (showHiddenUsers) return true;
			const isHidden = userHides.some((hide) => hide.Username === set.User.Name);
			return !isHidden;
		});

		filtered.sort((a, b) => {
			const isAFollow = userFollows.some((follow) => follow.Username === a.User.Name);
			const isBFollow = userFollows.some((follow) => follow.Username === b.User.Name);
			if (isAFollow && !isBFollow) return -1;
			if (!isAFollow && isBFollow) return 1;

			if (sortOption === "username") {
				// If users are the same, sort by date updated
				if (a.User.Name === b.User.Name) {
					const dateA = new Date(a.Posters[0]?.Modified || "");
					const dateB = new Date(b.Posters[0]?.Modified || "");
					return dateB.getTime() - dateA.getTime();
				}
				// Otherwise, sort by user name
				return sortOrder === "asc"
					? a.User.Name.localeCompare(b.User.Name)
					: b.User.Name.localeCompare(a.User.Name);
			}

			const dateA = new Date(a.Posters[0]?.Modified || "");
			const dateB = new Date(b.Posters[0]?.Modified || "");
			if (sortOption === "dateUpdated") {
				return sortOrder === "asc" ? dateA.getTime() - dateB.getTime() : dateB.getTime() - dateA.getTime();
			}

			return dateB.getTime() - dateA.getTime();
		});
		log("INFO", "Collection Item Page", "Filter", "Filtered Collection Item Sets", { filtered });
		setFilteredCollectionItemSets(filtered);
	}, [
		hasError,
		responseLoading,
		collectionItem,
		collectionItemSets,
		sortOption,
		sortOrder,
		showHiddenUsers,
		userHides,
		userFollows,
	]);

	// 4. Compute hiddenCount based on filtering
	const hiddenCount = useMemo(() => {
		if (!collectionItemSets || collectionItemSets.length === 0) return 0;
		if (!userHides || userHides.length === 0) return 0;
		const uniqueHiddenUsers = new Set<string>();
		collectionItemSets.forEach((set) => {
			const isHidden = userHides.some((hide) => hide.Username === set.User.Name);
			if (isHidden) {
				uniqueHiddenUsers.add(set.User.Name);
			}
		});
		return uniqueHiddenUsers.size;
	}, [collectionItemSets, userHides]);

	// 5. Compute adjacent items when collectionItem changes
	useEffect(() => {
		if (!collectionItem) return;
		if (!collectionItem?.RatingKey) return;
		setNextCollectionItem(getAdjacentCollectionItem(collectionItem.RatingKey, "next"));
		setPreviousCollectionItem(getAdjacentCollectionItem(collectionItem.RatingKey, "previous"));
	}, [getAdjacentCollectionItem, collectionItem, setNextCollectionItem, setPreviousCollectionItem]);

	const handleShowHiddenUsers = () => {
		setShowHiddenUsers(!showHiddenUsers);
	};

	// Calculate number of active filters
	const numberOfActiveFilters = useMemo(() => {
		let count = 0;
		if (!showHiddenUsers) count++;

		return count;
	}, [showHiddenUsers]);

	if (!partialCollectionItem && !collectionItem && hasError) {
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

	if (!collectionItem && hasError) {
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
				backdropURL={`/api/mediaserver/image?ratingKey=${collectionItem?.RatingKey}&imageType=backdrop&cb=${imageVersion}`}
			/>

			<div className="p-4 lg:p-6">
				<div className="pb-6">
					{/* Header */}
					<CollectionItemDetails collectionItem={collectionItem || partialCollectionItem!} />

					{/* Loading and Error States */}
					{isLoading && (
						<div className={cn("mt-4 flex flex-col items-center", hasError ? "hidden" : "block")}>
							<Loader message={loadingMessage} />
						</div>
					)}
					{hasError && error && <ErrorMessage error={error} />}

					{/* Render filtered poster sets */}
					{collectionItemSets && collectionItemSets.length > 0 && collectionItem && (
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
								<CollectionItemFilter
									numberOfActiveFilters={numberOfActiveFilters}
									hiddenCount={hiddenCount}
									showHiddenUsers={showHiddenUsers}
									handleShowHiddenUsers={handleShowHiddenUsers}
								/>

								{/* Right column: sort options */}
								<div className="flex items-center sm:justify-end sm:ml-4">
									<SortControl
										options={[
											{
												value: "dateUpdated",
												label: "Date Updated",
												ascIcon: <CalendarArrowUp />,
												descIcon: <CalendarArrowDown />,
												type: "date",
											},
											{
												value: "username",
												label: "User Name",
												ascIcon: <ArrowDownAZ />,
												descIcon: <ArrowDownZA />,
												type: "string",
											},
										]}
										sortOption={sortOption}
										sortOrder={sortOrder}
										setSortOption={setSortOption}
										setSortOrder={setSortOrder}
										showLabel={false}
									/>
								</div>
							</div>

							<div className="text-center mb-4">
								{filteredCollectionItemSets &&
								filteredCollectionItemSets.length !== collectionItemSets.length ? (
									<div className="flex items-center justify-center gap-2 text-sm text-muted-foreground">
										<span>
											Showing {filteredCollectionItemSets.length} of {collectionItemSets.length}{" "}
											Collection Set
											{collectionItemSets.length > 1 ? "s" : ""}
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
											</ul>
											<p>You can adjust your filters using the checkboxes on this page.</p>
										</PopoverHelp>
									</div>
								) : (
									<p className="text-sm text-muted-foreground">
										{collectionItemSets.length} Collection Set
										{collectionItemSets.length > 1 ? "s" : ""}
									</p>
								)}
							</div>

							{filteredCollectionItemSets &&
								filteredCollectionItemSets.length === 0 &&
								collectionItemSets.length > 0 && (
									<div className="flex flex-col items-center">
										<ErrorMessage
											error={ReturnErrorMessage<string>(
												"All sets are hidden. Check your filters or hidden users."
											)}
										/>
										{!showHiddenUsers && (
											<Button
												className="mt-4"
												variant="secondary"
												onClick={handleShowHiddenUsers}
											>
												Show Hidden Users
											</Button>
										)}
									</div>
								)}

							<div className="divide-y divide-primary-dynamic/20 space-y-6">
								{(filteredCollectionItemSets ?? []).map((set) => (
									<div key={set.ID} className="pb-6">
										<Carousel
											opts={{
												align: "start",
												dragFree: true,
												slidesToScroll: "auto",
											}}
											className="w-full"
										>
											<div className="flex flex-col">
												<div className="flex flex-row items-center justify-between mb-1">
													<P className="text-primary-dynamic text-md font-semibold ml-1 w-3/4">
														{set.Title}
													</P>
													<div
														className={cn(
															"ml-auto flex space-x-2",
															set.Title.length > 29 && "mb-5 xs:mb-0"
														)}
													>
														<CollectionsDownloadModal
															collectionItem={collectionItem}
															collectionItemSet={set}
														/>
													</div>
												</div>
												<div className="text-md text-muted-foreground mb-1 flex items-center">
													<User />
													<Link
														href={`/user/${set.User.Name}`}
														className="hover:text-primary cursor-pointer underline"
													>
														{set.User.Name}
													</Link>
												</div>
												<Lead className="text-sm text-muted-foreground flex items-center mb-1 ml-1">
													Last Update:{" "}
													{formatLastUpdatedDate(
														set.Posters[0]?.Modified || "",
														set.Backdrops[0]?.Modified || ""
													)}
												</Lead>
											</div>
											<CarouselContent>
												<CarouselItem key={`${set.ID}`}>
													{set.Posters.length > 0 && set.Posters[0] && set.Posters[0].ID && (
														<AssetImage
															image={set.Posters[0]}
															aspect="poster"
															className={`w-full`}
														/>
													)}
													{set.Backdrops.length > 0 &&
														set.Backdrops[0] &&
														set.Backdrops[0].ID && (
															<AssetImage
																image={set.Backdrops[0]}
																aspect="backdrop"
																className={`w-full`}
															/>
														)}
												</CarouselItem>
											</CarouselContent>
											<CarouselNext className="right-2 bottom-0" />
											<CarouselPrevious className="right-8 bottom-0" />
										</Carousel>
									</div>
								))}
							</div>
						</>
					)}
				</div>
			</div>
		</>
	);
}
