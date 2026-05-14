package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/CAATHARSIS/music-service/internal/playlist/models"
	"github.com/go-sqlx/sqlx"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	CreatePlaylist(ctx context.Context, userID, name, description string, isPublic bool) (*models.Playlist, error)
	GetPlaylist(ctx context.Context, id string) (*models.Playlist, error)
	ListUserPlaylists(ctx context.Context, userID string, page, pageSize int) ([]*models.Playlist, error)
	UpdatePlaylist(ctx context.Context, id string, name, description *string, isPublic *bool) (*models.Playlist, error)
	DeletePlaylist(ctx context.Context, id string) error
	AddTrack(ctx context.Context, playlistID, trackID string) error
	RemoveTrack(ctx context.Context, playlistID, trackID string) error
	GetPlaylistTracks(ctx context.Context, playlistID string) ([]models.PlaylistTrack, error)
}

type repository struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewRepository(db *sqlx.DB, log *slog.Logger) Repository {
	return &repository{
		db:  db,
		log: log,
	}
}

func (r *repository) CreatePlaylist(ctx context.Context, userID, name, description string, isPublic bool) (*models.Playlist, error) {
	query := `
		INSERT INTO playlists (
			id,
			user_id,
			name,
			descritption,
			is_public,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			NOW(),
			NOW()
		)
		RETURNING
			id,
			user_id,
			name,
			descritpion,
			is_public,
			created_at,
			updated_at
	`

	p := &models.Playlist{ID: uuid.New().String()}
	err := r.db.QueryRowContext(ctx, query, p.ID, userID, name, description, isPublic).Scan(
		&p.ID,
		&p.UserID,
		&p.Name,
		&p.Description,
		&p.IsPublic,
		&p.Type,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	return p, err
}

func (r *repository) GetPlaylist(ctx context.Context, id string) (*models.Playlist, error) {
	query := `
		SELECT
			id,
			user_id,
			name,
			description,
			is_public,
			created_at,
			updated_at
		FROM
			playlists
		WHERE
			id = $1
	`

	var p models.Playlist
	err := r.db.GetContext(ctx, &p, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *repository) ListUserPlaylists(ctx context.Context, userID string, page, pageSize int) ([]*models.Playlist, error) {
	query := `
	SELECT
			id,
			user_id,
			name,
			description,
			is_public,
			created_at,
			updated_at
		FROM
			playlists
		WHERE
			user_id = $1
		ORDER BY
			created_at DESC
		LIMIT
			$2
		OFFSET
			$3
	`

	var playlists []*models.Playlist
	err := r.db.SelectContext(ctx, &playlists, query, userID, pageSize, pageSize*(page-1))
	return playlists, err
}

func (r *repository) UpdatePlaylist(ctx context.Context, id string, name, description *string, isPublic *bool) (*models.Playlist, error) {
	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{id}
	argIdx := 2

	if name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, name)
		argIdx++
	}

	if description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, description)
		argIdx++
	}

	if isPublic != nil {
		setParts = append(setParts, fmt.Sprintf("is_public = $%d", argIdx))
		args = append(args, isPublic)
	}

	query := fmt.Sprintf(`
		UPDATE
			playlists
		SET
			%s
		WHERE
			id = $1
		RETURNING
			id,
			user_id,
			name,
			descritpion,
			is_public,
			created_at,
			updated_at
	`, strings.Join(setParts, ", "))

	var p models.Playlist
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&p.ID,
		&p.UserID,
		&p.Name,
		&p.Description,
		&p.IsPublic,
		&p.Type,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	
	return &p, err
}

func (r *repository) DeletePlaylist(ctx context.Context, id string) error {
	result, _ := r.db.ExecContext(ctx, "DELETE FROM playlists WHERE id = $1", id)
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) AddTrack(ctx context.Context, playlistID, trackID string) error {
	query := `
		INSERT INTO playlist_tracks (
			id,
			playlist_id,
			track_id,
			position,
			added_at
		)
		SELECT
			$1,
			$2,
			$3,
			COALESCE((SELECT MAX(position) FROM playlist_tracks WHERE playlist_id = $2), -1) + 1,
			NOW()
		WHERE NOT EXISTS (
			SELECT
				1
			FROM
				playlist_tracks
			WHERE
				playlist_id = $2 AND
				track_id = $3 
		)
		ON CONFLICT (playlist_id, track_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, uuid.New().String(), playlistID, trackID)
	return err
}

func (r *repository) RemoveTrack(ctx context.Context, playlistID, trackID string) error {
	query := `
		DELETE FROM
			playlist_tracks
		WHERE
			playlist_id = $1 AND
			track_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, playlistID, trackID)
	return err
}

func (r *repository) GetPlaylistTracks(ctx context.Context, playlistID string) ([]models.PlaylistTrack, error) {
	query := `
		SELECT
			id,
			playlist_id,
			track_id,
			position,
			added_at
		FROM
			playlist_tracks
		WHERE
			playlist_id = $1
		ORDER BY
			position
	`

	var tracks []models.PlaylistTrack
	err := r.db.SelectContext(ctx, &tracks, query, playlistID)
	return tracks, err
}