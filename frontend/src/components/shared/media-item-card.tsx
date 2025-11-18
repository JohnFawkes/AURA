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
			className="relative items-center cursor-pointer border border-1 hover:shadow-xl transition-shadow p-0 rounded-xl"
			onClick={() => {
				if (isMediaItem(item)) {
					handleMediaItemCardClick(item);
				} else if (isCollectionItem(item)) {
					handleCollectionItemCardClick(item);
				}
			}}
		>
			{/* Database Existence Indicator */}
			{isMediaItem(item) && item.ExistInDatabase && (
				<div className="absolute top-1 right-1 z-10 rounded-full p-1 border border-green-800">
					<Database className="text-green-500" size={20} />
				</div>
			)}

			{/* Poster Image */}
			<AssetImage image={item} className="w-[100%] h-auto transition-transform hover:scale-102 rounded-xl mb-0" />

			{/* Badges */}
			<CardContent className="flex flex-col justify-center items-center mt-0">
				<div className="flex flex-row gap-2">
					<Badge variant="default" className="text-xs">
						{isMediaItem(item) ? item.Year : `${item.ChildCount} items`}
					</Badge>
					<Badge variant="default" className="text-xs">
						{item.LibraryTitle}
					</Badge>
				</div>
				{/* Title */}
				<span className="text-center text-md text-foreground font-semibold mt-2 mb-2">
					{item.Title.length > 55 ? `${item.Title.slice(0, 55)}...` : item.Title}
				</span>
			</CardContent>
		</Card>
	);
};

export default HomeMediaItemCard;
