import { useEffect, useState } from "react";
import { Badge } from "./ui/badge";
import { H1, Lead } from "./ui/typography";
import { MediaItemRatings } from "./media_item_ratings";
import { Guid } from "@/types/mediaItem";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "@/components/ui/accordion";
import { fetchMediaServerType } from "@/services/api.mediaserver";
import { log } from "@/lib/logger";

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
};

export function MediaItemDetails({
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
	status,
}: MediaItemDetailsProps) {
	const [serverType, setServerType] = useState<string>("");

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
				log(
					"Media Item Details Section - Error fetching server type:",
					error
				);
				setServerType("Media Server");
			}
		};
		fetchServerType();
	}, []);

	return (
		<div>
			{/* Title and Summary */}
			<div className="flex flex-col pt-40 justify-end items-center text-center lg:items-start lg:text-left">
				<H1 className="mb-1">{title}</H1>
				{/* Hide summary on mobile */}
				<Lead className="text-primary-dynamic max-w-4xl hidden md:block">
					{summary}
				</Lead>
			</div>

			{/* Year, Content Rating And External Ratings/Links */}
			<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-4">
				{/* Year */}
				{year && (
					<Badge className="flex items-center text-sm">{year}</Badge>
				)}

				{/* Content Rating */}
				{contentRating && (
					<Badge className="flex items-center text-sm">
						{contentRating}
					</Badge>
				)}

				{/* Status */}
				{status && (
					<Badge
						className={`flex items-center text-sm ${
							status.toLowerCase() === "ended"
								? "bg-red-700 text-white"
								: status.toLowerCase().startsWith("returning")
								? "bg-green-700 text-white"
								: ""
						}`}
					>
						{status.toLowerCase().startsWith("returning")
							? "Continuing"
							: status}
					</Badge>
				)}

				{/* External Ratings/Links from GUIDs */}
				<MediaItemRatings guids={guids} mediaItemType={mediaItemType} />
			</div>

			{/* Show Information for TV Shows */}
			{mediaItemType === "show" && seasonCount && episodeCount > 0 && (
				<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-4">
					<Lead className="flex items-center text-md text-primary-dynamic">
						{seasonCount} {seasonCount > 1 ? "Seasons" : "Season"}{" "}
						with {episodeCount}{" "}
						{episodeCount > 1 ? "Episodes" : "Episode"} in{" "}
						{serverType}
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
											<span className="font-semibold">
												File Path:
											</span>{" "}
											{moviePath}
										</p>
									)}

									{/* Show the Movie File Size */}
									{movieSize && (
										<p>
											<span className="font-semibold">
												File Size:
											</span>{" "}
											{movieSize >= 1024 * 1024 * 1024
												? `${(
														movieSize /
														(1024 * 1024 * 1024)
												  ).toFixed(2)} GB`
												: `${(
														movieSize /
														(1024 * 1024)
												  ).toFixed(2)} MB`}
										</p>
									)}

									{/* Show the Movie Duration */}
									{movieDuration && (
										<p>
											<span className="font-semibold">
												Duration:
											</span>{" "}
											{movieDuration < 3600000
												? `${Math.floor(
														movieDuration / 60000
												  )} minutes`
												: `${Math.floor(
														movieDuration / 3600000
												  )}hr ${Math.floor(
														(movieDuration %
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
		</div>
	);
}
