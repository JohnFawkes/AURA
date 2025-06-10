"use client";

import { useEffect, useState } from "react";

import { AssetImage } from "@/components/shared/asset-image";
import { CarouselItem } from "@/components/ui/carousel";

import { PosterFile, PosterSet } from "@/types/posterSets";

import { searchIDBForTMDBID } from "../../helper/searchIDBForTMDBID";

export function CarouselMovie({ set, librarySection }: { set: PosterSet; librarySection: string }) {
	const [idbResults, setIdbResults] = useState<{ [key: string]: boolean }>({});

	useEffect(() => {
		let isMounted = true;
		if (set.OtherPosters) {
			const fetchData = async () => {
				const posterPromises = (set.OtherPosters ?? []).map(async (poster) => {
					if (!poster.Movie?.RatingKey) {
						const result = await searchIDBForTMDBID(
							poster.Movie?.ID || "",
							librarySection
						);
						return { id: poster.ID, isInIDB: !!result };
					}
					return null;
				});
				const results = await Promise.all(posterPromises);
				if (isMounted) {
					const newResults: { [key: string]: boolean } = {};
					results.forEach((res) => {
						if (res) {
							newResults[res.id] = res.isInIDB;
						}
					});
					setIdbResults(newResults);
				}
			};
			fetchData();
		}
		return () => {
			isMounted = false;
		};
	}, [set.OtherPosters, librarySection]);

	return (
		<>
			{/* Main Movie Poster and Backdrop */}
			{(set.Poster || set.Backdrop) && (
				<CarouselItem key={`${set.ID}-${set.Poster?.ID || "no-poster"}`}>
					<div className="space-y-2">
						{set.Poster && (
							<AssetImage
								image={set.Poster as unknown as PosterFile}
								aspect="poster"
								imageClassName="w-full"
							/>
						)}

						{set.Backdrop && (
							<AssetImage
								image={set.Backdrop as unknown as PosterFile}
								aspect="backdrop"
								className="w-full"
							/>
						)}
					</div>
				</CarouselItem>
			)}

			{/* Other Posters and Backdrops in Set */}
			{set.OtherPosters?.sort(
				(a: PosterFile, b: PosterFile) =>
					new Date(b.Movie?.ReleaseDate ?? 0).getTime() -
					new Date(a.Movie?.ReleaseDate ?? 0).getTime()
			).map((poster: PosterFile) => {
				const matchingBackdrop = set.OtherBackdrops?.find(
					(backdrop: PosterFile) => backdrop.Movie?.ID === poster.Movie?.ID
				);
				const isInIDB = idbResults[poster.ID] || false;
				return (
					<CarouselItem key={`${set.ID}-other-${poster.ID}`}>
						<div className="space-y-2">
							<AssetImage
								image={poster as unknown as PosterFile}
								aspect="poster"
								className={`w-full h-auto ${
									!poster.Movie?.RatingKey && !isInIDB ? "opacity-35" : ""
								}`}
							/>

							{matchingBackdrop && (
								<AssetImage
									image={matchingBackdrop as unknown as PosterFile}
									aspect="backdrop"
									className={`w-full h-auto ${
										!matchingBackdrop.Movie?.RatingKey ? "opacity-35" : ""
									}`}
								/>
							)}
						</div>
					</CarouselItem>
				);
			})}
		</>
	);
}
