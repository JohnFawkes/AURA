"use client";

import { useState } from "react";

import Image from "next/image";

import { Skeleton } from "@/components/ui/skeleton";

import { cn } from "@/lib/cn";
import { type AspectRatio, getAspectRatioClass, getImageSizes } from "@/lib/image-sizes";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterFile } from "@/types/media-and-posters/poster-sets";

interface AssetImageProps {
	image: PosterFile | MediaItem | string;
	aspect?: AspectRatio;
	className?: string;
	imageClassName?: string;
	priority?: boolean;
}

export function AssetImage({ image, aspect = "poster", className, imageClassName, priority = false }: AssetImageProps) {
	const [imageLoaded, setImageLoaded] = useState(false);

	const imageContent = (
		<>
			{/* Skeleton Loader */}
			{!imageLoaded && (
				<Skeleton className={cn("absolute inset-0 rounded-md animate-pulse", getAspectRatioClass(aspect))} />
			)}

			<Image
				src={
					typeof image === "string"
						? image
						: "ID" in image && "Modified" in image
							? `/api/mediux/image?assetID=${image.ID}&modifiedDate=${image.Modified}`
							: "RatingKey" in image
								? `/api/mediaserver/image?ratingKey=${image.RatingKey}&imageType=${aspect}`
								: ""
				}
				alt={typeof image === "string" ? `${image} ${aspect}` : "ID" in image ? image.ID : ""}
				fill
				sizes={getImageSizes(aspect)}
				className={cn(
					"transition-opacity duration-500",
					aspect === "logo" ? "object-contain" : "object-cover",
					imageLoaded ? "opacity-100" : "opacity-0",
					imageClassName
				)}
				priority={priority}
				onLoad={() => setImageLoaded(true)}
				unoptimized
				loading="lazy"
				draggable={false}
				style={{ userSelect: "none" }}
			/>
		</>
	);

	return (
		<>
			<div className={cn("relative flex flex-col", className)}>
				<div
					className={cn(
						"relative overflow-hidden rounded-md border border-primary-dynamic/40 hover:border-primary-dynamic transition-all duration-300 group",
						getAspectRatioClass(aspect)
					)}
				>
					{imageContent}
				</div>
			</div>
		</>
	);
}
