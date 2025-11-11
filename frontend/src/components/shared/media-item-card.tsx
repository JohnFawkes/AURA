"use client";

import { CollectionItem } from "@/app/collections/page";
import { Database } from "lucide-react";

import React from "react";

import { useRouter } from "next/navigation";

import { AssetImage } from "@/components/shared/asset-image";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

import { useCollectionStore } from "@/lib/stores/global-store-collection-store";
import { useMediaStore } from "@/lib/stores/global-store-media-store";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

interface HomeMediaItemCardProps {
	item: MediaItem | CollectionItem;
}

const HomeMediaItemCard: React.FC<HomeMediaItemCardProps> = ({ item }) => {
	const router = useRouter();

	const { setMediaItem } = useMediaStore();
	const { setCollectionItem } = useCollectionStore();

	// Helper type guards
	function isMediaItem(item: any): item is MediaItem {
		return "ExistInDatabase" in item || "TMDB_ID" in item;
	}

	function isCollectionItem(item: any): item is CollectionItem {
		return "ChildCount" in item && "MediaItems" in item;
	}

	const handleMediaItemCardClick = (mediaItem: MediaItem) => {
		setMediaItem(mediaItem);
		//router.push(formatMediaItemUrl(mediaItem));
		router.push("/media-item/");
	};

	const handleCollectionItemCardClick = (collectionItem: CollectionItem) => {
		setCollectionItem(collectionItem);
		router.push("/collection-item/");
	};

	return (
		<Card
			key={item.RatingKey}
			className="relative items-center cursor-pointer hover:shadow-xl transition-shadow"
			onClick={() => {
				if (isMediaItem(item)) {
					handleMediaItemCardClick(item);
				} else if (isCollectionItem(item)) {
					handleCollectionItemCardClick(item);
				}
			}}
		>
			{isMediaItem(item) && item.ExistInDatabase && (
				<div className="absolute top-2 left-2 z-10">
					<Database className="text-green-500" size={20} />
				</div>
			)}

			{/* Poster Image */}
			<AssetImage image={item} className="w-[80%] h-auto transition-transform hover:scale-105" />

			{/* Title */}
			<span className="text-center text-lg text-foreground font-semibold mb-0">
				{item.Title.length > 55 ? `${item.Title.slice(0, 55)}...` : item.Title}
			</span>

			{/* Badges */}
			<CardContent className="flex flex-col md:flex-row justify-center items-center gap-1 p-1">
				<Badge variant="default" className="text-xs">
					{isMediaItem(item) ? item.Year : `${item.ChildCount} items`}
				</Badge>
				<Badge variant="default" className="text-xs">
					{item.LibraryTitle}
				</Badge>
			</CardContent>
		</Card>
	);
};

export default HomeMediaItemCard;
