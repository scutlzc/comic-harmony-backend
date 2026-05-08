package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muyue/comic-harmony-backend/internal/model"
)

type FavoriteRepository interface {
	Add(ctx context.Context, userID, comicID int64) error
	Remove(ctx context.Context, userID, comicID int64) error
	List(ctx context.Context, userID int64, limit, offset int) ([]model.Favorite, error)
	Count(ctx context.Context, userID int64) (int, error)
	IsFavorited(ctx context.Context, userID, comicID int64) (bool, error)
}

type favoriteRepo struct {
	pool *pgxpool.Pool
}

func NewFavoriteRepository(pool *pgxpool.Pool) FavoriteRepository {
	return &favoriteRepo{pool: pool}
}

func (r *favoriteRepo) Add(ctx context.Context, userID, comicID int64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO favorites (user_id, comic_id) VALUES ($1, $2)
		 ON CONFLICT (user_id, comic_id) DO NOTHING`,
		userID, comicID)
	return err
}

func (r *favoriteRepo) Remove(ctx context.Context, userID, comicID int64) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM favorites WHERE user_id = $1 AND comic_id = $2`,
		userID, comicID)
	return err
}

func (r *favoriteRepo) List(ctx context.Context, userID int64, limit, offset int) ([]model.Favorite, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT f.id, f.user_id, f.comic_id, f.created_at,
		        c.title, c.cover_url
		 FROM favorites f
		 JOIN comics c ON c.id = f.comic_id
		 WHERE f.user_id = $1
		 ORDER BY f.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []model.Favorite
	for rows.Next() {
		var f model.Favorite
		if err := rows.Scan(&f.ID, &f.UserID, &f.ComicID, &f.CreatedAt,
			&f.ComicTitle, &f.ComicCover); err != nil {
			return nil, err
		}
		favorites = append(favorites, f)
	}
	return favorites, nil
}

func (r *favoriteRepo) Count(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM favorites WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func (r *favoriteRepo) IsFavorited(ctx context.Context, userID, comicID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND comic_id = $2)`,
		userID, comicID).Scan(&exists)
	return exists, err
}
