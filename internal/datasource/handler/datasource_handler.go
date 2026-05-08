package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/muyue/comic-harmony-backend/internal/datasource/clouddrive"
	"github.com/muyue/comic-harmony-backend/internal/datasource/core"
	dsmodel "github.com/muyue/comic-harmony-backend/internal/datasource/model"
	"github.com/muyue/comic-harmony-backend/internal/datasource/komga"
	"github.com/muyue/comic-harmony-backend/internal/datasource/webdav"
	uh "github.com/muyue/comic-harmony-backend/internal/handler"
	"github.com/muyue/comic-harmony-backend/internal/model"
	"github.com/muyue/comic-harmony-backend/internal/repository"
	"github.com/muyue/comic-harmony-backend/internal/response"
)

type DataSourceHandler struct {
	manager *core.DataSourceManager
	repo    repository.DataSourceRepository
}

func NewDataSourceHandler(manager *core.DataSourceManager, repo repository.DataSourceRepository) *DataSourceHandler {
	return &DataSourceHandler{manager: manager, repo: repo}
}

type addSourceRequest struct {
	Name     string            `json:"name"`
	Type     dsmodel.DataSourceType `json:"type"`
	URL      string            `json:"url"`
	Username string            `json:"username,omitempty"`
	Password string            `json:"password,omitempty"`
}

func (h *DataSourceHandler) getUserID(r *http.Request) int64 {
	if uid, ok := r.Context().Value(uh.UserIDKey).(int64); ok {
		return uid
	}
	return 0
}

func (h *DataSourceHandler) Add(w http.ResponseWriter, r *http.Request) {
	var req addSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	// Create DB record
	dbDS := model.DataSource{
		UserID:     userID,
		Name:       req.Name,
		SourceType: string(req.Type),
		URL:        req.URL,
		Username:   req.Username,
		Password:   req.Password,
		RootPath:   "/",
		Enabled:    true,
	}

	if err := h.repo.Create(r.Context(), &dbDS); err != nil {
		response.InternalError(w, "failed to save data source: "+err.Error())
		return
	}

	// Add to runtime manager
	cfg := dsmodel.DataSourceConfig{
		ID:       dbDS.ID,
		Name:     dbDS.Name,
		Type:     dsmodel.DataSourceType(dbDS.SourceType),
		URL:      dbDS.URL,
		Username: dbDS.Username,
		Password: dbDS.Password,
		RootPath: dbDS.RootPath,
		Enabled:  dbDS.Enabled,
	}

	source, err := createSource(cfg)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if _, err := h.manager.Add(cfg); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	// Background health check
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := source.HealthCheck(ctx); err != nil {
			dbDS.Enabled = false
			h.repo.Update(context.Background(), &dbDS)
		}
	}()

	response.Created(w, dbDS)
}

func (h *DataSourceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	sources, err := h.repo.ListByUser(r.Context(), userID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	if sources == nil {
		sources = []model.DataSource{}
	}
	response.Success(w, sources)
}

func (h *DataSourceHandler) Remove(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}
	// Remove from runtime
	h.manager.Remove(id)
	// Remove from DB
	if err := h.repo.Delete(r.Context(), id); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, nil)
}

func (h *DataSourceHandler) GetLibraries(w http.ResponseWriter, r *http.Request) {
	source, ok := h.getSource(w, r)
	if !ok {
		return
	}
	libraries, err := source.GetLibraries(r.Context())
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, libraries)
}

func (h *DataSourceHandler) GetComics(w http.ResponseWriter, r *http.Request) {
	source, ok := h.getSource(w, r)
	if !ok {
		return
	}
	libraryID := r.URL.Query().Get("library_id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	result, err := source.GetComics(r.Context(), libraryID, page, pageSize)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, result)
}

func (h *DataSourceHandler) GetChapters(w http.ResponseWriter, r *http.Request) {
	source, ok := h.getSource(w, r)
	if !ok {
		return
	}
	comicID := r.URL.Query().Get("comic_id")
	if comicID == "" {
		response.BadRequest(w, "comic_id is required")
		return
	}
	chapters, err := source.GetChapters(r.Context(), comicID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, chapters)
}

func (h *DataSourceHandler) Search(w http.ResponseWriter, r *http.Request) {
	source, ok := h.getSource(w, r)
	if !ok {
		return
	}
	query := r.URL.Query().Get("q")
	if query == "" {
		response.BadRequest(w, "query is required")
		return
	}
	result, err := source.SearchComics(r.Context(), query, 1, 20)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, result)
}

func (h *DataSourceHandler) getSource(w http.ResponseWriter, r *http.Request) (core.IDataSource, bool) {
	idStr := chi.URLParam(r, "sourceId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid source id")
		return nil, false
	}
	source, ok := h.manager.Get(id)
	if !ok {
		response.NotFound(w, "source not found")
		return nil, false
	}
	return source, true
}

// LoadFromDB loads all enabled data sources from DB into the runtime manager.
// Called at server startup.
func (h *DataSourceHandler) LoadFromDB(ctx context.Context) error {
	sources, err := h.repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("load data sources: %w", err)
	}
	for _, ds := range sources {
		if !ds.Enabled {
			continue
		}
		cfg := dsmodel.DataSourceConfig{
			ID:       ds.ID,
			Name:     ds.Name,
			Type:     dsmodel.DataSourceType(ds.SourceType),
			URL:      ds.URL,
			Username: ds.Username,
			Password: ds.Password,
			RootPath: ds.RootPath,
			Enabled:  ds.Enabled,
		}
		if _, err := h.manager.Add(cfg); err != nil {
			fmt.Printf("[datasource] failed to load %s (id=%d): %v\n", ds.Name, ds.ID, err)
		}
	}
	fmt.Printf("[datasource] loaded %d data sources\n", len(sources))
	return nil
}

func createSource(cfg dsmodel.DataSourceConfig) (core.IDataSource, error) {
	switch cfg.Type {
	case dsmodel.SourceKomga:
		return komga.NewKomgaDataSource(cfg)
	case dsmodel.SourceWebDAV:
		return webdav.NewWebDAVDataSource(cfg)
	case dsmodel.SourceCloudDrive:
		return clouddrive.NewCloudDriveDataSource(cfg)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", cfg.Type)
	}
}
