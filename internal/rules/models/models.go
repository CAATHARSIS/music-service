package models

import (
	"encoding/json"
	"time"
)

type RuleCondition struct {
	Genres        []string `json:"genres,omitempty"`
	YearFrom      *int     `json:"year_from,omitempty"`
	YearTo        *int     `json:"year_to,omitempty"`
	MinDuration   *int     `json:"min_duration,omitempty"`
	MaxDuration   *int     `json:"max_duration,omitempty"`
	MinPlaysCount *int     `json:"min_plays_count,omitempty"`
}

type Rule struct {
	ID            string        `db:"id"`
	UserID        string        `db:"user_id"`
	Name          string        `db:"name"`
	Condition     RuleCondition `db:"-"`
	ConditionJSON []byte        `db:"condition"`
	TrackLimit    int           `db:"track_limit"`
	CronSchedule  *string       `db:"cron_schedule"`
	IsActive      bool          `db:"is_active"`
	LastExecuted  *time.Time    `db:"last_executed"`
	CreatedAt     time.Time     `db:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at"`
}

func (r *Rule) MarshalCondition() error {
	data, err := json.Marshal(r.Condition)
	r.ConditionJSON = data
	return err
}

func (r *Rule) UnmarshalCondition() error {
	if r.ConditionJSON == nil {
		r.Condition = RuleCondition{}
		return nil
	}
	return json.Unmarshal(r.ConditionJSON, &r.Condition)
}
