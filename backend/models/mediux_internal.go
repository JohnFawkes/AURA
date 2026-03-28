package models

type MediuxContentIdsResponse struct {
	Items []MediuxContentID `json:"items"`
}

type MediuxContentID struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
