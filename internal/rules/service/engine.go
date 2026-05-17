package service

import (
	"context"
	"fmt"

	catalogpb "github.com/CAATHARSIS/music-service/api/gen/catalog"
	commonpb "github.com/CAATHARSIS/music-service/api/gen/common"
	"github.com/CAATHARSIS/music-service/internal/rules/models"
)

func (s ruleEngineService) executeRule(ctx context.Context, rule *models.Rule) ([]string, error) {
	req := &catalogpb.ListTracksRequest{
		Pagination: &commonpb.PaginationRequest{
			Page:     1,
			PageSize: int32(rule.TrackLimit),
		},
	}

	if len(rule.Condition.Genres) > 0 {
		req.GenreIds = rule.Condition.Genres
	}
	if rule.Condition.YearFrom != nil {
		req.YearFrom = toPtr(int32(*req.YearFrom))
	}
	if rule.Condition.YearTo != nil {
		req.YearTo = toPtr(int32(*req.YearTo))
	}

	resp, err := s.catalogClient.ListTracks(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("catalog list tracks: %w", err)
	}

	trackIDs := make([]string, len(resp.Tracks))
	for i, t := range resp.Tracks {
		if rule.Condition.MinPlaysCount != nil && t.PlaysCount < int64(*rule.Condition.MinPlaysCount) {
			continue
		}
		if rule.Condition.MinDuration != nil && t.Duration < int32(*rule.Condition.MinDuration) {
			continue
		}
		if rule.Condition.MaxDuration != nil && t.Duration > int32(*rule.Condition.MaxDuration) {
			continue
		}

		trackIDs[i] = t.Id
	}

	return trackIDs, nil
}
