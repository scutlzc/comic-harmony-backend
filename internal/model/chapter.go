package model

import "time"

type Chapter struct {
	ID        int64     `json:"id"`
	ComicID   int64     `json:"comic_id"`
	Title     string    `json:"title"`
	SortOrder int       `json:"sort_order"`
	PageCount int       `json:"page_count"`
	CreatedAt time.Time `json:"created_at"`
}
