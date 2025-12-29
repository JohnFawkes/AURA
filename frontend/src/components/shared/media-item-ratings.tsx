import apiClient from "@/services/api-client";

import { useEffect, useState } from "react";

import { AssetImage } from "@/components/shared/asset-image";

import { log } from "@/lib/logger";

import { Guid } from "@/types/media-and-posters/media-item-and-library";

interface ProviderInfo {
	id: string;
	rating: string;
	logoUrl: string;
	linkUrl: string;
}

const providerLogoMap: {
	[key: string]: { logoUrl: string; urlPrefix: string };
} = {
	imdb: {
		logoUrl: "/imdb-icon.png",
		urlPrefix: "https://www.imdb.com/title/",
	},
	tmdb: {
		logoUrl: "/tmdb-icon.svg",
		urlPrefix: "https://www.themoviedb.org/movie/",
	},
	tvdb: {
		logoUrl: "/tvdb-icon.svg",
		urlPrefix: "https://thetvdb.com/dereferrer/",
	},
	rottentomatoes: {
		logoUrl: "/rottentomatoes-icon.png",
		urlPrefix: "https://www.rottentomatoes.com/",
	},
	community: {
		logoUrl: "",
		urlPrefix: "",
	},
	user: {
		logoUrl: "/plex-icon.png",
		urlPrefix: "",
	},
};

type MediaItemRatingsProps = {
	guids: Guid[];
	mediaItemType: string;
	title: string;
};

export function MediaItemRatings({ guids, mediaItemType, title }: MediaItemRatingsProps) {
	const [mediuxURL, setMediuxURL] = useState<string>("");

	const guidMap: { [provider: string]: ProviderInfo } = {};

	const convertTitleToSlug = (title: string): string => {
		return title
			.toLowerCase()
			.replace(/[^a-z0-9\s-]/g, "") // Remove special characters
			.replace(/\s+/g, "_"); // Replace spaces with underscores
	};

	guids.forEach((guid: Guid) => {
		if (guid.Provider) {
			const providerInfo = providerLogoMap[guid.Provider];
			if (providerInfo) {
				guidMap[guid.Provider] = {
					id: guid.ID || "",
					rating: guids.find((g) => g.Provider === guid.Provider)?.Rating || "",
					logoUrl: providerInfo.logoUrl,
					linkUrl:
						guid.Provider === "tvdb"
							? `https://www.thetvdb.com/dereferrer/${
									mediaItemType === "show" ? "series" : "movie"
								}/${guid.ID}`
							: guid.Provider === "tmdb"
								? mediaItemType === "show"
									? `https://www.themoviedb.org/tv/${guid.ID}`
									: `https://www.themoviedb.org/movie/${guid.ID}`
								: guid.Provider === "rottentomatoes"
									? `https://www.rottentomatoes.com/${mediaItemType === "show" ? "tv" : "m"}/${convertTitleToSlug(title)}`
									: guid.Provider === "imdb"
										? `https://www.imdb.com/title/${guid.ID}`
										: guid.Provider === "community"
											? ""
											: // Default case for any other provider
												`${providerInfo.urlPrefix}${guid.ID}`,
				};
			}
		}
	});
	const tmdbID = guids?.find((g) => g.Provider === "tmdb")?.ID;

	// On Load -- Set the MediUX URL
	useEffect(() => {
		async function fetchMediuxURL() {
			if (tmdbID && (mediaItemType === "movie" || mediaItemType === "show")) {
				const response = await apiClient.get<{
					data: { status: string; exists: boolean; url: string };
				}>(`/mediux/check-link?itemType=${mediaItemType}&tmdbID=${tmdbID}`);
				if (response.data.data.url) {
					setMediuxURL(response.data.data.url);
				}
				log("INFO", "MediaItemRatings", "Fetched MediUX URL", `Found: ${response.data.data.exists}`, {
					response: response.data.data,
				});
			}
		}
		fetchMediuxURL();
	}, [tmdbID, mediaItemType]);

	return (
		<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide">
			{Object.entries(guidMap).map(([provider, info]) => (
				<div key={provider} className="flex items-center gap-2">
					{provider === "community" ? (
						<>
							{/* Display a star icon with the rating */}
							<span className="text-sm flex items-center gap-2">
								<svg
									xmlns="http://www.w3.org/2000/svg"
									width="17"
									height="17"
									fill="currentColor"
									viewBox="0 0 16 16"
								>
									<path d="M3.612 15.443c-.396.198-.86-.106-.746-.592l.83-4.73L.173 6.765c-.329-.32-.158-.888.283-.95l4.898-.696 2.189-4.327c.197-.39.73-.39.927 0l2.189 4.327 4.898.696c.441.062.612.63.282.95l-3.522 3.356.83 4.73c.114.486-.35.79-.746.592L8 13.187l-4.389 2.256z" />
								</svg>
							</span>
							{info.rating}
						</>
					) : provider !== "user" ? (
						<>
							<a href={info.linkUrl!} target="_blank" rel="noopener noreferrer">
								<AssetImage
									image={info.logoUrl}
									aspect="logo"
									className="relative mt-2 w-[40px] h-[30px]"
									imageClassName="object-contain"
								/>
							</a>
							{/* Only display rating if it exists */}
							{info.rating && <span className="text-sm">{info.rating}</span>}
						</>
					) : (
						<>
							<AssetImage
								image={info.logoUrl}
								className="relative w-[35px] h-[30px]"
								imageClassName="object-fill"
							/>
							{/* Only display rating if it exists */}
							{info.rating && <span className="text-sm">{info.rating}</span>}
						</>
					)}
				</div>
			))}

			{tmdbID && mediuxURL && (
				<a href={mediuxURL} target="_blank" rel="noopener noreferrer" className="border-none">
					<AssetImage
						image={"/mediux_logo.svg"}
						aspect="logo"
						className="relative lg:mt-5 w-[50px] h-[45px]"
						imageClassName="object-contain border-none"
					/>
				</a>
			)}
		</div>
	);
}
