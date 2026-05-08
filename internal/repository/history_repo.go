package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muyue/comic-harmony-backend/internal/model"
)

type HistoryRepository interface {
	Upsert(ctx context.Context, h *model.ReadingHistory) error
	List(ctx context.Context, userID int64, limit, offset int) ([]model.ReadingHistory, error)
	Clear(ctx context.Context, userID int64) error
}

type historyRepo struct {
	pool *pgxpool.Pool
}

func NewHistoryRepository(pool *pgxpool.Pool) HistoryRepository {
	return &historyRepo{pool: pool}
}

func (r *historyRepo) Upsert(ctx context.Context, h *model.ReadingHistory) error {
	now := time.Now()
	err := r.pool.QueryRow(ctx,
		`INSERT INTO reading_history (user_id, comic_id, chapter_id, chapter_title, page, total_pages, progress, read_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (user_id, comic_id) DO UPDATE SET
		   chapter_id   = EXCLUDED.chapter_id,
		   chapter_title = EXCLUDED.chapter_title,
		   page         = EXCLUDED.page,
		   total_pages  = EXCLUDED.total_pages,
		   progress     = EXCLUDED.progress,
		   read_at      = EXCLUDED.read_at
		 RETURNING id`,
		h.UserID, h.ComicID, h.ChapterID, h.ChapterTitle,
		h.Page, h.TotalPages, h.Progress, now,
	).Scan(&h.ID)
	if err != nil {
		return err
	}
	h.ReadAt = now
	return nil
}

func (r *historyRepo) List(ctx context.Context, userID int64, limit, offset int) ([]model.ReadingHistory, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT h.id, h.user_id, h.comic_id, h.chapter_id, h.chapter_title,
		        h.page, h.total_pages, h.progress, h.read_at,
		        c.title, c.cover_url
		 FROM reading_history h
		 JOIN comics c ON c.id = h.comic_id
		 WHERE h.user_id = $1
		 ORDER BY h.read_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []model.ReadingHistory
	for rows.Next() {
		var h model.ReadingHistory
		if err := rows.Scan(&h.ID, &h.UserID, &h.ComicID, &h.ChapterID, &h.ChapterTitle,
			&h.Page, &h.TotalPages, &h.Progress, &h.ReadAt,
			&h.ComicTitle, &h.ComicCover); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func (r *historyRepo) Clear(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM reading_history WHERE user_id = $1`, userID)
	return err
}
