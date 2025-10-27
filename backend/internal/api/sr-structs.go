package api

import (
	"time"
)

type SR_ItemInfoBase struct {
	Type             string      `json:"type,omitempty"` // "Sonarr" or "Radarr"
	ID               int64       `json:"id,omitempty"`
	Title            string      `json:"title,omitempty"`
	Added            time.Time   `json:"added"`
	Certification    string      `json:"certification,omitempty"`
	CleanTitle       string      `json:"cleanTitle"`
	Folder           string      `json:"folder,omitempty"`
	Genres           []string    `json:"genres,omitempty"`
	Images           []SR_Image  `json:"images,omitempty"`
	ImdbID           string      `json:"imdbId,omitempty"`
	Monitored        bool        `json:"monitored,omitempty"`
	OriginalLanguage SR_Language `json:"originalLanguage"`
	Overview         string      `json:"overview,omitempty"`
	Path             string      `json:"path,omitempty"`
	QualityProfileID int64       `json:"qualityProfileId,omitempty"`
	RemotePoster     string      `json:"remotePoster,omitempty"`
	RootFolderPath   string      `json:"rootFolderPath,omitempty"`
	Runtime          int64       `json:"runtime,omitempty"`
	SortTitle        string      `json:"sortTitle,omitempty"`
	Status           string      `json:"status,omitempty"`
	Tags             []int64     `json:"tags,omitempty"`
	TitleSlug        string      `json:"titleSlug,omitempty"`
	TmdbID           int64       `json:"tmdbId,omitempty"`
	Year             int64       `json:"year,omitempty"`
}

type SR_SonarrItem struct {
	SR_ItemInfoBase
	AddOptions        Sonarr_AddOptions       `json:"addOptions"`
	AirTime           string                  `json:"airTime,omitempty"`
	AlternateTitles   []Sonarr_AlternateTitle `json:"alternateTitles"`
	EpisodesChanged   string                  `json:"episodesChanged,omitempty"`
	FirstAired        string                  `json:"firstAired,omitempty"`
	LastAired         string                  `json:"lastAired,omitempty"`
	Monitored         bool                    `json:"monitored,omitempty"`
	MonitorNewItems   string                  `json:"monitorNewItems,omitempty"`
	Network           string                  `json:"network,omitempty"`
	NextAiring        string                  `json:"nextAiring,omitempty"`
	PreviousAiring    string                  `json:"previousAiring,omitempty"`
	ProfileName       string                  `json:"profileName,omitempty"`
	Ratings           SR_Ratings              `json:"ratings"`
	SeasonFolder      bool                    `json:"seasonFolder,omitempty"`
	Seasons           []Sonarr_Season         `json:"seasons"`
	SeriesType        string                  `json:"seriesType,omitempty"`
	Statistics        Sonarr_ItemStatistics   `json:"statistics"`
	Status            string                  `json:"status,omitempty"`
	TmdbID            int64                   `json:"tmdbId,omitempty"`
	TvdbID            int64                   `json:"tvdbId,omitempty"`
	TvMazeID          int64                   `json:"tvMazeId,omitempty"`
	TvRageID          int64                   `json:"tvRageId,omitempty"`
	UseSceneNumbering bool                    `json:"useSceneNumbering,omitempty"`
}

type SR_RadarrItem struct {
	SR_ItemInfoBase
	AddOptions            Radarr_AddOptions       `json:"addOptions"`
	AlternateTitles       []Radarr_AlternateTitle `json:"alternateTitles"`
	Collection            Radarr_Collection       `json:"collection"`
	DigitalRelease        time.Time               `json:"digitalRelease"`
	FolderName            string                  `json:"folderName,omitempty"`
	HasFile               bool                    `json:"hasFile,omitempty"`
	InCinemas             time.Time               `json:"inCinemas"`
	IsAvailable           bool                    `json:"isAvailable,omitempty"`
	Keywords              []string                `json:"keywords,omitempty"`
	LastSearchTime        time.Time               `json:"lastSearchTime"`
	MinimumAvailability   string                  `json:"minimumAvailability,omitempty"`
	MovieFile             Radarr_MovieFile        `json:"movieFile"`
	MovieFileID           int64                   `json:"movieFileId,omitempty"`
	OriginalTitle         string                  `json:"originalTitle,omitempty"`
	PhysicalRelease       time.Time               `json:"physicalRelease"`
	PhysicalReleaseNote   string                  `json:"physicalReleaseNote,omitempty"`
	Popularity            float64                 `json:"popularity,omitempty"`
	Ratings               Radarr_Ratings          `json:"ratings"`
	ReleaseDate           time.Time               `json:"releaseDate"`
	SecondaryYear         int64                   `json:"secondaryYear,omitempty"`
	SecondaryYearSourceID int64                   `json:"secondaryYearSourceId,omitempty"`
	SizeOnDisk            int64                   `json:"sizeOnDisk,omitempty"`
	Statistics            Radarr_ItemStatistics   `json:"statistics"`
	Studio                string                  `json:"studio,omitempty"`
}

