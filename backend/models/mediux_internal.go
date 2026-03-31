package models

type MediuxContentIdsResponse struct {
	Movies []MediuxContentID `json:"movies"`
	Shows  []MediuxContentID `json:"shows"`
}

type MediuxContentID struct {
	ID string `json:"id"`
}
