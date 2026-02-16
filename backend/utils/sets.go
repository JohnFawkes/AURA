package utils

import "aura/models"

func UpdatePosterSetInDBItem(posterSets []models.DBPosterSetDetail, newSetInfo models.DBPosterSetDetail) (found bool, updatedPosterSets []models.DBPosterSetDetail) {
	found = false
	for idx, set := range posterSets {
		if set.ID == newSetInfo.ID {
			posterSets[idx] = newSetInfo
			found = true
			break
		}
	}
	if !found {
		return false, posterSets
	}
	return true, posterSets
}
