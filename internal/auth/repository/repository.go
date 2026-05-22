package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/CAATHARSIS/music-service/internal/auth/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Repository interface {
	CreateUser(ctx context.Context, username, email, password string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	VerifyPassword(user *models.User, password string) bool
	SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, id string) error
	SetUserRole(ctx context.Context, userID, role string) (*models.User, error)
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

func (r *repository) CreateUser(ctx context.Context, username, email, password string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	query := `
		INSERT INTO users (
			id,
			username,
			email,
			password_hash,
			role,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			'user',
			NOW(),
			NOW()
		)
		RETURNING
			id,
			username,
			email,
			role,
			created_at,
			updated_at
	`

	user := &models.User{ID: uuid.New().String()}

	err = r.db.QueryRowContext(ctx, query,
		user.ID,
		username,
		email,
		string(hash),
	).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT
			id,
			username,
			email,
			password_hash,
			role,
			avatar_image_id,
			created_at,
			updated_at
		FROM
			users
		WHERE
			email = $1
	`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &user, nil
}

func (r *repository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT
			id,
			username,
			email,
			role,
			avatar_image_id,
			created_at,
			updated_at
		FROM
			users
		WHERE
			id = $1
	`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &user, nil
}

func (r *repository) VerifyPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

func (r *repository) SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (
			id,
			user_id,
			token_hash,
			expires_at,
			created_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			NOW()
		)
	`

	_, err := r.db.ExecContext(ctx, query, uuid.New().String(), userID, tokenHash, expiresAt)
	return err
}

func (r *repository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			created_at
		FROM
			refresh_tokens
		WHERE
			token_hash = $1 AND
			expires_at > NOW()
	`

	var token models.RefreshToken
	err := r.db.GetContext(ctx, &token, query, tokenHash)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	return &token, nil
}

func (r *repository) DeleteRefreshToken(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE id = $1", id)
	return err
}

func (r *repository) SetUserRole(ctx context.Context, userID, role string) (*models.User, error) {
	query := `
		UPDATE
			users
		SET
			role = $1,
			updated_at = NOW()
		WHERE
			id = $2
		RETURNING
			id,
			username,
			email,
			role,
			avatar_image_id,
			created_at,
			updated_at
	`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, role, userID)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return &user, err
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}