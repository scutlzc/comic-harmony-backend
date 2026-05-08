package model

import "time"

type Favorite struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"-"`
	ComicID   int64     `json:"comic_id"`
	CreatedAt time.Time `json:"created_at"`
	// Joined fields
	ComicTitle    string `json:"comic_title,omitempty"`
	ComicCover    string `json:"comic_cover,omitempty"`
	LatestChapter string `json:"latest_chapter,omitempty"`
}

type ReadingHistory struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"-"`
	ComicID      int64     `json:"comic_id"`
	ChapterID    int64     `json:"chapter_id"`
	ChapterTitle string    `json:"chapter_title"`
	Page         int       `json:"page"`
	TotalPages   int       `json:"total_pages"`
	Progress     float32   `json:"progress"`
	ReadAt       time.Time `json:"read_at"`
	// Joined fields
	ComicTitle string `json:"comic_title,omitempty"`
	ComicCover string `json:"comic_cover,omitempty"`
}
