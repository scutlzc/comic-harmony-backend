package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/muyue/comic-harmony-backend/internal/model"
	"github.com/muyue/comic-harmony-backend/internal/repository"
	"github.com/muyue/comic-harmony-backend/internal/response"
)

type FavoriteHandler struct {
	favRepo repository.FavoriteRepository
}

func NewFavoriteHandler(favRepo repository.FavoriteRepository) *FavoriteHandler {
	return &FavoriteHandler{favRepo: favRepo}
}

func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(int64)
	var req struct {
		ComicID int64 `json:"comic_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.ComicID <= 0 {
		response.BadRequest(w, "invalid comic_id")
		return
	}
	if err := h.favRepo.Add(r.Context(), userID, req.ComicID); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Created(w, map[string]int64{"comic_id": req.ComicID})
}

func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(int64)
	comicID, err := strconv.ParseInt(chi.URLParam(r, "comicId"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid comic id")
		return
	}
	if err := h.favRepo.Remove(r.Context(), userID, comicID); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, nil)
}

func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(int64)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	favorites, err := h.favRepo.List(r.Context(), userID, limit, offset)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	total, _ := h.favRepo.Count(r.Context(), userID)

	if favorites == nil {
		favorites = []model.Favorite{}
	}

	response.Success(w, map[string]interface{}{
		"favorites": favorites,
		"total":     total,
		"page":      page,
	})
}

func (h *FavoriteHandler) Check(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(int64)
	comicID, err := strconv.ParseInt(chi.URLParam(r, "comicId"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid comic id")
		return
	}
	favorited, _ := h.favRepo.IsFavorited(r.Context(), userID, comicID)
	response.Success(w, map[string]bool{"favorited": favorited})
}
