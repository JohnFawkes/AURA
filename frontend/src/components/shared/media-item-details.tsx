"use client";

import { postAddItemToDB } from "@/services/database/api-db-item-add";
import { ChevronDown, Database, MoreHorizontal, RefreshCcw, Star, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { useEffect, useRef, useState } from "react";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { AssetImage } from "@/components/shared/asset-image";
import { MediaItemRatingModal } from "@/components/shared/media-item-rating-modal";
import { MediaItemRatings } from "@/components/shared/media-item-ratings";
import { RefreshMetadataModal } from "@/components/shared/media-item-refresh-modal";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { H1, Lead } from "@/components/ui/typography";

import { cn } from "@/lib/cn";
import { useMediaStore } from "@/lib/stores/global-store-media-store";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";

import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

type MediaItemDetailsProps = {
	mediaItem?: MediaItem;
	existsInDB: boolean;
	onExistsInDBChange?: (existsInDB: boolean) => void;
	status: string;
	otherMediaItem: MediaItem | null;
	posterImageKeys?: string[];
	serverType: string;
};

export function MediaItemDetails({
	mediaItem,
	existsInDB,
	onExistsInDBChange,
	status,
	otherMediaItem,
	posterImageKeys,
	serverType,
}: MediaItemDetailsProps) {
	const [isInDB, setIsInDBLocal] = useState(existsInDB);
	const router = useRouter();
	const { setMediaItem } = useMediaStore();
	const { setSearchQuery } = useSearchQueryStore();

	const [currentPosterIndex, setCurrentPosterIndex] = useState(0);
	const [isRefreshMetadataModalOpen, setIsRefreshMetadataModalOpen] = useState(false);
	const [isRatingModalOpen, setIsRatingModalOpen] = useState(false);

	const touchStartXRef = useRef<number | undefined>(undefined);
	const mouseStartXRef = useRef<number | undefined>(undefined);

	const title = mediaItem?.Title || "";
	const year = mediaItem?.Year || 0;
	const contentRating = mediaItem?.ContentRating || "";
	const mediaItemType = mediaItem?.Type || "";
	const summary = mediaItem?.Summary || "";
	const tmdbID = mediaItem?.TMDB_ID || "";
	const libraryTitle = mediaItem?.LibraryTitle || "";
	const seasonCount = mediaItemType === "show" ? mediaItem?.Series?.SeasonCount || 0 : 0;
	const episodeCount = mediaItemType === "show" ? mediaItem?.Series?.EpisodeCount || 0 : 0;
	const moviePath = mediaItem?.Movie?.File?.Path || "N/A";
	const movieSize = mediaItem?.Movie?.File?.Size || 0;
	const movieDuration = mediaItem?.Movie?.File?.Duration || 0;
	const guids = mediaItem?.Guids || [];

	// Sync when parent prop changes (e.g. route change or external update)
	useEffect(() => {
		setIsInDBLocal(existsInDB);
	}, [existsInDB]);

	const updateInDB = (next: boolean) => {
		setIsInDBLocal(next); // local optimistic
		onExistsInDBChange?.(next); // notify parent
	};

	const handleSavedSetsPageClick = () => {
		if (isInDB) {
			setSearchQuery(`${title} Y:${year}: ID:${tmdbID}: L:${libraryTitle}:`);
			router.push("/saved-sets");
		}
	};

	const handleAddToIgnoredClick = async () => {
		// Create a DBMediaItemWithPosterSets object to add to DB
		// with minimal required fields
		const ignoreDBItem: DBMediaItemWithPosterSets = {
			TMDB_ID: tmdbID,
			LibraryTitle: libraryTitle,
			MediaItem: mediaItem as MediaItem,
			PosterSets: [
				{
					PosterSetID: "ignore",
					PosterSet: {
						Title: "",
						ID: "ignore",
						Type: mediaItem?.Type as "movie" | "show",
						User: {
							Name: "",
						},
						DateCreated: new Date().toISOString(),
						DateUpdated: new Date().toISOString(),
						Status: "",
					},
					LastDownloaded: new Date().toISOString(),
					SelectedTypes: [],
					AutoDownload: false,
					ToDelete: false,
				},
			],
		};

		const addToDBResp = await postAddItemToDB(ignoreDBItem);
		if (addToDBResp.status === "error") {
			toast.error(`Failed to add ${title} to DB`);
			return;
		}

		updateInDB(true);
		toast.success(`Will successfully ignore ${title} in the future`);
	};

	return (
		<div>
			<div className="flex flex-col lg:flex-row pt-30 items-center lg:items-start text-center lg:text-left">
				{/* Poster Image */}
				{posterImageKeys && posterImageKeys.length > 0 && (
					<div
						className="flex-shrink-0 mb-4 lg:mb-0 lg:mr-8 flex justify-center"
						onDoubleClick={() => setCurrentPosterIndex(0)}
						onTouchStart={(e) => {
							touchStartXRef.current = e.touches[0].clientX;
						}}
						onTouchEnd={(e) => {
							const startX = touchStartXRef.current;
							const endX = e.changedTouches[0].clientX;
							if (startX !== undefined) {
								if (endX - startX > 50) {
									setCurrentPosterIndex(
										(prev) => (prev - 1 + posterImageKeys.length) % posterImageKeys.length
									);
								} else if (startX - endX > 50) {
									setCurrentPosterIndex((prev) => (prev + 1) % posterImageKeys.length);
								}
							}
							touchStartXRef.current = undefined;
						}}
						onMouseDown={(e) => {
							mouseStartXRef.current = e.clientX;
						}}
						onMouseUp={(e) => {
							const startX = mouseStartXRef.current;
							const endX = e.clientX;
							if (startX !== undefined) {
								if (endX - startX > 50) {
									setCurrentPosterIndex(
										(prev) => (prev - 1 + posterImageKeys.length) % posterImageKeys.length
									);
								} else if (startX - endX > 50) {
									setCurrentPosterIndex((prev) => (prev + 1) % posterImageKeys.length);
								}
							}
							mouseStartXRef.current = undefined;
						}}
					>
						<AssetImage
							image={`/api/mediaserver/image?ratingKey=${posterImageKeys[currentPosterIndex]}&imageType=poster&cb=${Date.now()}`}
							className="w-[200px] h-auto transition-transform hover:scale-105 select-none"
						/>
					</div>
				)}

				{/* Title and Summary */}
				<div className="flex flex-col items-center lg:items-start">
					<H1 className="mb-1">{title}</H1>
					{/* Hide summary on mobile */}
					<Lead className="text-primary-dynamic max-w-xl hidden lg:block">{summary}</Lead>

					{/* Year, Content Rating And External Ratings/Links */}
					<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-4">
						{/* Year */}
						{year && <Badge className="flex items-center text-sm">{year}</Badge>}

						{/* Content Rating */}
						{contentRating && <Badge className="flex items-center text-sm">{contentRating}</Badge>}

						{status ? (
							<Badge
								className={cn(
									"flex items-center text-sm",
									(status.toLowerCase() === "ended" ||
										status.toLowerCase() === "cancelled" ||
										status.toLowerCase() === "canceled") &&
										"bg-red-700 text-white",
									status.toLowerCase().startsWith("returning") && "bg-green-700 text-white"
								)}
							>
								{status.toLowerCase().startsWith("returning") ? "Continuing" : status}
							</Badge>
						) : null}
					</div>

					{/* External Ratings/Links from GUIDs */}
					<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-4">
						<MediaItemRatings guids={guids} mediaItemType={mediaItemType} title={title} />
					</div>
				</div>
			</div>

			{/* Refresh Metadata Modal */}
			{mediaItem && (
				<RefreshMetadataModal
					mediaItem={mediaItem}
					isOpen={isRefreshMetadataModalOpen}
					onClose={() => setIsRefreshMetadataModalOpen(false)}
				/>
			)}

			{/* Rating Modal */}
			{mediaItem && (
				<MediaItemRatingModal
					mediaItem={mediaItem}
					isOpen={isRatingModalOpen}
					onClose={() => setIsRatingModalOpen(false)}
				/>
			)}

			{/* Library Information */}
			{libraryTitle ? (
				<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-0 md:mt-2">
					<Lead className="text-md text-primary-dynamic ml-1">
						<span className="font-semibold">{libraryTitle} Library</span>{" "}
					</Lead>
				</div>
			) : null}

			{/* Other Media Item Information */}
			{otherMediaItem ? (
				<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-1">
					<Lead className="text-md text-primary-dynamic ml-1">
						Also available in{" "}
						<Link
							href={"/media-item/"}
							onClick={() => {
								setMediaItem(otherMediaItem);
								router.push("/media-item/");
							}}
							className="text-primary-dynamic hover:text-primary underline"
						>
							{otherMediaItem.LibraryTitle}
						</Link>{" "}
						Library
					</Lead>
				</div>
			) : null}

			{/* Show Existence in Database */}
			<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-2">
				<Lead
					className={cn("text-md", isInDB ? "hover:cursor-pointer text-green-500" : "text-red-500")}
					onClick={() => {
						if (isInDB) handleSavedSetsPageClick();
					}}
				>
					<Database className="inline ml-1 mr-1" /> {isInDB ? "Already in Database" : "Not in Database"}
				</Lead>
			</div>

			{/* Actions */}
			<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-2">
				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<Button
							variant="outline"
							size="sm"
							className={cn("gap-2 px-3", "border border-border shadow-sm", "hover:bg-secondary/80")}
						>
							Actions
							<ChevronDown className="h-4 w-4 opacity-80" />
						</Button>
					</DropdownMenuTrigger>

					<DropdownMenuContent align="start" side="bottom" className="min-w-[220px]">
						<DropdownMenuItem
							className="cursor-pointer"
							onSelect={() => setIsRefreshMetadataModalOpen(true)}
						>
							<RefreshCcw className="mr-2 h-4 w-4" />
							Refresh Metadata
						</DropdownMenuItem>

						<DropdownMenuItem className="cursor-pointer" onSelect={() => setIsRatingModalOpen(true)}>
							<Star className="mr-2 h-4 w-4" />
							Rate {mediaItemType === "movie" ? "Movie" : "Show"}
						</DropdownMenuItem>

						<DropdownMenuSeparator />

						{isInDB ? (
							<DropdownMenuItem
								className="cursor-pointer"
								onSelect={() => {
									setSearchQuery(`${title} Y:${year}: ID:${tmdbID}: L:${libraryTitle}:`);
									router.push("/saved-sets");
								}}
							>
								<Database className="mr-2 h-4 w-4" />
								View in Saved Sets
							</DropdownMenuItem>
						) : (
							<DropdownMenuItem
								className="cursor-pointer text-destructive focus:text-destructive"
								onSelect={() => {
									void handleAddToIgnoredClick();
								}}
							>
								<Trash2 className="mr-2 h-4 w-4" />
								Mark as Ignored
							</DropdownMenuItem>
						)}
					</DropdownMenuContent>
				</DropdownMenu>
			</div>

			{/* Season/Episode Information */}
			{mediaItemType === "show" && seasonCount > 0 && episodeCount > 0 ? (
				<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-2">
					<Lead className="flex items-center text-md text-primary-dynamic ml-1">
						{seasonCount} {seasonCount > 1 ? "Seasons" : "Season"} with {episodeCount}{" "}
						{episodeCount > 1 ? "Episodes" : "Episode"} in {serverType}
					</Lead>
				</div>
			) : null}

			{/* Movie Information */}
			<div className="lg:flex items-center text-white gap-8 tracking-wide mt-4">
				{mediaItemType === "movie" ? (
					<div className="flex flex-col w-full items-center">
						<Accordion type="single" collapsible className="w-full">
							<AccordionItem value="movie-details">
								<AccordionTrigger className="text-primary font-semibold">
									Movie Details
								</AccordionTrigger>
								<AccordionContent className="mt-2 text-sm text-muted-foreground space-y-2">
									{moviePath ? (
										<p>
											<span className="font-semibold">File Path:</span> {moviePath}
										</p>
									) : null}

									{movieSize ? (
										<p>
											<span className="font-semibold">File Size:</span>{" "}
											{movieSize >= 1024 * 1024 * 1024
												? `${(movieSize / (1024 * 1024 * 1024)).toFixed(2)} GB`
												: `${(movieSize / (1024 * 1024)).toFixed(2)} MB`}
										</p>
									) : null}

									{movieDuration ? (
										<p>
											<span className="font-semibold">Duration:</span>{" "}
											{movieDuration < 3600000
												? `${Math.floor(movieDuration / 60000)} minutes`
												: `${Math.floor(movieDuration / 3600000)}hr ${Math.floor(
														(movieDuration % 3600000) / 60000
													)}min`}
										</p>
									) : null}
								</AccordionContent>
							</AccordionItem>
						</Accordion>
					</div>
				) : null}
			</div>
		</div>
	);
}
