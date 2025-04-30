import { fetchMediuxSets } from "@/services/api.mediux";
import { MediaItem } from "@/types/mediaItem";
import { PosterSets } from "@/types/posterSets";

export async function getPosterSets(mediaItem: MediaItem): Promise<{
	posterSets: PosterSets;
}> {
	try {
		if (!mediaItem.Guids || mediaItem.Guids.length === 0) {
			return Promise.reject("No GUIDs found in Media Item");
		}
		const tmdbID = mediaItem.Guids.find(
			(guid) => guid.Provider === "tmdb"
		)?.ID;
		if (!tmdbID) {
			return Promise.reject("No TMDB ID found in Media Item");
		}
		const resp = await fetchMediuxSets(tmdbID, mediaItem.Type);
		if (!resp) {
			throw new Error("No response from Mediux API");
		} else if (resp.status !== "success") {
			throw new Error(resp.message);
		}
		const sets = resp.data;
		if (!sets) {
			throw new Error("No sets found in response");
		}
		return { posterSets: sets };
	} catch (error) {
		if (error instanceof Error) {
			console.error("Error fetching Media Item:", error.message);
			return Promise.reject(error.message);
		} else {
			console.error("An unknown error occurred:", error);
			return Promise.reject("An unknown error occurred, check logs.");
		}
	}
}
