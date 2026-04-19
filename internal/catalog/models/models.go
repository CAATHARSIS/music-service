package models

import "time"

type Track struct {
	ID          string    `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Duration    int       `db:"duration" json:"duration"`
	Year        int       `db:"year" json:"year"`
	FileID      string    `db:"file_id" json:"file_id"`
	CoverImagID *string   `db:"cover_image_id" json:"cover_image_id,omitempty"`
	TrackNumber *int      `db:"track_number" json:"track_number,omitempty"`
	Lyrics      *string   `db:"lyrics" json:"lyrics,omitempty"`
	PlaysCount  int       `db:"plays_count" json:"plays_count"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`

	Artist *Artist  `db:"artist" json:"artist"`
	Album  *Album   `db:"album" json:"album"`
	Genres []*Genre `json:"genres,omitempty"`

	ArtistID string  `db:"artist_id" json:"-"`
	AlbumID  *string `db:"album_id" json:"-"`
}

type Artist struct {
	ID            string    `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Country       *string   `db:"country" json:"country,omitempty"`
	AvatarImageID *string   `db:"avatar_image_id" json:"avatar_image_id,omitempty"`
	TotalPlays    int64     `db:"total_plays" json:"total_plays"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`

	Genres []*Genre `json:"genres,omitempty"`
}

type Album struct {
	ID           string  `db:"id" json:"id"`
	Title        string  `db:"title" json:"title"`
	Year         int32   `db:"year" json:"year"`
	AlbumType    string  `db:"album_type" json:"album_type"`
	CoverImageID *string `db:"cover_image_id" json:"cover_image_id"`

	Artist *Artist  `db:"artist" json:"artist,omitempty"`
	Genres []*Genre `json:"genres,omitempty"`
}

type Genre struct {
	ID          string  `db:"id" json:"id"`
	Name        string  `db:"name" json:"name"`
	Description *string `db:"description" json:"description"`
}