type Sonarr_AlternateTitle struct {
	Title             string `json:"title,omitempty"`
	SeasonNumber      int    `json:"seasonNumber,omitempty"`
	SceneSeasonNumber string `json:"sceneSeasonNumber,omitempty"`
	SceneOrigin       string `json:"sceneOrigin,omitempty"`
	Comment           string `json:"comment,omitempty"`
}

type Radarr_AlternateTitle struct {
	ID              int64  `json:"id,omitempty"`
	SourceType      string `json:"sourceType,omitempty"`
	MovieMetadataID int64  `json:"movieMetadataId,omitempty"`
	Title           string `json:"title,omitempty"`
	CleanTitle      string `json:"cleanTitle,omitempty"`
}

type Radarr_Collection struct {
	Title  string `json:"title,omitempty"`
	TmdbID int64  `json:"tmdbId,omitempty"`
}

type SR_Image struct {
	CoverType string `json:"coverType,omitempty"`
	URL       string `json:"url,omitempty"`
	RemoteURL string `json:"remoteUrl,omitempty"`
}

type SR_Language struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type SR_Ratings struct {
	Votes int64   `json:"votes,omitempty"`
	Value float64 `json:"value,omitempty"`
	Type  string  `json:"type,omitempty"`
}

type Radarr_Ratings struct {
	Imdb           SR_Ratings `json:"imdb"`
	Tmdb           SR_Ratings `json:"tmdb"`
	Metacritic     SR_Ratings `json:"metacritic"`
	RottenTomatoes SR_Ratings `json:"rottenTomatoes"`
	Trakt          SR_Ratings `json:"trakt"`
}

type Sonarr_ItemStatistics struct {
	SeasonCount       int64    `json:"seasonCount,omitempty"`
	EpisodeFileCount  int64    `json:"episodeFileCount,omitempty"`
	EpisodeCount      int64    `json:"episodeCount,omitempty"`
	TotalEpisodeCount int64    `json:"totalEpisodeCount,omitempty"`
	SizeOnDisk        int64    `json:"sizeOnDisk,omitempty"`
	ReleaseGroups     []string `json:"releaseGroups,omitempty"`
}

type Radarr_ItemStatistics struct {
	MovieFileCount int64    `json:"movieFileCount,omitempty"`
	SizeOnDisk     int64    `json:"sizeOnDisk,omitempty"`
	ReleaseGroups  []string `json:"releaseGroups,omitempty"`
}

type Sonarr_Season struct {
	SeasonNumber int64                   `json:"seasonNumber,omitempty"`
	Monitored    bool                    `json:"monitored,omitempty"`
	Statistics   Sonarr_SeasonStatistics `json:"statistics"`
	Images       []SR_Image              `json:"images"`
}

type Sonarr_SeasonStatistics struct {
	NextAiring        string   `json:"nextAiring,omitempty"`
	PreviousAiring    string   `json:"previousAiring,omitempty"`
	EpisodeFileCount  int64    `json:"episodeFileCount,omitempty"`
	EpisodeCount      int64    `json:"episodeCount,omitempty"`
	TotalEpisodeCount int64    `json:"totalEpisodeCount,omitempty"`
	SizeOnDisk        int64    `json:"sizeOnDisk,omitempty"`
	ReleaseGroups     []string `json:"releaseGroups,omitempty"`
}

type Sonarr_AddOptions struct {
	IgnoreEpisodesWithFiles      bool   `json:"ignoreEpisodesWithFiles,omitempty"`
	IgnoreEpisodesWithoutFiles   bool   `json:"ignoreEpisodesWithoutFiles,omitempty"`
	Monitor                      string `json:"monitor,omitempty"`
	SearchForMissingEpisodes     bool   `json:"searchForMissingEpisodes,omitempty"`
	SearchForCutoffUnmetEpisodes bool   `json:"searchForCutoffUnmetEpisodes,omitempty"`
}

type Radarr_AddOptions struct {
	IgnoreEpisodesWithFiles    bool   `json:"ignoreEpisodesWithFiles,omitempty"`
	IgnoreEpisodesWithoutFiles bool   `json:"ignoreEpisodesWithoutFiles,omitempty"`
	Monitor                    string `json:"monitor,omitempty"`
	SearchForMovie             bool   `json:"searchForMovie,omitempty"`
	AddMethod                  string `json:"addMethod,omitempty"`
}

