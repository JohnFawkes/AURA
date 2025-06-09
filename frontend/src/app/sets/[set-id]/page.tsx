"use client";

import { fetchShowSetByID } from "@/services/api.mediux";

import { useEffect, useState } from "react";

import { useRouter } from "next/navigation";

import { DimmedBackground } from "@/components/dimmed_backdrop";
import DownloadModalMovie from "@/components/download-modal-movie";
import DownloadModalShow from "@/components/download-modal-show";
import { SetFileCounts } from "@/components/set_file_counts";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "@/components/ui/accordion";
import { AssetImage } from "@/components/ui/asset-image";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ErrorMessage } from "@/components/ui/error-message";
import { H1, Lead } from "@/components/ui/typography";

import { useMediaStore } from "@/lib/mediaStore";
import { usePosterSetStore } from "@/lib/posterSetStore";

import { PosterFile } from "@/types/posterSets";

const SetPage = () => {
	const { posterSet, setPosterSet } = usePosterSetStore();
	const { mediaItem } = useMediaStore();
	const [backdropURL, setBackdropURL] = useState("");
	const [isDownloadModalOpen, setIsDownloadModalOpen] = useState(false);
	const router = useRouter();

	// Update the Poster Set
	useEffect(() => {
		if (posterSet?.ID) {
			const getShowSetByID = async () => {
				const resp = await fetchShowSetByID(posterSet.ID);
				if (resp.status !== "success") {
					return;
				}

				// Update the posterSet state with the latest data
				if (resp.data) {
					setPosterSet(resp.data);
				}
			};
			getShowSetByID();
		}
	}, [setPosterSet, posterSet?.ID]);

	// Construct the backdrop URL
	// If the posterSet has a backdrop, use that
	// Otherwise, use the mediaItem's backdrop
	useEffect(() => {
		if (typeof window !== "undefined") {
			// Safe to use document here.
			document.title = "Aura | Poster Set";
		}
		if (posterSet && posterSet.Backdrop) {
			const backdropFile = posterSet.Backdrop;
			if (backdropFile) {
				setBackdropURL(
					`/api/mediux/image/${backdropFile.ID}?modifiedDate=${backdropFile.Modified}&quality=full`
				);
			}
		} else {
			setBackdropURL(`/api/mediaserver/image/${mediaItem?.RatingKey}/backdrop`);
		}
	}, [mediaItem?.RatingKey, posterSet]);

	// Check if posterSet and mediaItem are defined
	if (!posterSet || !mediaItem) {
		return (
			<div className="flex flex-col items-center p-6 gap-4">
				<ErrorMessage message="Poster set or media item not found." />
				<Button
					className="mt-4"
					onClick={() => {
						window.location.href = "/";
					}}
				>
					Go to Home Page
				</Button>
			</div>
		);
	}

	const goToUserPage = () => {
		router.push(`/user/${posterSet.User.Name}`);
	};

	return (
		<div>
			{/* Backdrop Background */}
			{backdropURL && <DimmedBackground backdropURL={backdropURL} />}

			<div className="p-4 lg:p-6">
				<div className="pb-6">
					{/* Title */}
					<div className="flex flex-col pt-40 justify-end items-center text-center lg:items-start lg:text-left">
						<H1 className="mb-1">{posterSet.Title}</H1>
					</div>

					{/* Set Author */}
					<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center text-white gap-4 tracking-wide mt-4">
						{posterSet.User.Name && (
							<>
								<Badge
									className="flex items-center text-sm hover:text-white transition-colors"
									onClick={(e) => {
										e.stopPropagation();
										goToUserPage();
									}}
								>
									Set Author: {posterSet.User.Name}
								</Badge>
								<Badge
									className="flex items-center text-sm hover:text-white transition-colors"
									onClick={(e) => {
										e.stopPropagation();
										window.open(
											`https://mediux.pro/sets/${posterSet.ID}`,
											"_blank"
										);
									}}
								>
									View on Mediux
								</Badge>
							</>
						)}
					</div>
					<div className="flex flex-wrap justify-center lg:justify-start items-center text-white gap-4 tracking-wide mt-4">
						{posterSet.User.Name && (
							<>
								<SetFileCounts mediaItem={mediaItem} set={posterSet} />
							</>
						)}
					</div>
					<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center text-white gap-4 tracking-wide mt-4">
						{posterSet.User.Name && (
							<>
								<div className="ml-auto">
									{mediaItem.Type === "show" ? (
										<button className="btn">
											<DownloadModalShow
												posterSet={posterSet}
												mediaItem={mediaItem}
												open={isDownloadModalOpen}
												onOpenChange={setIsDownloadModalOpen}
											/>
										</button>
									) : mediaItem.Type === "movie" ? (
										<button className="btn">
											<DownloadModalMovie
												posterSet={posterSet}
												mediaItem={mediaItem}
												open={isDownloadModalOpen}
												onOpenChange={setIsDownloadModalOpen}
											/>
										</button>
									) : null}
								</div>
							</>
						)}
					</div>

					{/* Display the main poster */}
					{posterSet.Poster && (
						<div className="flex flex-col gap-2 mt-4">
							<Lead className="text-muted-foreground text-md">Poster</Lead>
							<div className="flex flex-wrap justify-center lg:justify-start items-center gap-4">
								<AssetImage
									image={posterSet.Poster as unknown as PosterFile}
									displayUser={true}
									displayMediaType={true}
									aspect="poster"
									className="w-[200px] h-auto"
								/>
							</div>
						</div>
					)}

					{/* Display the main backdrop */}
					{posterSet.Backdrop && (
						<div className="flex flex-col gap-2 mt-4">
							<Lead className="text-muted-foreground text-md">Backdrop</Lead>
							<div className="flex flex-wrap justify-center lg:justify-start items-center gap-4">
								<AssetImage
									image={posterSet.Backdrop as unknown as PosterFile}
									displayUser={true}
									displayMediaType={true}
									aspect="backdrop"
									className="w-[200px] h-auto"
								/>
							</div>
						</div>
					)}

					{/* Display all Other Posters (if any) */}
					{posterSet.OtherPosters && posterSet.OtherPosters.length > 0 && (
						<div className="flex flex-col gap-2 mt-4">
							<Lead className="text-muted-foreground text-md">
								{posterSet.OtherPosters.length > 1
									? "Other Posters"
									: "Other Poster"}
							</Lead>
							<div className="flex flex-wrap justify-center lg:justify-start items-center gap-4">
								{posterSet.OtherPosters.map((file) => {
									return (
										<AssetImage
											key={file.ID}
											image={file as unknown as PosterFile}
											displayUser={true}
											displayMediaType={true}
											aspect="poster"
											className="w-[200px] h-auto"
										/>
									);
								})}
							</div>
						</div>
					)}

					{/* Display all Other Backdrops (if any) */}
					{posterSet.OtherBackdrops && posterSet.OtherBackdrops.length > 0 && (
						<div className="flex flex-col gap-2 mt-4">
							<Lead className="text-muted-foreground text-md">
								{posterSet.OtherBackdrops.length > 1
									? "Other Backdrops"
									: "Other Backdrop"}
							</Lead>
							<div className="flex flex-wrap justify-center lg:justify-start items-center gap-4">
								{posterSet.OtherBackdrops.map((file) => {
									return (
										<AssetImage
											key={file.ID}
											image={file as unknown as PosterFile}
											displayUser={true}
											displayMediaType={true}
											aspect="backdrop"
											className="w-[200px] h-auto"
										/>
									);
								})}
							</div>
						</div>
					)}

					{/* Display all Season Posters (if any) */}
					{posterSet.SeasonPosters && posterSet.SeasonPosters.length > 0 && (
						<div className="flex flex-col gap-2 mt-4">
							<Lead className="text-muted-foreground text-md">
								{posterSet.SeasonPosters.length > 1
									? "Season Posters"
									: "Season Poster"}
							</Lead>
							<div className="flex flex-wrap justify-center lg:justify-start items-center gap-4">
								{[...posterSet.SeasonPosters]
									.sort(
										(a, b) => (b.Season?.Number ?? 0) - (a.Season?.Number ?? 0)
									)
									.map((file) => (
										<AssetImage
											key={file.ID}
											image={file as unknown as PosterFile}
											displayUser={true}
											displayMediaType={true}
											aspect="poster"
											className="w-[200px] h-auto"
										/>
									))}
							</div>
						</div>
					)}

					{/* Display all titlecards grouped by season (if any) */}
					{posterSet.TitleCards && posterSet.TitleCards.length > 0 && (
						<div className="flex flex-col gap-2 mt-4">
							<Lead className="text-muted-foreground text-md">
								{posterSet.TitleCards.length > 1 ? "Titlecards" : "Titlecard"}
							</Lead>
							<Accordion type="single" collapsible className="w-full">
								{Object.entries(
									posterSet.TitleCards.reduce(
										(acc, file) => {
											const season =
												file.Episode?.SeasonNumber ?? "Unknown Season";
											if (!acc[season]) {
												acc[season] = [];
											}
											acc[season].push(file);
											return acc;
										},
										{} as Record<string | number, PosterFile[]>
									)
								)
									.sort(([a], [b]) => {
										// Sort numerically, "Unknown Season" last
										if (a === "Unknown Season") return 1;
										if (b === "Unknown Season") return -1;
										return Number(b) - Number(a);
									})
									.map(([season, files]) => (
										<AccordionItem key={season} value={`season-${season}`}>
											<AccordionTrigger>
												<Lead className="text-muted-foreground text-md">
													Season {season}
												</Lead>
											</AccordionTrigger>
											<AccordionContent>
												<div className="flex flex-wrap justify-center lg:justify-start items-center gap-4">
													{files.map((file) => (
														<AssetImage
															key={file.ID}
															image={file as unknown as PosterFile}
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
