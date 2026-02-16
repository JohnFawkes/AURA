import { AlertTriangle } from "lucide-react";

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Lead } from "@/components/ui/typography";

import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { IncludedItem, SetRef } from "@/types/media-and-posters/sets";

interface SetFileCountsProps {
  mediaItem: MediaItem;
  set: SetRef;
  includedItems?: { [tmdb_id: string]: IncludedItem };
}

type EpisodeInfo = {
  episode_number: number;
  inMediaItem: boolean;
  setHasTitlecard: boolean;
  inSet: boolean;
  title?: string;
};

type SeasonInfo = {
  season_number: number;
  inMediaItem: boolean;
  setHasSeasonPoster: boolean;
  inSet: boolean;
  episodes: EpisodeInfo[];
};

type SeasonsAndEpisodesMap = Map<number, SeasonInfo>;

export function SetFileCounts({ mediaItem, set, includedItems }: SetFileCountsProps) {
  const { sections } = useLibrarySectionsStore();

  if (!set || !set.images || set.images.length === 0) return null;

  const librarySection = sections[mediaItem.library_title];

  // --- MOVIE LOGIC ---
  const getMovieFileCounts = () => {
    let primary = "";
    let secondary = "";

    const posterCount = set.images?.filter((img) => img.type === "poster").length || 0;
    const backdropCount = set.images?.filter((img) => img.type === "backdrop").length || 0;

    const parts: string[] = [];
    if (posterCount > 0) parts.push(`${posterCount} Poster${posterCount > 1 ? "s" : ""}`);
    if (backdropCount > 0) parts.push(`${backdropCount} Backdrop${backdropCount > 1 ? "s" : ""}`);
    primary = parts.join(" • ");

    const itemIDs = Array.from(
      new Set(
        (set.images ?? [])
          .map((img) => img.item_tmdb_id)
          .filter((id): id is string => typeof id === "string" && id.trim() !== "")
      )
    );
    let missingMovies: { id: string; title: string }[] = [];
    if (librarySection && itemIDs.length) {
      const libraryTmdbIDs = new Set(
        librarySection.media_items.filter((item) => item.tmdb_id).map((item) => item.tmdb_id)
      );

      missingMovies = itemIDs
        .filter((item_id) => !libraryTmdbIDs.has(item_id))
        .map((item_id) => {
          const title =
            includedItems?.[item_id]?.mediux_info.title ||
            librarySection.media_items.find((item) => item.tmdb_id === item_id)?.title ||
            includedItems?.[item_id]?.media_item.title ||
            item_id;
          return { id: item_id, title };
        });

      const totalTmdbIDs = itemIDs.length;
      const missingTmdbIDs = missingMovies.length;

      if (missingTmdbIDs === 0) {
        secondary = "";
      } else if (missingTmdbIDs === totalTmdbIDs) {
        secondary = `No items in Library`;
      } else {
        secondary = `${missingTmdbIDs} of ${totalTmdbIDs} movie${totalTmdbIDs > 1 ? "s" : ""} missing from library`;
      }
    }

    return {
      primary,
      secondary,
      breakdown: {
        missingFromLibrary: missingMovies,
      },
    };
  };

  // --- SHOW LOGIC ---
  const getShowFileCounts = () => {
    if (!mediaItem.series)
      return {
        line1: "",
        line2: "",
        hasMissing: false,
        missingFromSet: [],
        missingFromServer: [],
      };

    const posterCount = set.images?.filter((img) => img.type === "poster").length || 0;
    const backdropCount = set.images?.filter((img) => img.type === "backdrop").length || 0;
    const seasonPosterCount =
      set.images?.filter((img) => img.type === "season_poster" && img.season_number !== 0).length || 0;
    const specialSeasonPosterCount =
      set.images?.filter((img) => img.type === "season_poster" && img.season_number === 0).length || 0;
    const titleCardCount = set.images?.filter((img) => img.type === "titlecard").length || 0;

    const line1Parts: string[] = [];
    if (posterCount > 0) line1Parts.push(`${posterCount} Poster${posterCount > 1 ? "s" : ""}`);
    if (backdropCount > 0) line1Parts.push(`${backdropCount} Backdrop${backdropCount > 1 ? "s" : ""}`);
    const line2Parts: string[] = [];
    if (seasonPosterCount > 0) line2Parts.push(`${seasonPosterCount} Season Poster${seasonPosterCount > 1 ? "s" : ""}`);
    if (specialSeasonPosterCount > 0)
      line2Parts.push(`${specialSeasonPosterCount} Special Season Poster${specialSeasonPosterCount > 1 ? "s" : ""}`);
    if (titleCardCount > 0) line2Parts.push(`${titleCardCount} Titlecard${titleCardCount > 1 ? "s" : ""}`);

    const missingFromSet: string[] = [];
    const missingFromServer: string[] = [];

    // --- Build SeasonsAndEpisodesMap ---
    const mediaSeasons = mediaItem.series?.seasons ?? [];
    const setSeasonPosters = set.images.filter(
      (img) => img.type === "season_poster" && typeof img.season_number === "number"
    );
    const setTitlecards = set.images.filter(
      (img) =>
        img.type === "titlecard" && typeof img.season_number === "number" && typeof img.episode_number === "number"
    );

    const allSeasonNumbers = new Set<number>([
      ...mediaSeasons.map((s) => s.season_number).filter((n): n is number => typeof n === "number"),
      ...setSeasonPosters.map((img) => img.season_number as number),
      ...setTitlecards.map((img) => img.season_number as number),
    ]);

    const SeasonsAndEpisodesMap: SeasonsAndEpisodesMap = new Map();

    allSeasonNumbers.forEach((season_number) => {
      const mediaSeason = mediaSeasons.find((s) => s.season_number === season_number);
      const setHasSeasonPoster = setSeasonPosters.some((img) => img.season_number === season_number);
      const setHasTitlecards = setTitlecards.some((img) => img.season_number === season_number);

      const mediaEpisodes = mediaSeason?.episodes ?? [];
      const setEpisodes = setTitlecards.filter((img) => img.season_number === season_number);

      const allEpisodeNumbers = new Set<number>([
        ...mediaEpisodes.map((e) => e.episode_number).filter((n): n is number => typeof n === "number"),
        ...setEpisodes.map((img) => img.episode_number as number),
      ]);

      const episodes: EpisodeInfo[] = [];
      allEpisodeNumbers.forEach((episode_number) => {
        const mediaEp = mediaEpisodes.find((e) => e.episode_number === episode_number);
        const setEp = setEpisodes.find((img) => img.episode_number === episode_number);

        episodes.push({
          episode_number,
          inMediaItem: !!mediaEp,
          setHasTitlecard: !!setEp,
          inSet: !!setEp,
          title: mediaEp?.title || setEp?.title || "",
        });
      });

      SeasonsAndEpisodesMap.set(season_number, {
        season_number,
        inMediaItem: !!mediaSeason,
        setHasSeasonPoster,
        inSet: setHasSeasonPoster || setHasTitlecards,
        episodes,
      });
    });

    // --- Analyze for missing items ---
    SeasonsAndEpisodesMap.forEach((seasonInfo, season_number) => {
      let missingEntireSeason = false;
      // Season/Season Poster
      if (seasonInfo.inMediaItem) {
        if (seasonInfo.inSet) {
          // All good
        } else {
          if (seasonInfo.setHasSeasonPoster) {
            missingFromSet.push(`Season ${season_number} Poster`);
          } else {
            // Set doesn't contain season posters, all good
          }
        }
      } else if (seasonInfo.inSet) {
        if (seasonInfo.setHasSeasonPoster) {
          missingEntireSeason = true;
          if (season_number === 0) {
            missingFromServer.push(`Special Season`);
          } else {
            missingFromServer.push(
              `Season ${season_number} ${seasonInfo.episodes.length > 0 ? `(${seasonInfo.episodes.length} Episode${seasonInfo.episodes.length > 1 ? "s" : ""}) ` : ""}`
            );
            seasonInfo.episodes.forEach(() => {
              missingFromServer.push("");
            });
          }
        }
      }

      // Episodes/Titlecards
      seasonInfo.episodes.forEach((epInfo) => {
        if (epInfo.inMediaItem) {
          if (epInfo.inSet) {
            // All good
          } else {
            if (epInfo.setHasTitlecard) {
              missingFromSet.push(
                `S${season_number}E${epInfo.episode_number} Titlecard ${epInfo.title ? `- ${epInfo.title}` : ""}`
              );
            } else {
              // Set doesn't contain titlecards, all good
            }
          }
        } else if (epInfo.inSet) {
          if (epInfo.setHasTitlecard) {
            if (seasonInfo.setHasSeasonPoster) {
              if (!missingEntireSeason) {
                missingFromServer.push(
                  `Season ${season_number} Episode ${epInfo.episode_number} ${epInfo.title ? `- ${epInfo.title}` : ""}`
                );
              }
            } else {
              // If all episodes are missing, list season instead
              const allEpisodesMissing = seasonInfo.episodes.every((ep) => ep.inSet && ep.setHasTitlecard);
              if (allEpisodesMissing) {
                if (
                  !missingFromServer.includes(
                    `Season ${season_number} (${seasonInfo.episodes.length} Episode${seasonInfo.episodes.length > 1 ? "s" : ""})`
                  )
                ) {
                  missingFromServer.push(
                    `Season ${season_number} (${seasonInfo.episodes.length} Episode${seasonInfo.episodes.length > 1 ? "s" : ""})`
                  );
                  // Subtract one episode, this is so that we don't double-list episodes
                  seasonInfo.episodes.slice(1).forEach(() => {
                    missingFromServer.push("");
                  });
                }
              } else {
                missingFromServer.push(
                  `S${season_number}E${epInfo.episode_number} ${epInfo.title ? `- ${epInfo.title}` : ""}`
                );
              }
            }
          }
        }
      });
    });

    missingFromSet.sort();
    missingFromServer.sort();

    return {
      line1: line1Parts.join(" • "),
      line2: line2Parts.join(" • "),
      hasMissing: missingFromSet.length > 0 || missingFromServer.length > 0,
      missingFromSet,
      missingFromServer,
    };
  };

  const movieResult = mediaItem.type === "movie" ? getMovieFileCounts() : null;
  const showResult = mediaItem.type === "show" ? getShowFileCounts() : null;

  return (
    <div>
      {mediaItem.type === "movie" && movieResult && (
        <>
          {movieResult.primary && <Lead className="text-sm text-muted-foreground ml-1">{movieResult.primary}</Lead>}
          {movieResult.secondary && (
            <div className="flex items-center mb-2">
              <Lead className="text-sm ml-1 text-yellow-500">{movieResult.secondary}</Lead>
              <Popover>
                <PopoverTrigger asChild>
                  <button className="ml-1 text-muted-foreground hover:text-primary">
                    <AlertTriangle className="text-yellow-500" size={16} />
                  </button>
                </PopoverTrigger>
                <PopoverContent className="p-4 max-w-xs">
                  <div>
                    {"missingFromLibrary" in movieResult.breakdown &&
                    movieResult.breakdown.missingFromLibrary.length > 0 ? (
                      <div>
                        <div className="font-semibold">Missing from Library</div>
                        <ul className="text-sm list-disc list-inside">
                          {movieResult.breakdown.missingFromLibrary.map((movie, idx) => (
                            <li key={idx}>{movie.title}</li>
                          ))}
                        </ul>
                      </div>
                    ) : (
                      <div className="text-sm text-muted-foreground">No missing items</div>
                    )}
                  </div>
                </PopoverContent>
              </Popover>
            </div>
          )}
        </>
      )}
      {mediaItem.type === "show" && showResult && (
        <>
          {showResult.line1 && <Lead className="text-sm text-muted-foreground ml-1">{showResult.line1}</Lead>}
          {showResult.line2 && <Lead className="text-sm text-muted-foreground ml-1">{showResult.line2}</Lead>}
          {showResult.hasMissing && (
            <div className="flex items-center mb-2">
              {showResult.missingFromSet.length > 0 && showResult.missingFromServer.length > 0 ? (
                <Lead className="text-sm ml-1 text-yellow-500">
                  {showResult.missingFromSet.length} item
                  {showResult.missingFromSet.length > 1 ? "s" : ""} missing from set •{" "}
                  {showResult.missingFromServer.length} item
                  {showResult.missingFromServer.length > 1 ? "s" : ""} missing from server
                </Lead>
              ) : showResult.missingFromSet.length > 0 ? (
                <Lead className="text-sm ml-1 text-yellow-500">
                  {showResult.missingFromSet.length} item
                  {showResult.missingFromSet.length > 1 ? "s" : ""} missing from set
                </Lead>
              ) : (
                <Lead className="text-sm ml-1 text-yellow-500">
                  {showResult.missingFromServer.length} item
                  {showResult.missingFromServer.length > 1 ? "s" : ""} missing from server
                </Lead>
              )}

              <Popover>
                <PopoverTrigger asChild>
                  <button className="ml-1 text-muted-foreground hover:text-primary">
                    <AlertTriangle className="text-yellow-500" size={16} />
                  </button>
                </PopoverTrigger>
                <PopoverContent className="p-4 max-w-xs">
                  <div>
                    {showResult.missingFromSet.length > 0 && (
                      <>
                        <div className="font-semibold">Missing from Set</div>
                        <ul className="text-sm list-disc list-inside">
                          {showResult.missingFromSet.map((item, idx) => (
                            <li key={idx}>{item}</li>
                          ))}
                        </ul>
                      </>
                    )}
                    {showResult.missingFromServer.length > 0 && (
                      <>
                        <div className="font-semibold mt-2">Missing from Server</div>
                        <ul className="text-sm list-disc list-inside">
                          {showResult.missingFromServer.map((item, idx) => item !== "" && <li key={idx}>{item}</li>)}
                        </ul>
                      </>
                    )}
                    {showResult.missingFromSet.length === 0 && showResult.missingFromServer.length === 0 && (
                      <div className="text-sm text-muted-foreground">No missing items</div>
                    )}
                  </div>
                </PopoverContent>
              </Popover>
            </div>
          )}
        </>
      )}
    </div>
  );
}
