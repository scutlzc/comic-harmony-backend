package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muyue/comic-harmony-backend/internal/model"
)

type ComicRepository interface {
	List(ctx context.Context, offset, limit int) ([]model.Comic, error)
	GetByID(ctx context.Context, id int64) (*model.Comic, error)
	Create(ctx context.Context, req model.CreateComicRequest) (*model.Comic, error)
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int, error)
}

type comicRepo struct {
	pool *pgxpool.Pool
}

func NewComicRepository(pool *pgxpool.Pool) ComicRepository {
	return &comicRepo{pool: pool}
}

func (r *comicRepo) List(ctx context.Context, offset, limit int) ([]model.Comic, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, author, description, cover_url, status,
		        total_views, total_likes, total_favorites, category_id,
		        created_at, updated_at
		 FROM comics ORDER BY updated_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comics []model.Comic
	for rows.Next() {
		var c model.Comic
		if err := rows.Scan(&c.ID, &c.Title, &c.Author, &c.Description, &c.CoverURL,
			&c.Status, &c.TotalViews, &c.TotalLikes, &c.TotalFavs, &c.CategoryID,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comics = append(comics, c)
	}
	if comics == nil {
		comics = []model.Comic{}
	}
	return comics, nil
}

func (r *comicRepo) GetByID(ctx context.Context, id int64) (*model.Comic, error) {
	c := &model.Comic{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, author, description, cover_url, status,
		        total_views, total_likes, total_favorites, category_id,
		        created_at, updated_at
		 FROM comics WHERE id = $1`, id).Scan(
		&c.ID, &c.Title, &c.Author, &c.Description, &c.CoverURL,
		&c.Status, &c.TotalViews, &c.TotalLikes, &c.TotalFavs, &c.CategoryID,
		&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *comicRepo) Create(ctx context.Context, req model.CreateComicRequest) (*model.Comic, error) {
	c := &model.Comic{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO comics (title, author, description, cover_url, category_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, title, author, description, cover_url, status,
		           total_views, total_likes, total_favorites, category_id,
		           created_at, updated_at`,
		req.Title, req.Author, req.Description, req.CoverURL, req.CategoryID).Scan(
		&c.ID, &c.Title, &c.Author, &c.Description, &c.CoverURL,
		&c.Status, &c.TotalViews, &c.TotalLikes, &c.TotalFavs, &c.CategoryID,
		&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
		c.UpdatedAt = time.Now()
	}
	return c, nil
}

func (r *comicRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM comics WHERE id = $1`, id)
	return err
}

func (r *comicRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM comics`).Scan(&count)
	return count, err
}
