package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muyue/comic-harmony-backend/internal/model"
)

type ChapterRepository interface {
	Create(ctx context.Context, chapter *model.Chapter) error
	ListByComic(ctx context.Context, comicID int64) ([]model.Chapter, error)
	GetByID(ctx context.Context, id int64) (*model.Chapter, error)
}

type chapterRepo struct {
	pool *pgxpool.Pool
}

func NewChapterRepository(pool *pgxpool.Pool) ChapterRepository {
	return &chapterRepo{pool: pool}
}

func (r *chapterRepo) Create(ctx context.Context, chapter *model.Chapter) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO chapters (comic_id, title, sort_order, page_count)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		chapter.ComicID, chapter.Title, chapter.SortOrder, chapter.PageCount,
	).Scan(&chapter.ID, &chapter.CreatedAt)
}

func (r *chapterRepo) ListByComic(ctx context.Context, comicID int64) ([]model.Chapter, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, comic_id, title, sort_order, page_count, created_at
		 FROM chapters WHERE comic_id = $1 ORDER BY sort_order`, comicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chapters []model.Chapter
	for rows.Next() {
		var c model.Chapter
		if err := rows.Scan(&c.ID, &c.ComicID, &c.Title, &c.SortOrder, &c.PageCount, &c.CreatedAt); err != nil {
			return nil, err
		}
		chapters = append(chapters, c)
	}
	if chapters == nil {
		chapters = []model.Chapter{}
	}
	return chapters, nil
}

func (r *chapterRepo) GetByID(ctx context.Context, id int64) (*model.Chapter, error) {
	c := &model.Chapter{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, comic_id, title, sort_order, page_count, created_at
		 FROM chapters WHERE id = $1`, id,
	).Scan(&c.ID, &c.ComicID, &c.Title, &c.SortOrder, &c.PageCount, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}
