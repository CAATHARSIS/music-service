package models

import "time"

type File struct {
	ID           string    `db:"id"`
	OriginalName string    `db:"original_name"`
	Bucket       string    `db:"bucket"`
	Key          string    `db:"key"`
	Size         int64     `db:"size"`
	MimeType     string    `db:"mime_type"`
	UploadedBy   *string   `db:"uploaded_by"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
