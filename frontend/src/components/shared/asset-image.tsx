"use client";

import { CollectionItem } from "@/app/collections/page";
import { decode } from "blurhash";

import { useMemo, useState } from "react";

import Image from "next/image";

import { Skeleton } from "@/components/ui/skeleton";

import { cn } from "@/lib/cn";
import { type AspectRatio, getAspectRatioClass, getImageSizes } from "@/lib/image-sizes";
import { log } from "@/lib/logger";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterFile } from "@/types/media-and-posters/poster-sets";

interface AssetImageProps {
	image: PosterFile | MediaItem | CollectionItem | string;
	aspect?: AspectRatio;
	className?: string;
	imageClassName?: string;
	priority?: boolean;
}

/**
 * Decodes a blurhash string to a data URL using canvas
 * @param blurhash The blurhash string (3x3 components, ~15-20 chars)
 * @returns Data URL string ready for blurDataURL prop, or undefined on error
 */
function decodeBlurhashToDataURL(blurhash: string): string | undefined {
	try {
		// Debug: Warn if blurhash is longer than expected for 3x3 components (~15-20 chars)
		if (blurhash.length > 30) {
			log(
				"WARN",
				"AssetImage",
				"decodeBlurhashToDataURL",
				`Large blurhash detected (${blurhash.length} chars). Expected 3x3 components (~15-20 chars).`
			);
		}

		// Decode blurhash to pixel data (2x2 for absolute minimal size)
		// For blur placeholders, very small dimensions are sufficient
		const width = 2;
		const height = 2;
		const pixels = decode(blurhash, width, height);

		// Create canvas and draw pixels
		const canvas = document.createElement("canvas");
		canvas.width = width;
		canvas.height = height;
		const ctx = canvas.getContext("2d");
		if (!ctx) {
			throw new Error("Failed to get canvas context");
		}

		// Create ImageData from pixels and draw to canvas
		const imageData = ctx.createImageData(width, height);
		imageData.data.set(pixels);
		ctx.putImageData(imageData, 0, 0);

		// PNG is typically smaller than JPEG for very small images
		return canvas.toDataURL("image/png");
	} catch (error) {
		log("ERROR", "AssetImage", "decodeBlurhashToDataURL", "Failed to decode blurhash", error);
		return undefined;
	}
}

export function AssetImage({ image, aspect = "poster", className, imageClassName, priority = false }: AssetImageProps) {
	const [imageLoaded, setImageLoaded] = useState(false);
	const [imageError, setImageError] = useState(false);

	// Decode blurhash string to data URL client-side
	const blurDataURL = useMemo(() => {
		const blurhash = typeof image === "object" && "Blurhash" in image ? image.Blurhash : undefined;
		if (!blurhash) return undefined;
		return decodeBlurhashToDataURL(blurhash);
	}, [image]);

	const imageSrc =
		typeof image === "string"
			? image
			: "ID" in image && "Modified" in image
				? `/api/mediux/image?assetID=${image.ID}&modifiedDate=${image.Modified}`
				: "RatingKey" in image
					? `/api/mediaserver/image?ratingKey=${image.RatingKey}&imageType=${aspect}`
					: "";

	const imageContent = (
		<>
			{!imageError ? (
				<Image
					src={imageSrc}
					alt={typeof image === "string" ? `${image} ${aspect}` : "ID" in image ? image.ID : ""}
					fill
					sizes={getImageSizes(aspect)}
					className={cn(
						"object-cover",
						"border border-transparent hover:border-primary-dynamic/30",
						"rounded-sm",
						"transition-all duration-300",
						imageClassName
					)}
					unoptimized
					loading="lazy"
					draggable={false}
					style={{ userSelect: "none" }}
					priority={priority}
					placeholder={blurDataURL ? "blur" : undefined}
					blurDataURL={blurDataURL}
					onLoad={() => setImageLoaded(true)}
					onError={() => setImageError(true)}
				/>
			) : (
				<div
					className={cn(
						"flex items-center justify-center w-full h-full bg-muted text-muted-foreground",
						getAspectRatioClass(aspect)
					)}
				>
					<div className="flex flex-col items-center">
						<span className="text-xs">No Image Available</span>
						<Image
							src="/aura_logo.svg"
							alt="Aura Logo"
							width={40}
							height={40}
							className="mt-1 opacity-70"
							draggable={false}
						/>
					</div>

					<Skeleton
						className={cn("absolute inset-0 rounded-md animate-pulse", getAspectRatioClass(aspect))}
					/>
				</div>
			)}
			{/* Overlay that fades out when image loads, revealing the sharp image underneath */}
			{blurDataURL && !imageLoaded && !imageError && (
				<Skeleton className={cn("absolute inset-0 rounded-md animate-pulse", getAspectRatioClass(aspect))} />
			)}
		</>
	);

	return (
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
	);
}
