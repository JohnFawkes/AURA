package modals

import "encoding/xml"

type PlexPhotosResponse struct {
	XMLName xml.Name        `xml:"MediaContainer"`
	Size    int             `xml:"size,attr"`
	Photos  []PlexPhotoItem `xml:"Photo"`
}

type PlexPhotoItem struct {
	Key       string `xml:"key,attr"`
	RatingKey string `xml:"ratingKey,attr"`
	Thumb     string `xml:"thumb,attr"`
	Selected  int    `xml:"selected,attr"`
	Provider  string `xml:"provider,attr,omitempty"`
}
