import React, { useContext } from "react";
import { useRouter } from "next/navigation";
import { Card, CardContent } from "@/components/ui/card";
import Image from "next/image";
import { Badge } from "@/components/ui/badge";
import { usePosterMediaStore } from "@/lib/setStore";
import { MediaItem } from "@/types/mediaItem";
import { H4 } from "./typography";
import { SearchContext } from "@/app/layout";

interface HomeMediaItemCardProps {
	mediaItem: MediaItem;
}

const HomeMediaItemCard: React.FC<HomeMediaItemCardProps> = ({ mediaItem }) => {
	const router = useRouter();
	const { setSearchQuery } = useContext(SearchContext);

	const setMediaItem = usePosterMediaStore((state) => state.setMediaItem);

	const handleCardClick = (mediaItem: MediaItem) => {
		// Store the mediaItem in Zustand
		setMediaItem(mediaItem);

		// Clear the search query
		setSearchQuery("");

		// Replace space with underscore for URL compatibility
		const formattedTitle = mediaItem.Title.replace(/\s+/g, "_");
		// Replace special characters with empty string
		const sanitizedTitle = formattedTitle.replace(/[^a-zA-Z0-9_]/g, "");
		router.push(`/media/${mediaItem.RatingKey}/${sanitizedTitle}`);
	};

	return (
		<Card
			key={mediaItem.RatingKey}
			className="items-center cursor-pointer hover:shadow-xl transition-shadow"
			style={{
				backgroundColor: "var(--card)",
			}}
			onClick={() => handleCardClick(mediaItem)}
		>
			{/* Poster Image */}
			<div className="relative w-[150px] h-[220px] rounded-md overflow-hidden transform transition-transform duration-300 hover:scale-105">
				<Image
					src={`/api/mediaserver/image/${mediaItem.RatingKey}/poster`}
					alt={mediaItem.Title}
					fill
					className="object-cover"
					loading="lazy"
					unoptimized
				/>
			</div>

			<H4 className="text-center font-semibold mb-2 px-2">
				{mediaItem.Title.length > 45
					? `${mediaItem.Title.slice(0, 45)}...`
					: mediaItem.Title}
			</H4>
			{/* Badges */}
			<CardContent className="flex justify-center gap-2">
				<Badge variant="default" className=" text-xs">
					{mediaItem.Year}
				</Badge>
				<Badge variant="default" className="text-xs">
					{mediaItem.LibraryTitle}
				</Badge>
			</CardContent>
		</Card>
	);
};

export default HomeMediaItemCard;
