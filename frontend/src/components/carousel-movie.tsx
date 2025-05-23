"use client";

import { CarouselItem } from "@/components/ui/carousel";
import { AssetImage } from "@/components/ui/asset-image";
import { PosterFile, PosterSet } from "@/types/posterSets";

export function CarouselMovie({ set }: { set: PosterSet }) {
	return (
		<>
			{/* Main Movie Poster and Backdrop */}
			{(set.Poster || set.Backdrop) && (
				<CarouselItem
					key={`${set.ID}-${set.Poster?.ID || "no-poster"}`}
				>
					<div className="space-y-2">
						{set.Poster && (
							<AssetImage
								image={set.Poster as unknown as PosterFile}
								displayUser={true}
								displayMediaType={true}
								aspect="poster"
								className={`w-full ${
									!set.Poster.Movie?.RatingKey
										? "opacity-35"
										: ""
								}`}
							/>
						)}

						{set.Backdrop && (
							<AssetImage
								image={set.Backdrop as unknown as PosterFile}
								displayUser={true}
								displayMediaType={true}
								aspect="backdrop"
								className={`w-full ${
									!set.Backdrop.Movie?.RatingKey
										? "opacity-35"
										: ""
								}`}
							/>
						)}
					</div>
				</CarouselItem>
			)}

			{/* Other Posters and Backdrops in Set*/}
			{set.OtherPosters?.sort(
				(a: PosterFile, b: PosterFile) =>
					new Date(b.Movie?.ReleaseDate ?? 0).getTime() -
					new Date(a.Movie?.ReleaseDate ?? 0).getTime()
			).map((poster: PosterFile) => {
				const matchingBackdrop = set.OtherBackdrops?.find(
					(backdrop: PosterFile) =>
						backdrop.Movie?.ID === poster.Movie?.ID
				);
				return (
					<CarouselItem key={`${set.ID}-other-${poster.ID}`}>
						<div className="space-y-2">
							<AssetImage
								image={poster as unknown as PosterFile}
								displayUser={true}
								displayMediaType={true}
								aspect="poster"
								className={`w-full ${
									!poster.Movie?.RatingKey ? "opacity-35" : ""
								}`}
							/>

							{matchingBackdrop && (
								<AssetImage
									image={
										matchingBackdrop as unknown as PosterFile
									}
									displayUser={true}
									displayMediaType={true}
									aspect="backdrop"
									className={`w-full ${
										!matchingBackdrop.Movie?.RatingKey
											? "opacity-35"
											: ""
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
