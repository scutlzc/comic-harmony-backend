package model

import "time"

type DataSourceType string

const (
	SourceKomga      DataSourceType = "komga"
	SourceWebDAV     DataSourceType = "webdav"
	SourceCloudDrive DataSourceType = "clouddrive"
)

type DataSourceConfig struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	Type      DataSourceType `json:"type"`
	URL       string         `json:"url"`
	Username  string         `json:"username,omitempty"`
	Password  string         `json:"-"`
	Token     string         `json:"-"`
	RootPath  string         `json:"root_path,omitempty"`
	Enabled   bool           `json:"enabled"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Library struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Icon    string `json:"icon,omitempty"`
	ComicCount int `json:"comic_count,omitempty"`
}

type SourceComic struct {
	ID          string `json:"id"`
	SourceID    int64  `json:"source_id"`
	Title       string `json:"title"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description,omitempty"`
	CoverURL    string `json:"cover_url,omitempty"`
	Status      string `json:"status,omitempty"` // ONGOING, ENDED
	SeriesCount int    `json:"series_count,omitempty"`
}

type SourceChapter struct {
	ID          string `json:"id"`
	SourceID    int64  `json:"source_id"`
	ComicID     string `json:"comic_id"`
	Title       string `json:"title"`
	SortOrder   int    `json:"sort_order"`
	PageCount   int    `json:"page_count,omitempty"`
}

type SourcePage struct {
	URL       string `json:"url"`
	Number    int    `json:"number"`
	MediaType string `json:"media_type,omitempty"`
}
