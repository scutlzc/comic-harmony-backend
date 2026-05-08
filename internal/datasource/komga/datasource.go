package komga

import (
	"context"
	"fmt"

	"github.com/muyue/comic-harmony-backend/internal/datasource/core"
	ds "github.com/muyue/comic-harmony-backend/internal/datasource/model"
)

type KomgaDataSource struct {
	config ds.DataSourceConfig
	client *Client
}

func NewKomgaDataSource(config ds.DataSourceConfig) (*KomgaDataSource, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("komga URL is required")
	}
	client := NewClient(config.URL, config.Username, config.Password)
	return &KomgaDataSource{
		config: config,
		client: client,
	}, nil
}

func (k *KomgaDataSource) Type() ds.DataSourceType {
	return ds.SourceKomga
}

func (k *KomgaDataSource) Config() ds.DataSourceConfig {
	return k.config
}

func (k *KomgaDataSource) HealthCheck(ctx context.Context) error {
	_, err := k.client.GetLibraries(ctx)
	return err
}

func (k *KomgaDataSource) GetLibraries(ctx context.Context) ([]ds.Library, error) {
	komgaLibs, err := k.client.GetLibraries(ctx)
	if err != nil {
		return nil, err
	}

	var libraries []ds.Library
	for _, l := range komgaLibs {
		libraries = append(libraries, ds.Library{
			ID:   l.ID,
			Name: l.Name,
		})
	}
	return libraries, nil
}

func (k *KomgaDataSource) GetComics(ctx context.Context, libraryID string, page, pageSize int) (*core.PaginatedComics, error) {
	series, total, err := k.client.GetSeries(ctx, libraryID, page-1, pageSize)
	if err != nil {
		return nil, err
	}

	var comics []ds.SourceComic
	for _, s := range series {
		comics = append(comics, ds.SourceComic{
			ID:          s.ID,
			SourceID:    k.config.ID,
			Title:       s.Metadata.Title,
			Author:      ExtractAuthor(&s),
			Description: s.Metadata.Summary,
			Status:      s.Metadata.Status,
			SeriesCount: s.BooksCount,
		})
	}

	if comics == nil {
		comics = []ds.SourceComic{}
	}

	return &core.PaginatedComics{
		Comics:   comics,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  (page * pageSize) < total,
	}, nil
}

func (k *KomgaDataSource) SearchComics(ctx context.Context, query string, page, pageSize int) (*core.PaginatedComics, error) {
	series, total, err := k.client.SearchSeries(ctx, query, page-1, pageSize)
	if err != nil {
		return nil, err
	}

	var comics []ds.SourceComic
	for _, s := range series {
		comics = append(comics, ds.SourceComic{
			ID:          s.ID,
			SourceID:    k.config.ID,
			Title:       s.Metadata.Title,
			Author:      ExtractAuthor(&s),
			Description: s.Metadata.Summary,
		})
	}

	if comics == nil {
		comics = []ds.SourceComic{}
	}

	return &core.PaginatedComics{
		Comics:   comics,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  (page * pageSize) < total,
	}, nil
}

func (k *KomgaDataSource) GetChapters(ctx context.Context, seriesID string) ([]ds.SourceChapter, error) {
	books, err := k.client.GetBooks(ctx, seriesID)
	if err != nil {
		return nil, err
	}

	var chapters []ds.SourceChapter
	for _, b := range books {
		chapters = append(chapters, ds.SourceChapter{
			ID:        b.ID,
			ComicID:   b.SeriesID,
			Title:     b.Name,
			SortOrder: int(b.Number),
			PageCount: b.PageCount,
		})
	}
	return chapters, nil
}

func (k *KomgaDataSource) GetPageURLs(ctx context.Context, bookID string) ([]ds.SourcePage, error) {
	// Komga returns pages sequentially numbered from 0
	// We need page count to generate URLs; return first page to let caller discover
	pages := []ds.SourcePage{
		{
			URL:    k.client.GetPageURL(bookID, 0),
			Number: 0,
		},
	}
	return pages, nil
}

func (k *KomgaDataSource) GetCoverURL(ctx context.Context, seriesID string) (string, error) {
	return fmt.Sprintf("%s/api/v1/series/%s/thumbnail", k.config.URL, seriesID), nil
}
