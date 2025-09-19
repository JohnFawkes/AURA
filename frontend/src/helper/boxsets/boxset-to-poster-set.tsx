import { BoxsetCollectionToPosterSet } from "@/helper/boxsets/boxset-to-collection-poster-set";
import { BoxsetMovieToPosterSet } from "@/helper/boxsets/boxset-to-movie-poster-set";
import { BoxsetShowToPosterSet } from "@/helper/boxsets/boxset-to-show-poster-set";

import { log } from "@/lib/logger";

import { PosterSet } from "@/types/media-and-posters/poster-sets";
import { MediuxUserBoxset } from "@/types/mediux/mediux-sets";

// Function - Convert Boxset to Poster Set
export const BoxsetToPosterSet = (boxset: MediuxUserBoxset) => {
	const posterSets: PosterSet[] = [];

	// Convert Show Sets
	if (boxset.show_sets && boxset.show_sets.length > 0) {
		boxset.show_sets.forEach((showSet) => {
			const posterSet = BoxsetShowToPosterSet(showSet);
			posterSet.every((set) => {
				set.User.Name = boxset.user_created.username;
			});
			posterSets.push(...posterSet);
		});
	}

	// Convert Movie Sets
	if (boxset.movie_sets && boxset.movie_sets.length > 0) {
		boxset.movie_sets.forEach((movieSet) => {
			const posterSet = BoxsetMovieToPosterSet(movieSet);
			posterSet.every((set) => {
				set.User.Name = boxset.user_created.username;
			});
			posterSets.push(...posterSet);
		});
	}

	// Convert Collection Sets
	if (boxset.collection_sets && boxset.collection_sets.length > 0) {
		boxset.collection_sets.forEach((collectionSet) => {
			const posterSet = BoxsetCollectionToPosterSet(collectionSet);
			posterSet.every((set) => {
				set.User.Name = boxset.user_created.username;
			});
			posterSets.push(...posterSet);
		});
	}

	log("BoxsetToPosterSet - Converted Boxset to Poster Sets", {
		boxset: boxset.boxset_title,
		posterSets: posterSets,
	});
	return posterSets;
};
