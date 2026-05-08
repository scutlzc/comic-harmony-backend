package service

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/muyue/comic-harmony-backend/internal/model"
	"github.com/muyue/comic-harmony-backend/internal/repository"
)

type ComicService interface {
	List(ctx context.Context, page, pageSize int) ([]model.Comic, int, error)
	GetByID(ctx context.Context, id int64) (*model.Comic, error)
	Create(ctx context.Context, req model.CreateComicRequest) (*model.Comic, error)
	Delete(ctx context.Context, id int64) error
}

type comicService struct {
	repo repository.ComicRepository
}

func NewComicService(repo repository.ComicRepository) ComicService {
	return &comicService{repo: repo}
}

func (s *comicService) List(ctx context.Context, page, pageSize int) ([]model.Comic, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := s.repo.Count(ctx)
	if err != nil {
		log.Printf("[service] count error: %v", err)
		return nil, 0, err
	}

	comics, err := s.repo.List(ctx, offset, pageSize)
	if err != nil {
		log.Printf("[service] list error: %v", err)
		return nil, 0, err
	}
	return comics, total, nil
}

func (s *comicService) GetByID(ctx context.Context, id int64) (*model.Comic, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return c, nil
}

func (s *comicService) Create(ctx context.Context, req model.CreateComicRequest) (*model.Comic, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	return s.repo.Create(ctx, req)
}

func (s *comicService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
