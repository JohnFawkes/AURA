import { FormItemDisplay } from "@/components/shared/download-modal";

import { PosterFile, PosterSet } from "@/types/posterSets";

const createBaseItem = (set: PosterSet): FormItemDisplay => ({
	MediaItemRatingKey: "",
	MediaItemTitle: "",
	SetID: set.ID,
	Set: {
		ID: set.ID,
		Title: set.Title,
		Type: set.Type,
		User: { Name: set.User.Name },
		DateCreated: set.DateCreated,
		DateUpdated: set.DateUpdated,
		Status: set.Status,
	},
});

const handleAdditionalFiles = (
	items: FormItemDisplay[],
	files: PosterFile[] | undefined,
	fileType: "Poster" | "Backdrop",
	set: PosterSet
) => {
	files?.forEach((file) => {
		const ratingKey = file.Movie?.MediaItem.RatingKey;
		const existingItem = items.find((item) => item.MediaItemRatingKey === ratingKey);

		if (existingItem && !existingItem.Set[fileType]) {
			existingItem.Set[fileType] = file;
		} else if (!existingItem && ratingKey) {
			const newItem = createBaseItem(set);
			newItem.MediaItemRatingKey = ratingKey;
			newItem.MediaItemTitle = file.Movie?.Title || file.Show?.Title || "";
			newItem.Set[fileType] = file;
			items.push(newItem);
		}
	});
};

export const posterSetToFormItem = (set: PosterSet): FormItemDisplay[] => {
	const items: FormItemDisplay[] = [];
	const item = createBaseItem(set);
	if (set.Type === "show") {
		item.MediaItemRatingKey =
			set.Poster?.Show?.MediaItem.RatingKey ||
			set.Backdrop?.Show?.MediaItem.RatingKey ||
			set.SeasonPosters?.find((poster) => poster.Show?.MediaItem?.RatingKey)?.Show?.MediaItem.RatingKey ||
			set.TitleCards?.find((card) => card.Show?.MediaItem?.RatingKey)?.Show?.MediaItem.RatingKey ||
			"";
		item.MediaItemTitle =
			set.Poster?.Show?.Title ||
			set.Backdrop?.Show?.Title ||
			set.SeasonPosters?.find((poster) => poster.Show?.Title)?.Show?.Title ||
			set.TitleCards?.find((card) => card.Show?.Title)?.Show?.Title ||
			"";

		if (!item.MediaItemRatingKey) return [];

		Object.assign(item.Set, {
			...(set.Poster && { Poster: set.Poster }),
			...(set.Backdrop && { Backdrop: set.Backdrop }),
			// Filter SeasonPosters to only include those with a valid RatingKey
			...((set.SeasonPosters ?? []).filter((poster) => poster.Show?.MediaItem?.RatingKey).length > 0 && {
				SeasonPosters: (set.SeasonPosters ?? []).filter((poster) => poster.Show?.MediaItem?.RatingKey),
			}),
			// Filter TitleCards to only include those with a valid RatingKey
			...((set.TitleCards ?? []).filter((card) => card.Show?.MediaItem?.RatingKey).length > 0 && {
				TitleCards: (set.TitleCards ?? []).filter((card) => card.Show?.MediaItem?.RatingKey),
			}),
		});

		items.push(item);
		return items;
	} else if (set.Type === "movie" || set.Type === "collection") {
		item.MediaItemRatingKey =
			set.Poster?.Movie?.MediaItem.RatingKey || set.Backdrop?.Movie?.MediaItem.RatingKey || "";
		item.MediaItemTitle = set.Poster?.Movie?.Title || set.Backdrop?.Movie?.Title || "";
		if (!item.MediaItemRatingKey) return [];

		Object.assign(item.Set, {
			...(set.Poster && { Poster: set.Poster }),
			...(set.Backdrop && { Backdrop: set.Backdrop }),
		});

		handleAdditionalFiles(items, set.OtherPosters, "Poster", set);
		handleAdditionalFiles(items, set.OtherBackdrops, "Backdrop", set);

		items.push(item);
	}

	return items;
};
