package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/muyue/comic-harmony-backend/internal/repository"
	"github.com/muyue/comic-harmony-backend/internal/response"
)

type ChapterHandler struct {
	chapterRepo repository.ChapterRepository
	comicRepo   repository.ComicRepository
}

func NewChapterHandler(chapterRepo repository.ChapterRepository, comicRepo repository.ComicRepository) *ChapterHandler {
	return &ChapterHandler{chapterRepo: chapterRepo, comicRepo: comicRepo}
}

func (h *ChapterHandler) ListByComic(w http.ResponseWriter, r *http.Request) {
	comicID, err := strconv.ParseInt(chi.URLParam(r, "comicId"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid comic id")
		return
	}

	// Verify comic exists
	comic, err := h.comicRepo.GetByID(r.Context(), comicID)
	if err != nil || comic == nil {
		response.NotFound(w, "comic not found")
		return
	}

	chapters, err := h.chapterRepo.ListByComic(r.Context(), comicID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, map[string]interface{}{
		"comic_id": comicID,
		"chapters": chapters,
		"total":    len(chapters),
	})
}

func (h *ChapterHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid chapter id")
		return
	}

	chapter, err := h.chapterRepo.GetByID(r.Context(), id)
	if err != nil || chapter == nil {
		response.NotFound(w, "chapter not found")
		return
	}

	// Return chapter with page URLs
	response.Success(w, map[string]interface{}{
		"chapter": chapter,
		"pages":   chapter.PageCount,
	})
}
