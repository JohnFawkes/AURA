package models

type MediuxContentIdsResponse struct {
	Data struct {
		ContentIds []MediuxContentID `json:"content_ids"`
	} `json:"data"`
}

type MediuxContentID struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
