"use client";

import Image from "next/image";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { FILE_TYPES } from "@/types";
import { PosterFile } from "@/types/posterSets";

interface AssetImageProps {
	image: PosterFile;
	displayUser?: boolean;
	displayMediaType?: boolean;
	displaySetLink?: boolean;
	openInDialog?: boolean;
	className?: string;
	imageClassName?: string;
	priority?: boolean;
	aspect?: "poster" | "backdrop" | "titlecard" | "album" | "logo";
	setId?: string;
}

export function AssetImage({
	image,
	className,
	imageClassName,
	priority = false,
	aspect = "poster",
}: AssetImageProps) {
	const [imageLoaded, setImageLoaded] = useState(false);

	// Determine aspect ratio based on aspect prop
	const getAspectRatioClass = () => {
		switch (aspect) {
			case "poster":
				return "aspect-[2/3]";
			case "backdrop":
			case "titlecard":
			case "logo":
				return "aspect-video";
			case "album":
				return "aspect-square";
			default:
				return "aspect-[2/3]";
		}
	};

	// Determine image sizes based on file type
	const sizes =
		image.Type === FILE_TYPES.BACKDROP ||
		image.Type === FILE_TYPES.TITLECARD
			? "(max-width: 640px) 50vw, 300px"
			: "300px";

	const imageContent = (
		<Image
			src={`/api/mediux/image/${image.ID}?modifiedDate=${image.Modified}`}
			alt={image.ID}
			fill
			sizes={sizes}
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
	);

	return (
		<>
			<div className={cn("relative flex flex-col", className)}>
				<div
					className={cn(
						"relative overflow-hidden rounded-md border border-primary-dynamic/40 hover:border-primary-dynamic transition-all duration-300 group",
						getAspectRatioClass()
					)}
				>
					{imageContent}
				</div>
			</div>
		</>
	);
}
