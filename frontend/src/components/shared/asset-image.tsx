"use client";

import { useState } from "react";

import Image from "next/image";

import { cn } from "@/lib/utils";
import { type AspectRatio, getAspectRatioClass, getImageSizes } from "@/lib/utils/image-utils";

import { MediaItem } from "@/types/mediaItem";
import { PosterFile } from "@/types/posterSets";

import { Skeleton } from "../ui/skeleton";

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
							? `/api/mediux/image/${image.ID}?modifiedDate=${image.Modified}`
							: "RatingKey" in image
								? `/api/mediaserver/image/${image.RatingKey}/${aspect}`
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
