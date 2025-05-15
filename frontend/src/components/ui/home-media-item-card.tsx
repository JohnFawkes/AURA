import React from "react";
import { useRouter } from "next/navigation";
import { Card, CardContent } from "@/components/ui/card";
import Image from "next/image";
import { Badge } from "@/components/ui/badge";
import { MediaItem } from "@/types/mediaItem";
import { H4 } from "./typography";
import { CheckCircle2 as Checkmark } from "lucide-react";
import { useMediaStore } from "@/lib/mediaStore";

interface HomeMediaItemCardProps {
	mediaItem: MediaItem;
}

const HomeMediaItemCard: React.FC<HomeMediaItemCardProps> = ({ mediaItem }) => {
	const router = useRouter();

	const { setMediaItem } = useMediaStore();

	const handleCardClick = (mediaItem: MediaItem) => {
		setMediaItem(mediaItem);
		// Format title for URL (replace spaces with underscores, remove special characters)
		const formattedTitle = mediaItem.Title.replace(/\s+/g, "_");
		const sanitizedTitle = formattedTitle.replace(/[^a-zA-Z0-9_]/g, "");
		router.push(`/media/${mediaItem.RatingKey}/${sanitizedTitle}`);
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
