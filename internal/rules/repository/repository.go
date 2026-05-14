package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/CAATHARSIS/music-service/internal/rules/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	CreateRule(ctx context.Context, rule *models.Rule) error
	GetRule(ctx context.Context, id string) (*models.Rule, error)
	ListUserRules(ctx context.Context, userID string) ([]*models.Rule, error)
	UpdateRule(ctx context.Context, rule *models.Rule) error
	DeleteRule(ctx context.Context, id string) error
	MarkExecuted(ctx context.Context, id string) error
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

func (r *repository) CreateRule(ctx context.Context, rule *models.Rule) error {
	if err := rule.MarshalCondition(); err != nil {
		return fmt.Errorf("marshal condition: %w", err)
	}

	query := `
		INSERT INTO rules (
			id,
			user_id,
			name,
			condition,
			track_limit,
			cron_schedule,
			is_active,
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

	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	_, err := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.UserID,
		rule.Name,
		rule.ConditionJSON,
		rule.TrackLimit,
		rule.CronSchedule,
		rule.IsActive,
	)
	return err
}

func (r *repository) GetRule(ctx context.Context, id string) (*models.Rule, error) {
	query := `
		SELECT
			id,
			user_id,
			name,
			condition,
			track_limit,
			cron_schedule,
			is_active,
			last_executed,
			created_at,
			updated_at
		FROM
			rules
		WHERE
			id = $1
	`

	var rule models.Rule
	err := r.db.GetContext(ctx, &rule, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	rule.UnmarshalCondition()
	return &rule, nil
}

func (r *repository) ListUserRules(ctx context.Context, userID string) ([]*models.Rule, error) {
	query := `
		SELECT
			id,
			user_id,
			name,
			condition,
			track_limit,
			cron_schedule,
			is_active,
			last_executed,
			created_at,
			updated_at
		FROM
			rules
		WHERE
			user_id = $1
		ORDER BY
			created_at DESC
	`

	var rules []*models.Rule
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rule models.Rule

		err := rows.Scan(
			&rule.ID,
			&rule.UserID,
			&rule.Name,
			&rule.ConditionJSON,
			&rule.TrackLimit,
			&rule.CronSchedule,
			&rule.IsActive,
			&rule.LastExecuted,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rule while listing: %w", err)
		}

		rules = append(rules, &rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error while listing: %w", err)
	}

	for i := range rules {
		rules[i].UnmarshalCondition()
	}

	return rules, err
}

func (r *repository) UpdateRule(ctx context.Context, rule *models.Rule) error {
	if err := rule.MarshalCondition(); err != nil {
		return fmt.Errorf("marshal condition: %w", err)
	}

	query := `
		UPDATE
			rules
		SET
			name = $2,
			condition = $3,
			track_limit = $4,
			cron_schedule = $5,
			is_active = $6,
			updated_at = NOW()
		WHERE
			id = $1
	`

	result, _ := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.Name,
		rule.Condition,
		rule.TrackLimit,
		rule.CronSchedule,
		rule.IsActive,
	)
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) DeleteRule(ctx context.Context, id string) error {
	result, _ := r.db.ExecContext(ctx, "DELETE FROM rules WHERE id = $1", id)
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) MarkExecuted(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE rules SET last_executed = NOW() WHERE id = $1", id)
	return err
}