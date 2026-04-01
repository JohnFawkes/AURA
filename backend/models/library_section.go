package models

type LibrarySectionBase struct {
	ID    string   `json:"id" yaml:"ID,omitempty" mapstructure:"ID"`                    // Unique identifier for the library section.
	Title string   `json:"title" yaml:"Title" mapstructure:"Title"`                     // Title of the library section.
	Type  string   `json:"type" yaml:"Type,omitempty" mapstructure:"Type"`              // "movie" or "show"
	Paths []string `json:"paths,omitempty" yaml:"Paths,omitempty" mapstructure:"Paths"` // Paths of the library section on the media server.
}

type LibrarySection struct {
	LibrarySectionBase `yaml:",inline" mapstructure:",squash"`
	TotalSize          int         `json:"total_size" yaml:"TotalSize,omitempty" mapstructure:"TotalSize"`
	MediaItems         []MediaItem `json:"media_items" yaml:"MediaItems,omitempty" mapstructure:"MediaItems"`
}
