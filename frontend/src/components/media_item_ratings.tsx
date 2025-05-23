import { Guid } from "@/types/mediaItem";
import Image from "next/image";

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
};

type MediaItemRatingsProps = {
	guids: Guid[];
	mediaItemType: string;
};

export function MediaItemRatings({
	guids,
	mediaItemType,
}: MediaItemRatingsProps) {
	const guidMap: { [provider: string]: ProviderInfo } = {};
	guids.forEach((guid: Guid) => {
		if (guid.Provider) {
			const providerInfo = providerLogoMap[guid.Provider];
			if (providerInfo) {
				guidMap[guid.Provider] = {
					id: guid.ID || "",
					rating:
						guids.find((g) => g.Provider === guid.Provider)
							?.Rating || "",
					logoUrl: providerInfo.logoUrl,
					linkUrl:
						guid.Provider === "tvdb"
							? `https://www.thetvdb.com/dereferrer/${
									mediaItemType === "show"
										? "series"
										: "movie"
							  }/${guid.ID}`
							: guid.Provider === "tmdb"
							? mediaItemType === "show"
								? `https://www.themoviedb.org/tv/${guid.ID}`
								: `https://www.themoviedb.org/movie/${guid.ID}`
							: `${providerInfo.urlPrefix}${guid.ID}`,
				};
			}
		}
	});

	return (
		<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide">
			{Object.entries(guidMap).map(([provider, info]) => (
				<div key={provider} className="flex items-center gap-2">
					{provider === "community" ? (
						<>
							{/* Display a star icon with the rating */}
							<span className="text-sm flex items-center gap-1">
								<svg
									xmlns="http://www.w3.org/2000/svg"
									width="16"
									height="16"
									fill="currentColor"
									viewBox="0 0 16 16"
								>
									<path d="M3.612 15.443c-.396.198-.86-.106-.746-.592l.83-4.73L.173 6.765c-.329-.32-.158-.888.283-.95l4.898-.696 2.189-4.327c.197-.39.73-.39.927 0l2.189 4.327 4.898.696c.441.062.612.63.282.95l-3.522 3.356.83 4.73c.114.486-.35.79-.746.592L8 13.187l-4.389 2.256z" />
								</svg>
								{info.rating}
							</span>
						</>
					) : (
						<>
							<a
								href={info.linkUrl!}
								target="_blank"
								rel="noopener noreferrer"
							>
								<div className="relative ml-1 w-[40px] h-[40px]">
									<Image
										src={info.logoUrl}
										alt={`${provider} Logo`}
										fill
										className="object-contain"
									/>
								</div>
							</a>
							{/* Only display rating if it exists */}
							{info.rating && (
								<span className="text-sm">{info.rating}</span>
							)}
						</>
					)}
				</div>
			))}
		</div>
	);
}
