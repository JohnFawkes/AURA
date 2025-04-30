"use client";

import ErrorMessage from "@/components/ui/error-message";
import { usePosterMediaStore } from "@/lib/setStore";
import { cn } from "@/lib/utils";
import Image from "next/image";
import { useEffect, useState } from "react";
import { H1, Lead } from "@/components/ui/typography";
import { Badge } from "@/components/ui/badge";
import { AssetImage } from "@/components/ui/asset-image";
import { PosterFile } from "@/types/posterSets";
import Link from "next/link";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "@/components/ui/accordion";
import PosterSetModal from "@/components/ui/poster-set-modal";

const SetPage = () => {
	const { posterSet, mediaItem } = usePosterMediaStore();
	const [isBlurred, setIsBlurred] = useState(false);
	const [backdropURL, setBackdropURL] = useState("");

	// Check if posterSet and mediaItem are defined
	if (!posterSet || !mediaItem) {
		return <ErrorMessage message="Poster set or media item not found." />;
	}

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

	// Construct the backdrop URL
	// If the posterSet.Files contains a backdrop, use that
	// Otherwise, use the mediaItem's backdrop
	useEffect(() => {
		if (posterSet.Files && posterSet.Files.length > 0) {
			const backdropFile = posterSet.Files.find(
				(file) => file.Type === "backdrop"
			);
			if (backdropFile) {
				setBackdropURL(
					`/api/mediux/image/${backdropFile.ID}?modifiedDate=${backdropFile.Modified}`
				);
			} else {
				setBackdropURL(
					`/api/mediaserver/image/${mediaItem.RatingKey}/backdrop`
				);
			}
		}
	}, [mediaItem.RatingKey, posterSet.Files]);

	// Helper function to get the count and label for each type
	const getFileTypeCount = (type: string, label: string) => {
		const count = posterSet.Files.filter(
			(file) => file.Type === type
		).length;
		return count > 0 ? `${count} ${label}${count > 1 ? "s" : ""}` : null;
	};

	// Get counts for each type
	const posterCount = getFileTypeCount("poster", "Poster");
	const backdropCount = getFileTypeCount("backdrop", "Backdrop");
	const seasonPosterCount = getFileTypeCount("seasonPoster", "Season Poster");
	const titlecardCount = getFileTypeCount("titlecard", "Titlecard");

	// Combine counts into a single string
	const fileCounts = [
		posterCount,
		backdropCount,
		seasonPosterCount,
		titlecardCount,
	]
		.filter(Boolean) // Remove null values
		.join(" • ");

	return (
		<div>
			<div
				className={cn(
					"fixed inset-0 -z-20 overflow-hidden w-full h-full transition-all duration-1000",
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
					<div
						className="absolute inset-0"
						style={
							{
								background: `
						radial-gradient(ellipse at 30% 30%, var(--dynamic-left) 0%, transparent 60%),
						radial-gradient(ellipse at bottom right, var(--dynamic-bottom) 0%, transparent 60%),
						radial-gradient(ellipse at center, var(--dynamic-dark-muted) 0%, transparent 80%),
						var(--background)
					  `,
								opacity: 0.5,
							} as React.CSSProperties
						}
					/>
				</div>

				{/* Image Container - Positioned Top Right, width driven, height by aspect ratio */}
				<div className="absolute top-0 right-0 w-full lg:w-[70vw] aspect-[16/9] z-50">
					<div className="relative w-full h-full">
						<Image
							src={backdropURL}
							alt={"Backdrop"}
							fill
							priority
							unoptimized
							className="object-cover object-right-top"
							style={{
								maskImage: `url(/gradient.svg)`,
								WebkitMaskImage: `url(/gradient.svg)`,
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

			<div className="p-4 lg:p-6">
				<div className="pb-6">
					{/* Title */}
					<div className="flex flex-col pt-40 justify-end items-center text-center lg:items-start lg:text-left">
						<H1 className="mb-1">{mediaItem?.Title}</H1>
					</div>

					{/* Set Author */}
					<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center text-white gap-4 tracking-wide mt-4">
						{posterSet.User.Name && (
							<>
								<Badge
									className="flex items-center text-sm hover:text-white transition-colors"
									onClick={() => {
										window.open(
											`https://mediux.pro/sets/${posterSet.ID}`,
											"_blank"
										);
									}}
								>
									Set by: {posterSet.User.Name}
								</Badge>
								<Badge className="flex items-center text-sm">
									{fileCounts}
								</Badge>
								<div className="ml-auto">
									<PosterSetModal
										posterSet={posterSet}
										mediaItem={mediaItem}
									/>
								</div>
							</>
						)}
					</div>

					{/* Display all posters (if any) */}
					{posterSet.Files.some((file) => file.Type === "poster") && (
						<div className="flex flex-col gap-2 mt-4">
							<span className="text-muted-foreground text-sm">
								{posterSet.Files.filter(
									(file) => file.Type === "poster"
								).length > 1
									? "Posters"
									: "Poster"}
							</span>
							<div className="flex flex-wrap justify-center lg:justify-start items-center gap-4">
								{posterSet.Files.map((file) => {
									if (file.Type === "poster") {
										return (
											<AssetImage
												key={file.ID}
												image={
													file as unknown as PosterFile
												}
												displayUser={true}
												displayMediaType={true}
												aspect="poster"
												className="w-[200px] h-auto"
											/>
										);
									}
									return null;
								})}
							</div>
						</div>
					)}

					{/* Display all backdrops (if any) */}
					{posterSet.Files.some(
						(file) => file.Type === "backdrop"
					) && (
						<div className="flex flex-col gap-2 mt-4">
							<span className="text-muted-foreground text-sm">
								{posterSet.Files.filter(
									(file) => file.Type === "backdrop"
								).length > 1
									? "Backdrops"
									: "Backdrop"}
							</span>
							<div className="flex flex-wrap justify-center lg:justify-start items-center gap-4">
								{posterSet.Files.map((file) => {
									if (file.Type === "backdrop") {
										return (
											<AssetImage
												key={file.ID}
												image={
													file as unknown as PosterFile
												}
												displayUser={true}
												displayMediaType={true}
												aspect="backdrop"
												className="w-[200px] h-auto"
											/>
										);
									}
									return null;
								})}
							</div>
						</div>
					)}

					{/* Display all season posters (if any) */}
					{posterSet.Files.some(
						(file) => file.Type === "seasonPoster"
					) && (
						<div className="flex flex-col gap-2 mt-4">
							<span className="text-muted-foreground text-sm">
								{posterSet.Files.filter(
									(file) => file.Type === "seasonPoster"
								).length > 1
									? "Season Posters"
									: "Season Poster"}
							</span>
							<div className="flex flex-wrap justify-center lg:justify-start items-center gap-4">
								{posterSet.Files.map((file) => {
									if (file.Type === "seasonPoster") {
										return (
											<AssetImage
												key={file.ID}
												image={
													file as unknown as PosterFile
												}
												displayUser={true}
												displayMediaType={true}
												aspect="poster"
												className="w-[200px] h-auto"
											/>
										);
									}
									return null;
								})}
							</div>
						</div>
					)}

					{/* Display all titlecards (if any) */}
					{posterSet.Files.some(
						(file) => file.Type === "titlecard"
					) && (
						<div className="flex flex-col gap-2 mt-4">
							<span className="text-muted-foreground text-sm">
								Titlecards
							</span>
							<Accordion
								type="single"
								collapsible
								className="w-full"
							>
								{/* Group titlecards by season */}
								{Object.entries(
									posterSet.Files.filter(
										(file) => file.Type === "titlecard"
									).reduce((acc, file) => {
										const season =
											file.Episode?.SeasonNumber ||
											"Unknown Season"; // Default to "Unknown Season" if no season is specified
										if (!acc[season]) {
											acc[season] = [];
										}
										acc[season].push(file);
										return acc;
									}, {} as Record<string, PosterFile[]>)
								).map(([season, files]) => (
									<AccordionItem
										key={season}
										value={`season-${season}`}
									>
										<AccordionTrigger>
											Season {season}
										</AccordionTrigger>
										<AccordionContent>
											<div className="flex flex-wrap justify-center lg:justify-start items-center gap-4">
												{files.map((file) => (
													<AssetImage
														key={file.ID}
														image={
															file as unknown as PosterFile
														}
														displayUser={true}
														displayMediaType={true}
														aspect="titlecard"
														className="w-[200px] h-auto"
													/>
												))}
											</div>
										</AccordionContent>
									</AccordionItem>
								))}
							</Accordion>
						</div>
					)}
				</div>
			</div>
		</div>
	);
};

export default SetPage;
