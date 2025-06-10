"use client";

import { formatMediaItemUrl } from "@/helper/formatMediaItemURL";
import { CheckCircle2 as Checkmark } from "lucide-react";

import React from "react";

import { useRouter } from "next/navigation";

import { AssetImage } from "@/components/shared/asset-image";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

import { useMediaStore } from "@/lib/mediaStore";

import { MediaItem } from "@/types/mediaItem";

import { H4 } from "../../../components/ui/typography";

interface HomeMediaItemCardProps {
	mediaItem: MediaItem;
}

const HomeMediaItemCard: React.FC<HomeMediaItemCardProps> = ({ mediaItem }) => {
	const router = useRouter();

	const { setMediaItem } = useMediaStore();

	const handleCardClick = (mediaItem: MediaItem) => {
		setMediaItem(mediaItem);
		router.push(formatMediaItemUrl(mediaItem));
	};

	return (
		<Card
			key={mediaItem.RatingKey}
			className="relative items-center cursor-pointer hover:shadow-xl transition-shadow"
			style={{ backgroundColor: "var(--card)" }}
			onClick={() => handleCardClick(mediaItem)}
		>
			{mediaItem.ExistInDatabase && (
				<div className="absolute top-2 left-2 z-10">
					<Checkmark className="text-green-500" size={20} />
				</div>
			)}

			{/* Poster Image */}
			<AssetImage
				image={mediaItem}
				className="w-[170px] h-auto transition-transform hover:scale-105"
			/>

			{/* Title */}
			<H4 className="text-center font-semibold mb-2 px-2">
				{mediaItem.Title.length > 55
					? `${mediaItem.Title.slice(0, 55)}...`
					: mediaItem.Title}
			</H4>

			{/* Badges */}
			<CardContent className="flex justify-center gap-2">
				<Badge variant="default" className="text-xs">
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
