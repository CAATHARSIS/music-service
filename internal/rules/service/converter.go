package service

import (
	"time"

	rulespb "github.com/CAATHARSIS/music-service/api/gen/rules"
	"github.com/CAATHARSIS/music-service/internal/rules/models"
)

func convertRuleToProto(r *models.Rule) *rulespb.Rule {
	pb := &rulespb.Rule{
		Id:        r.ID,
		UserId:    r.UserID,
		Name:      r.Name,
		Condition: convertConditionToProto(r.Condition),
		Limit:     int32(r.TrackLimit),
		IsActive:  r.IsActive,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
	}
	if r.CronSchedule != nil {
		pb.CronSchedule = *r.CronSchedule
	}
	if r.LastExecuted != nil {
		pb.LastExecutedAt = r.LastExecuted.Format(time.RFC3339)
	}
	return pb
}

func convertConditionFromProto(c *rulespb.RuleCondition) models.RuleCondition {
	if c == nil {
		return models.RuleCondition{}
	}
	return models.RuleCondition{
		Genres:        c.Genres,
		YearFrom:      toPtr(int(*c.YearFrom)),
		YearTo:        toPtr(int(*c.YearTo)),
		MinPlaysCount: toPtr(int(*c.MinPlaysCount)),
		MinDuration:   toPtr(int(*c.MinDuration)),
		MaxDuration:   toPtr(int(*c.MaxDuration)),
	}
}

func convertConditionToProto(c models.RuleCondition) *rulespb.RuleCondition {
	return &rulespb.RuleCondition{
		Genres:        c.Genres,
		YearFrom:      toPtr(int32(*c.YearFrom)),
		YearTo:        toPtr(int32(*c.YearTo)),
		MinPlaysCount: toPtr(int32(*c.MinPlaysCount)),
		MinDuration:   toPtr(int32(*c.MinDuration)),
		MaxDuration:   toPtr(int32(*c.MaxDuration)),
	}
}

func toPtr[T any](v T) *T { return &v }
