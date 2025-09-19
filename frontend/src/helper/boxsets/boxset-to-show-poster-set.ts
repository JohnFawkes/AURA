import { PosterFileShow, PosterSet } from "@/types/media-and-posters/poster-sets";
import { MediuxUserShowSet } from "@/types/mediux/mediux-sets";

// Function - Convert Boxset Show Set to Poster Set
export const BoxsetShowToPosterSet = (showSet: MediuxUserShowSet) => {
	const showDetails: PosterFileShow = {
		ID: showSet.show_id.id,
		Title: showSet.show_id.title,
		MediaItem: showSet.show_id.MediaItem,
	};

	const posterSet: PosterSet = {
		ID: showSet.id,
		Title: showSet.set_title,
		Type: "show",
		User: {
			Name: showSet?.user_created?.username || "",
		},
		DateCreated: showSet.date_created,
		DateUpdated: showSet.date_updated,
		Poster:
			showSet.show_poster && showSet.show_poster.length > 0
				? {
						ID: showSet.show_poster[0].id,
						Type: "poster",
						Modified: showSet.show_poster[0].modified_on,
						FileSize: Number(showSet.show_poster[0].filesize),
						Show: showDetails,
					}
				: undefined,
		Backdrop:
			showSet.show_backdrop && showSet.show_backdrop.length > 0
				? {
						ID: showSet.show_backdrop[0].id,
						Type: "backdrop",
						Modified: showSet.show_backdrop[0].modified_on,
						FileSize: Number(showSet.show_backdrop[0].filesize),
						Show: showDetails,
					}
				: undefined,
		SeasonPosters: showSet.season_posters.map((poster) => ({
			ID: poster.id,
			Type: "seasonPoster",
			Modified: poster.modified_on,
			FileSize: Number(poster.filesize),
			Show: showDetails,
			Season: {
				Number: poster.season.season_number,
			},
		})),
		TitleCards: showSet.titlecards.map((titlecard) => ({
			ID: titlecard.id,
			Type: "titlecard",
			Modified: titlecard.modified_on,
			FileSize: Number(titlecard.filesize),
			Show: showDetails,
			Episode: {
				Title: titlecard.episode.episode_title,
				EpisodeNumber: titlecard.episode.episode_number,
				SeasonNumber: titlecard.episode.season_id.season_number,
			},
		})),
		Status: showSet.show_id.status,
	};
	return [posterSet];
};
