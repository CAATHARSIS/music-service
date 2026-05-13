package repository

import (
	"context"
	"log/slog"

	"github.com/CAATHARSIS/music-service/internal/file/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostgresRepository interface {
	CreateFile(ctx context.Context, file *models.File) error
	GetFile(ctx context.Context, id string) (*models.File, error)
	DeleteFile(ctx context.Context, id string) error
}

type postgresRepo struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewPostgresRepo(db *sqlx.DB, log *slog.Logger) PostgresRepository {
	return &postgresRepo{
		db:  db,
		log: log,
	}
}

func (r *postgresRepo) CreateFile(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (
			id,
			original_name,
			bucket,
			key,
			size,
			mime_type,
			uploaded_by,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			NOW(),
			NOW()
		)
	`

	if file.ID == "" {
		file.ID = uuid.New().String()
	}
	_, err := r.db.ExecContext(ctx, query,
		file.ID,
		file.OriginalName,
		file.Bucket,
		file.Key,
		file.Size,
		file.MimeType,
		file.UploadedBy,
	)
	return err
}

func (r *postgresRepo) GetFile(ctx context.Context, id string) (*models.File, error) {
	query := `SELECT * FROM files WHERE id = $1`
	var file models.File
	err := r.db.GetContext(ctx, &file, query, id)
	return &file, err
}

func (r *postgresRepo) DeleteFile(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM files WHERE id = $1", id)
	return err
}