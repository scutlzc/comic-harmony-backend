package model

import "time"

type Comic struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	Description  string    `json:"description"`
	CoverURL     string    `json:"cover_url"`
	Status       int16     `json:"status"` // 0=连载, 1=完结
	TotalViews   int64     `json:"total_views"`
	TotalLikes   int64     `json:"total_likes"`
	TotalFavs    int64     `json:"total_favorites"`
	CategoryID   int64     `json:"category_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateComicRequest struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
	CategoryID  int64  `json:"category_id"`
}
