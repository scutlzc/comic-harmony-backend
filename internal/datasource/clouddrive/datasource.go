package clouddrive

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/muyue/comic-harmony-backend/internal/datasource/core"
	ds "github.com/muyue/comic-harmony-backend/internal/datasource/model"
)

type CloudDriveDataSource struct {
	config ds.DataSourceConfig
	client *http.Client
}

func NewCloudDriveDataSource(config ds.DataSourceConfig) (*CloudDriveDataSource, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("cloud drive URL is required")
	}
	return &CloudDriveDataSource{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *CloudDriveDataSource) Type() ds.DataSourceType         { return ds.SourceCloudDrive }
func (c *CloudDriveDataSource) Config() ds.DataSourceConfig      { return c.config }

func (c *CloudDriveDataSource) HealthCheck(ctx context.Context) error {
	return nil // Will be implemented per cloud provider
}

func (c *CloudDriveDataSource) GetLibraries(ctx context.Context) ([]ds.Library, error) {
	return []ds.Library{
		{ID: "/", Name: "根目录"},
	}, nil
}

func (c *CloudDriveDataSource) GetComics(ctx context.Context, libraryID string, page, pageSize int) (*core.PaginatedComics, error) {
	return &core.PaginatedComics{
		Comics: []ds.SourceComic{}, Total: 0, Page: page, PageSize: pageSize,
	}, nil
}

func (c *CloudDriveDataSource) SearchComics(ctx context.Context, query string, page, pageSize int) (*core.PaginatedComics, error) {
	return &core.PaginatedComics{
		Comics: []ds.SourceComic{}, Total: 0, Page: page, PageSize: pageSize,
	}, nil
}

func (c *CloudDriveDataSource) GetChapters(ctx context.Context, comicID string) ([]ds.SourceChapter, error) {
	return []ds.SourceChapter{}, nil
}

func (c *CloudDriveDataSource) GetPageURLs(ctx context.Context, chapterID string) ([]ds.SourcePage, error) {
	return []ds.SourcePage{}, nil
}

func (c *CloudDriveDataSource) GetCoverURL(ctx context.Context, comicID string) (string, error) {
	return "", nil
}

// Cloud drive providers
type Provider string
const (
	AliyunDrive  Provider = "aliyundrive"
	BaiduPan     Provider = "baidupan"
	OneDrive     Provider = "onedrive"
	GoogleDrive  Provider = "googledrive"
)

var _ core.IDataSource = (*CloudDriveDataSource)(nil)
