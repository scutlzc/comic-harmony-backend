package core

import (
	"context"

	ds "github.com/muyue/comic-harmony-backend/internal/datasource/model"
)

// IDataSource defines the interface all comic data sources must implement.
type IDataSource interface {
	// Type returns the data source type identifier
	Type() ds.DataSourceType

	// Config returns the current configuration
	Config() ds.DataSourceConfig

	// HealthCheck verifies the connection to the source
	HealthCheck(ctx context.Context) error

	// GetLibraries returns top-level libraries/collections
	GetLibraries(ctx context.Context) ([]ds.Library, error)

	// GetComics returns comics from a library with pagination
	GetComics(ctx context.Context, libraryID string, page, pageSize int) (*PaginatedComics, error)

	// SearchComics searches comics across the source
	SearchComics(ctx context.Context, query string, page, pageSize int) (*PaginatedComics, error)

	// GetChapters returns chapters for a comic
	GetChapters(ctx context.Context, comicID string) ([]ds.SourceChapter, error)

	// GetPageURLs returns page image URLs for a chapter
	GetPageURLs(ctx context.Context, chapterID string) ([]ds.SourcePage, error)

	// GetCoverURL returns the cover image URL for a comic
	GetCoverURL(ctx context.Context, comicID string) (string, error)
}

type PaginatedComics struct {
	Comics     []ds.SourceComic `json:"comics"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	HasMore    bool             `json:"has_more"`
}
