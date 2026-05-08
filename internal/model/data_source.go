package model

import "time"

type DataSource struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id,omitempty"`
	Name       string     `json:"name"`
	SourceType string     `json:"source_type"`
	URL        string     `json:"url"`
	Username   string     `json:"username,omitempty"`
	Password   string     `json:"-"`
	RootPath   string     `json:"root_path"`
	Enabled    bool       `json:"enabled"`
	LastHealth *time.Time `json:"last_health,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
