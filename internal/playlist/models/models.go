package models

import "time"

type Playlist struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	IsPublic    bool      `db:"is_public"`
	Type        string    `db:"type"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Tracks      []PlaylistTrack
}

type PlaylistTrack struct {
	ID         string    `db:"id"`
	PlaylistID string    `db:"playlist_id"`
	TrackID    string    `db:"track_id"`
	Position   int       `db:"position"`
	AddedAt    time.Time `db:"added_at"`
}
