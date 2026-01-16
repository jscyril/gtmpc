package api

import "time"

type Track struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Artist    string        `json:"artist"`
	Album     string        `json:"album"`
	Duration  time.Duration `json:"duration"`
	FilePath  string        `json:"file_path"`
	Genre     string        `json:"genre"`
	Year      int           `json:"year"`
	TrackNum  int           `json:"track_number"`
	CoverArt  []byte        `json:"-"`
	CreatedAt time.Time     `json:"created_at"`
}

type Playlist struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tracks      []Track   `json:"tracks"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