type Radarr_MovieFile struct {
	ID                  int64                 `json:"id,omitempty"`
	MovieID             int64                 `json:"movieId,omitempty"`
	RelativePath        string                `json:"relativePath,omitempty"`
	Path                string                `json:"path,omitempty"`
	Size                int64                 `json:"size,omitempty"`
	DateAdded           time.Time             `json:"dateAdded"`
	SceneName           string                `json:"sceneName,omitempty"`
	ReleaseGroup        string                `json:"releaseGroup,omitempty"`
	Edition             string                `json:"edition,omitempty"`
	Languages           []SR_Language         `json:"languages"`
	Quality             MovieFileQuality      `json:"quality"`
	CustomFormats       []Radarr_CustomFormat `json:"customFormats"`
	CustomFormatScore   int64                 `json:"customFormatScore,omitempty"`
	IndexerFlags        int64                 `json:"indexerFlags,omitempty"`
	MediaInfo           Radarr_MediaInfo      `json:"mediaInfo"`
	OriginalFilePath    string                `json:"originalFilePath,omitempty"`
	QualityCutoffNotMet bool                  `json:"qualityCutoffNotMet,omitempty"`
}

type Radarr_CustomFormat struct {
	ID                              int64                  `json:"id,omitempty"`
	Name                            string                 `json:"name,omitempty"`
	IncludeCustomFormatWhenRenaming bool                   `json:"includeCustomFormatWhenRenaming,omitempty"`
	Specifications                  []Radarr_Specification `json:"specifications,omitempty"`
}

type Radarr_Specification struct {
	ID                 int64                       `json:"id,omitempty"`
	Name               string                      `json:"name,omitempty"`
	Implementation     string                      `json:"implementation,omitempty"`
	ImplementationName string                      `json:"implementationName,omitempty"`
	InfoLink           string                      `json:"infoLink,omitempty"`
	Negate             bool                        `json:"negate,omitempty"`
	Required           bool                        `json:"required,omitempty"`
	Fields             []Radarr_SpecificationField `json:"fields"`
	Presets            []string                    `json:"presets,omitempty"`
}

type Radarr_SpecificationField struct {
	Order                       int64                      `json:"order,omitempty"`
	Name                        string                     `json:"name,omitempty"`
	Label                       string                     `json:"label,omitempty"`
	Unit                        string                     `json:"unit,omitempty"`
	HelpText                    string                     `json:"helpText,omitempty"`
	HelpTextWarning             string                     `json:"helpTextWarning,omitempty"`
	HelpLink                    string                     `json:"helpLink,omitempty"`
	Value                       string                     `json:"value,omitempty"`
	Type                        string                     `json:"type,omitempty"`
	Advanced                    bool                       `json:"advanced,omitempty"`
	SelectOptions               []Radarr_FieldSelectOption `json:"selectOptions"`
	SelectOptionsProviderAction string                     `json:"selectOptionsProviderAction,omitempty"`
	Section                     string                     `json:"section,omitempty"`
	Hidden                      string                     `json:"hidden,omitempty"`
	Privacy                     string                     `json:"privacy,omitempty"`
	Placeholder                 string                     `json:"placeholder,omitempty"`
	IsFloat                     bool                       `json:"isFloat,omitempty"`
}

type Radarr_FieldSelectOption struct {
	Value        int64  `json:"value,omitempty"`
	Name         string `json:"name,omitempty"`
	Order        int64  `json:"order,omitempty"`
	Hint         string `json:"hint,omitempty"`
	DividerAfter bool   `json:"dividerAfter,omitempty"`
}

type MovieFileQuality struct {
	Quality  QualityQuality                   `json:"quality"`
	Revision Radarr_MovieFileRevisionRevision `json:"revision"`
}

type QualityQuality struct {
	ID         int64  `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Source     string `json:"source,omitempty"`
	Resolution int64  `json:"resolution,omitempty"`
	Modifier   string `json:"modifier,omitempty"`
}
type Radarr_MovieFileRevisionRevision struct {
	Version  int64 `json:"version,omitempty"`
	Real     int64 `json:"real,omitempty"`
	IsRepack bool  `json:"isRepack,omitempty"`
}

type Radarr_MediaInfo struct {
	ID                    int64   `json:"id,omitempty"`
	AudioBitrate          int64   `json:"audioBitrate,omitempty"`
	AudioChannels         float64 `json:"audioChannels,omitempty"`
	AudioCodec            string  `json:"audioCodec,omitempty"`
	AudioLanguages        string  `json:"audioLanguages,omitempty"`
	AudioStreamCount      int64   `json:"audioStreamCount,omitempty"`
	VideoBitDepth         int64   `json:"videoBitDepth,omitempty"`
	VideoBitrate          int64   `json:"videoBitrate,omitempty"`
	VideoCodec            string  `json:"videoCodec,omitempty"`
	VideoFPS              float64 `json:"videoFps,omitempty"`
	VideoDynamicRange     string  `json:"videoDynamicRange,omitempty"`
	VideoDynamicRangeType string  `json:"videoDynamicRangeType,omitempty"`
	Resolution            string  `json:"resolution,omitempty"`
	RunTime               string  `json:"runTime,omitempty"`
	ScanType              string  `json:"scanType,omitempty"`
	Subtitles             string  `json:"subtitles,omitempty"`
}
