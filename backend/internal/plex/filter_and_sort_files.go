package plex

import (
	"poster-setter/internal/modals"
	"slices"
	"sort"
)

func FilterAndSortFiles(files []modals.PosterFile, selectTypes []string) []modals.PosterFile {

	// Filter out the selected types from the Set.Files
	selectedFiles := make([]modals.PosterFile, 0)
	for _, file := range files {
		if slices.Contains(selectTypes, file.Type) {
			selectedFiles = append(selectedFiles, file)
		}
	}
	// Sort the selected files in the specified order
	sort.Slice(selectedFiles, func(i, j int) bool {
		// Define the order of types
		typeOrder := map[string]int{
			"poster":       1,
			"backdrop":     2,
			"seasonPoster": 3,
			"titlecard":    4,
		}

		// Compare by type order first
		if typeOrder[selectedFiles[i].Type] != typeOrder[selectedFiles[j].Type] {
			return typeOrder[selectedFiles[i].Type] < typeOrder[selectedFiles[j].Type]
		}

		// If types are the same, handle specific sorting logic for each type
		switch selectedFiles[i].Type {
		case "seasonPoster":
			// Sort season posters by Season.Number
			return selectedFiles[i].Season.Number < selectedFiles[j].Season.Number
		case "titlecard":
			// Sort titlecards by Season.Number first, then by Episode.Number
			if selectedFiles[i].Episode.SeasonNumber != selectedFiles[j].Episode.SeasonNumber {
				return selectedFiles[i].Episode.SeasonNumber < selectedFiles[j].Episode.SeasonNumber
			}
			return selectedFiles[i].Episode.EpisodeNumber < selectedFiles[j].Episode.EpisodeNumber
		}

		// If all else is the same, compare alphabetically by ID
		return selectedFiles[i].ID < selectedFiles[j].ID
	})

	return selectedFiles
}
