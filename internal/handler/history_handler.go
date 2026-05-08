package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/muyue/comic-harmony-backend/internal/model"
	"github.com/muyue/comic-harmony-backend/internal/repository"
	"github.com/muyue/comic-harmony-backend/internal/response"
)

type HistoryHandler struct {
	histRepo repository.HistoryRepository
}

func NewHistoryHandler(histRepo repository.HistoryRepository) *HistoryHandler {
	return &HistoryHandler{histRepo: histRepo}
}

func (h *HistoryHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(int64)
	var req model.ReadingHistory
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	req.UserID = userID

	if err := h.histRepo.Upsert(r.Context(), &req); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, req)
}

func (h *HistoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(int64)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 30
	offset := (page - 1) * limit

	history, err := h.histRepo.List(r.Context(), userID, limit, offset)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	if history == nil {
		history = []model.ReadingHistory{}
	}
	response.Success(w, map[string]interface{}{
		"history": history,
		"page":    page,
	})
}

func (h *HistoryHandler) Clear(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(int64)
	if err := h.histRepo.Clear(r.Context(), userID); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, nil)
}
