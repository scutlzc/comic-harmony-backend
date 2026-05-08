package webdav

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/muyue/comic-harmony-backend/internal/datasource/core"
	ds "github.com/muyue/comic-harmony-backend/internal/datasource/model"
)

type WebDAVDataSource struct {
	config  ds.DataSourceConfig
	client  *http.Client
	rootURL string
}

func NewWebDAVDataSource(config ds.DataSourceConfig) (*WebDAVDataSource, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("webdav URL is required")
	}
	base := strings.TrimRight(config.URL, "/")
	return &WebDAVDataSource{
		config:  config,
		client: &http.Client{Timeout: 30 * time.Second},
		rootURL: base,
	}, nil
}

func (w *WebDAVDataSource) Type() ds.DataSourceType { return ds.SourceWebDAV }
func (w *WebDAVDataSource) Config() ds.DataSourceConfig { return w.config }

func (w *WebDAVDataSource) doPropfind(ctx context.Context, davPath string, depth int) ([]WebDAVResource, error) {
	u, _ := url.JoinPath(w.rootURL, davPath)
	req, err := http.NewRequestWithContext(ctx, "PROPFIND", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Depth", fmt.Sprintf("%d", depth))
	if w.config.Username != "" {
		req.SetBasicAuth(w.config.Username, w.config.Password)
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("webdav propfind: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 207 {
		return nil, fmt.Errorf("webdav status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseMultiStatus(body)
}

func (w *WebDAVDataSource) HealthCheck(ctx context.Context) error {
	_, err := w.doPropfind(ctx, "/", 0)
	return err
}

func (w *WebDAVDataSource) GetLibraries(ctx context.Context) ([]ds.Library, error) {
	dirs, err := w.doPropfind(ctx, "/", 1)
	if err != nil {
		return nil, err
	}
	var libraries []ds.Library
	for _, d := range dirs {
		if d.IsCollection && d.Path != "/" {
			libraries = append(libraries, ds.Library{
				ID:   d.Path,
				Name: strings.Trim(d.Name, "/"),
			})
		}
	}
	return libraries, nil
}

func (w *WebDAVDataSource) GetComics(ctx context.Context, dirPath string, page, pageSize int) (*core.PaginatedComics, error) {
	resources, err := w.doPropfind(ctx, dirPath, 1)
	if err != nil {
		return nil, err
	}

	var comics []ds.SourceComic
	for _, r := range resources {
		if r.IsCollection || !isSupportedComicExt(r.Name) {
			continue
		}
		comics = append(comics, ds.SourceComic{
			ID:       path.Join(dirPath, r.Name),
			SourceID: w.config.ID,
			Title:    strings.TrimSuffix(r.Name, path.Ext(r.Name)),
		})
	}

	// Paginate
	total := len(comics)
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	if start > total {
		comics = []ds.SourceComic{}
	} else {
		comics = comics[start:end]
	}

	return &core.PaginatedComics{
		Comics:   comics,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  end < total,
	}, nil
}

func (w *WebDAVDataSource) SearchComics(ctx context.Context, query string, page, pageSize int) (*core.PaginatedComics, error) {
	resources, err := w.doPropfind(ctx, "/", -1)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var comics []ds.SourceComic
	for _, r := range resources {
		if r.IsCollection || !isSupportedComicExt(r.Name) {
			continue
		}
		if strings.Contains(strings.ToLower(r.Name), q) {
			comics = append(comics, ds.SourceComic{
				ID:       path.Join(r.Path, r.Name),
				SourceID: w.config.ID,
				Title:    strings.TrimSuffix(r.Name, path.Ext(r.Name)),
			})
		}
	}
	total := len(comics)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		comics = []ds.SourceComic{}
	} else if end > total {
		comics = comics[start:]
	} else {
		comics = comics[start:end]
	}
	return &core.PaginatedComics{
		Comics: comics, Total: total, Page: page, PageSize: pageSize, HasMore: end < total,
	}, nil
}

func (w *WebDAVDataSource) GetChapters(ctx context.Context, comicID string) ([]ds.SourceChapter, error) {
	// For single file comics, return one chapter
	if !strings.HasSuffix(comicID, "/") {
		ext := strings.ToLower(path.Ext(comicID))
		if ext == ".cbz" || ext == ".cbr" || ext == ".zip" || ext == ".rar" {
			return []ds.SourceChapter{{
				ID: comicID, ComicID: comicID,
				Title: strings.TrimSuffix(path.Base(comicID), ext),
				SortOrder: 1,
			}}, nil
		}
	}

	// For directories, check subdirectories as chapters
	resources, err := w.doPropfind(ctx, comicID, 1)
	if err != nil {
		return nil, err
	}
	var chapters []ds.SourceChapter
	for _, r := range resources {
		if r.Path == comicID {
			continue
		}
		if r.IsCollection || isSupportedComicExt(r.Name) {
			chapters = append(chapters, ds.SourceChapter{
				ID: r.Path, ComicID: comicID,
				Title: strings.TrimSuffix(r.Name, path.Ext(r.Name)),
				SortOrder: len(chapters) + 1,
			})
		}
	}
	return chapters, nil
}

func (w *WebDAVDataSource) GetPageURLs(ctx context.Context, chapterID string) ([]ds.SourcePage, error) {
	// If it's an archive, we need to proxy through backend
	// For now, return single page pointing to the file URL
	u, _ := url.JoinPath(w.rootURL, chapterID)
	return []ds.SourcePage{{
		URL:    u,
		Number: 0,
	}}, nil
}

func (w *WebDAVDataSource) GetCoverURL(ctx context.Context, comicID string) (string, error) {
	dir := path.Dir(comicID)
	resources, err := w.doPropfind(ctx, dir, 1)
	if err != nil {
		return "", err
	}
	for _, r := range resources {
		if !r.IsCollection && isImageFile(r.Name) {
			u, _ := url.JoinPath(w.rootURL, path.Join(dir, r.Name))
			return u, nil
		}
	}
	return "", nil
}

// --- WebDAV XML parsing ---

type WebDAVResource struct {
	Path         string
	Name         string
	IsCollection bool
	Size         int64
}

type multistatus struct {
	XMLName xml.Name `xml:"multistatus"`
	Responses []response `xml:"response"`
}

type response struct {
	Href string `xml:"href"`
	Propstats []propstat `xml:"propstat"`
}

type propstat struct {
	Props props `xml:"prop"`
}

type props struct {
	DisplayName string `xml:"displayname"`
	ResType     *struct{} `xml:"resourcetype>collection"`
	Size        int64     `xml:"getcontentlength"`
}

func parseMultiStatus(data []byte) ([]WebDAVResource, error) {
	var ms multistatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return nil, err
	}
	var resources []WebDAVResource
	seen := make(map[string]bool)
	for _, r := range ms.Responses {
		p := r.Href
		p = strings.TrimRight(p, "/")
		if seen[p] {
			continue
		}
		seen[p] = true
		name := path.Base(p)
		if name == "" {
			name = "/"
		}
		isColl := false
		for _, ps := range r.Propstats {
			if ps.Props.ResType != nil {
				isColl = true
				break
			}
		}
		resources = append(resources, WebDAVResource{
			Path: p, Name: name, IsCollection: isColl,
		})
	}
	sort.Slice(resources, func(i, j int) bool {
		if resources[i].IsCollection != resources[j].IsCollection {
			return resources[i].IsCollection
		}
		return resources[i].Name < resources[j].Name
	})
	return resources, nil
}

var comicExts = map[string]bool{
	".cbz": true, ".cbr": true, ".cb7": true, ".zip": true,
	".rar": true, ".epub": true, ".pdf": true,
}

func isSupportedComicExt(name string) bool {
	return comicExts[strings.ToLower(path.Ext(name))]
}

func isImageFile(name string) bool {
	ext := strings.ToLower(path.Ext(name))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}

// Register factory
var _ core.IDataSource = (*WebDAVDataSource)(nil)
