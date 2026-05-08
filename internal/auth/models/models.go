package models

import "time"

type User struct {
	ID           string    `db:"id"`
	UserName     string    `db:"username"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	AvatarURL    *string   `db:"avatar_url"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type RefreshToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
