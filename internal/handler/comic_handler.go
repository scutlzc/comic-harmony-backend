package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/muyue/comic-harmony-backend/internal/model"
	"github.com/muyue/comic-harmony-backend/internal/response"
	"github.com/muyue/comic-harmony-backend/internal/service"
)

type ComicHandler struct {
	svc service.ComicService
}

func NewComicHandler(svc service.ComicService) *ComicHandler {
	return &ComicHandler{svc: svc}
}

func (h *ComicHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	comics, total, err := h.svc.List(r.Context(), page, pageSize)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"comics": comics,
			"total":  total,
			"page":   page,
			"page_size": pageSize,
		},
	})
}

func (h *ComicHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	comic, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	if comic == nil {
		response.NotFound(w, "comic not found")
		return
	}

	response.Success(w, comic)
}

func (h *ComicHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateComicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	comic, err := h.svc.Create(r.Context(), req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, comic)
}

func (h *ComicHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, response.APIResponse{
		Code:    200,
		Message: "deleted",
	})
}
