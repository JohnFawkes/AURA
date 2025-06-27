"use client";

import { formatMediaItemUrl } from "@/helper/formatMediaItemURL";
import { fetchMediaServerType } from "@/services/api.mediaserver";
import { Database } from "lucide-react";

import { useEffect, useState } from "react";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { AssetImage } from "@/components/shared/asset-image";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";

import { useSearchQueryStore } from "@/lib/homeSearchStore";
import { log } from "@/lib/logger";
import { useMediaStore } from "@/lib/mediaStore";

import { Guid, MediaItem } from "@/types/mediaItem";

import { Badge } from "../ui/badge";
import { H1, Lead } from "../ui/typography";
import { MediaItemRatings } from "./media_item_ratings";

type MediaItemDetailsProps = {
	ratingKey: string;
	mediaItemType: string;
	title: string;
	summary: string;
	year: number;
	contentRating: string;
	seasonCount: number;
	episodeCount: number;
	moviePath: string;
	movieSize: number;
	movieDuration: number;
	guids: Guid[];
	existsInDB: boolean;
	status: string;
	libraryTitle: string;
	otherMediaItem: MediaItem | null;
};

export function MediaItemDetails({
	ratingKey,
	mediaItemType,
	title,
	summary,
	year,
	contentRating,
	seasonCount,
	episodeCount,
	moviePath,
	movieSize,
	movieDuration,
	guids,
	existsInDB,
	status,
	libraryTitle,
	otherMediaItem,
}: MediaItemDetailsProps) {
	const [serverType, setServerType] = useState<string>("");
	const router = useRouter();
	const { setMediaItem } = useMediaStore();
	const { setSearchQuery } = useSearchQueryStore();

	useEffect(() => {
		const fetchServerType = async () => {
			try {
				const response = await fetchMediaServerType();
				if (!response) {
					throw new Error("Failed to fetch server type");
				}
				const serverType = response.data?.serverType || "Media Server";
				setServerType(serverType);
			} catch (error) {
				log("Media Item Details Section - Error fetching server type:", error);
				setServerType("Media Server");
			}
		};
		fetchServerType();
	}, []);

	const handleSavedSetsPageClick = () => {
		if (existsInDB) {
			setSearchQuery(`${title} y:${year}: id:${ratingKey}:`);
			router.push("/saved-sets");
		}
	};

	return (
		<div>
			<div className="flex flex-col lg:flex-row pt-40 items-center lg:items-start text-center lg:text-left">
				{/* Poster Image */}
				{ratingKey && (
					<div className="flex-shrink-0 mb-4 lg:mb-0 lg:mr-8 flex justify-center">
						<AssetImage
							image={`/api/mediaserver/image/${ratingKey}/poster?cb=${Date.now()}`}
							className="w-[200px] h-auto transition-transform hover:scale-105"
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

						{/* Status */}
						{status && (
							<Badge
								className={`flex items-center text-sm ${
									status.toLowerCase() === "ended" ||
									status.toLowerCase() === "cancelled" ||
									status.toLowerCase() === "canceled"
										? "bg-red-700 text-white"
										: status.toLowerCase().startsWith("returning")
											? "bg-green-700 text-white"
											: ""
								}`}
							>
								{status.toLowerCase().startsWith("returning") ? "Continuing" : status}
							</Badge>
						)}
					</div>

					{/* External Ratings/Links from GUIDs */}
					{/* Year, Content Rating And External Ratings/Links */}
					<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-4">
						<MediaItemRatings guids={guids} mediaItemType={mediaItemType} title={title} />
					</div>
				</div>
			</div>

			{/* Library Information */}
			{libraryTitle && (
				<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-4">
					<Lead className="text-md text-primary-dynamic ml-1">
						<span className="font-semibold">{libraryTitle} Library</span>{" "}
					</Lead>
				</div>
			)}

			{/* Other Media Item Information */}
			{otherMediaItem && (
				<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-4">
					<Lead className="text-md text-primary-dynamic ml-1">
						Also available in{" "}
						<Link
							href={formatMediaItemUrl(otherMediaItem)}
							onClick={() => {
								setMediaItem(otherMediaItem);
								router.push(formatMediaItemUrl(otherMediaItem));
							}}
							className="text-primary-dynamic hover:underline"
						>
							{otherMediaItem.LibraryTitle}
						</Link>{" "}
						Library
					</Lead>
				</div>
			)}

			{/* Show Existence in Database */}
			<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-4">
				<Lead
					className={`text-md ${existsInDB ? "hover:cursor-pointer text-green-500" : "text-red-500"}`}
					onClick={() => {
						if (existsInDB) {
							handleSavedSetsPageClick();
						}
					}}
				>
					<Database className="inline ml-1 mr-1" /> {existsInDB ? "Already in Database" : "Not in Database"}
				</Lead>
			</div>

			{/* Season/Episode Information */}
			{mediaItemType === "show" && seasonCount > 0 && episodeCount > 0 && (
				<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-4">
					<Lead className="flex items-center text-md text-primary-dynamic ml-1">
						{seasonCount} {seasonCount > 1 ? "Seasons" : "Season"} with {episodeCount}{" "}
						{episodeCount > 1 ? "Episodes" : "Episode"} in {serverType}
					</Lead>
				</div>
			)}

			{/* Movie Information */}
			<div className="lg:flex items-center text-white gap-8 tracking-wide mt-4">
				{mediaItemType === "movie" && (
					<div className="flex flex-col w-full items-center">
						<Accordion type="single" collapsible className="w-full">
							<AccordionItem value="movie-details">
								<AccordionTrigger className="text-primary font-semibold">
									Movie Details
								</AccordionTrigger>
								<AccordionContent className="mt-2 text-sm text-muted-foreground space-y-2">
									{/* Show the Movie File Path */}
									{moviePath && (
										<p>
											<span className="font-semibold">File Path:</span> {moviePath}
										</p>
									)}

									{/* Show the Movie File Size */}
									{movieSize && (
										<p>
											<span className="font-semibold">File Size:</span>{" "}
											{movieSize >= 1024 * 1024 * 1024
												? `${(movieSize / (1024 * 1024 * 1024)).toFixed(2)} GB`
												: `${(movieSize / (1024 * 1024)).toFixed(2)} MB`}
										</p>
									)}

									{/* Show the Movie Duration */}
									{movieDuration && (
										<p>
											<span className="font-semibold">Duration:</span>{" "}
											{movieDuration < 3600000
												? `${Math.floor(movieDuration / 60000)} minutes`
												: `${Math.floor(movieDuration / 3600000)}hr ${Math.floor(
														(movieDuration % 3600000) / 60000
													)}min`}
										</p>
									)}
								</AccordionContent>
							</AccordionItem>
						</Accordion>
					</div>
				)}
			</div>
		</div>
	);
}
