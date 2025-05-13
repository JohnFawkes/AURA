"use client";

import Image from "next/image";
import { cn } from "@/lib/utils";
import { DynamicOverlay } from "@/components/ui/backgrounds/dynamic-overlay";
import { useEffect, useState } from "react";
import { setCssVariable } from "@/lib/utils";
import { MediaItem } from "@/types/mediaItem";
import { PosterSet } from "@/types/posterSets";

interface BackgroundProps {
	src?: string | null;
	alt?: string;
	priority?: boolean;
	quality?: number;
	className?: string;
	mediaItem: MediaItem;
	posterSet?: PosterSet;
	dynamicPalette?: {
		dynamicLeft: string | null;
		dynamicBottom: string | null;
		vibrant: {
			Vibrant?: string;
			LightVibrant?: string;
			DarkVibrant?: string;
			Muted?: string;
			LightMuted?: string;
			DarkMuted?: string;
		};
	};
}

export function ContentBackground({
	src,
	alt = "Background image",
	priority = true,
	quality = 90,
	className,
	mediaItem,
	posterSet,
	dynamicPalette,
}: BackgroundProps) {
	// Helper function to get backdrop from set with fallback
	const getSetBackdrop = () => {
		if (posterSet) {
			// First try show set backdrop
			if (
				posterSet.Files.filter((file) => file.Type === "backdrop")
					.length > 0
			) {
				return `/api/mediaserver/image/${
					posterSet.Files.find((file) => file.Type === "backdrop")?.ID
				}?modifiedDate=${
					posterSet.Files.find((file) => file.Type === "backdrop")
						?.Modified
				}`;
			}
			// Fallback to show backdrop
			if (mediaItem.RatingKey) {
				return `/api/mediaserver/image/${mediaItem.RatingKey}/backdrop`;
			}
		}

		return null;
	};

	const imageUrl = src || getSetBackdrop();
	const [isBlurred, setIsBlurred] = useState(false);

	// Set dynamic CSS variables
	useEffect(() => {
		if (dynamicPalette?.vibrant.LightMuted) {
			setCssVariable(
				"--primary-dynamic",
				dynamicPalette.vibrant.LightMuted
			);
		}
		if (dynamicPalette?.dynamicLeft) {
			setCssVariable("--dynamic-left", dynamicPalette.dynamicLeft);
		}
		if (dynamicPalette?.dynamicBottom) {
			setCssVariable("--dynamic-bottom", dynamicPalette.dynamicBottom);
		}
		if (dynamicPalette?.vibrant.DarkMuted) {
			setCssVariable(
				"--dynamic-dark-muted",
				dynamicPalette.vibrant.DarkMuted
			);
		}

		// Clean up when component unmounts
		return () => {
			setCssVariable("--primary-dynamic", "");
			setCssVariable("--dynamic-left", "");
			setCssVariable("--dynamic-bottom", "");
			setCssVariable("--dynamic-dark-muted", "");
		};
	}, [
		dynamicPalette?.vibrant.LightMuted,
		dynamicPalette?.dynamicLeft,
		dynamicPalette?.dynamicBottom,
		dynamicPalette?.vibrant.DarkMuted,
	]);

	// Handle scroll event to blur the background
	useEffect(() => {
		const handleScroll = () => {
			// Check if the user has scrolled down 300px (adjust as needed)
			if (window.scrollY > 300) {
				setIsBlurred(true);
			} else {
				setIsBlurred(false);
			}
		};

		// Add scroll event listener
		window.addEventListener("scroll", handleScroll);

		// Cleanup event listener on component unmount
		return () => {
			window.removeEventListener("scroll", handleScroll);
		};
	}, []);

	if (!imageUrl) return null;

	// Define the URL for the external mask
	const maskUrl = "/gradient.svg";

	return (
		<div
			className={cn(
				"fixed inset-0 -z-20 overflow-hidden w-full h-full",
				className,
				isBlurred && "blur-lg"
			)}
		>
			{/* Background Gradient */}
			<div className="absolute inset-0 bg-background">
				{/* Subtle Noise Texture */}
				<div className="absolute inset-0 opacity-[0.015] mix-blend-overlay">
					<div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiB4bWxucz0iaHR0cDovL3d3dy53My5org/2000/svgIj48ZmlsdGVyIGlkPSJhIiB4PSIwIiB5PSIwIj48ZmVUdXJidWxlbmNlIHR5cGU9ImZyYWN0YWxOb2lzZSIgYmFzZUZyZXF1ZW5jeT0iLjc1IiBzdGl0Y2hUaWxlcz0ic3RpdGNoIi8+PGZlQ29sb3JNYXRyaXggdHlwZT0ic2F0dXJhdGUiIHZhbHVlcz0iMCIvPjwvZmlsdGVyPjxyZWN0IHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiIGZpbHRlcj0idXJsKCNhKSIvPjwvc3ZnPg==')]" />
				</div>

				{/* Dynamic Overlay */}
				<DynamicOverlay className="absolute inset-0" />
			</div>

			{/* Image Container - Positioned Top Right, width driven, height by aspect ratio */}
			<div className="absolute top-0 right-0 w-full lg:w-[70vw] aspect-[16/9] z-50">
				<div className="relative w-full h-full">
					<Image
						src={`/api/mediaserver/image/${mediaItem.RatingKey}/backdrop}`}
						alt={alt}
						quality={quality}
						priority={priority}
						fill
						unoptimized
						loading="lazy"
						className="object-cover object-right-top"
						style={{
							maskImage: `url(${maskUrl})`,
							WebkitMaskImage: `url(${maskUrl})`,
							maskSize: "100% 100%",
							WebkitMaskSize: "100% 100%",
							maskRepeat: "no-repeat",
							WebkitMaskRepeat: "no-repeat",
							maskPosition: "center",
							WebkitMaskPosition: "center",
						}}
					/>
				</div>
			</div>
		</div>
	);
}
