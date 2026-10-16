package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
)

type CycleProgress struct {
	Scope               float64 `json:"scope" doc:"Sum of estimates when the team estimates, else the issue count; canceled issues are left out"`
	Completed           float64 `json:"completed" doc:"Scope in completed states"`
	Started             float64 `json:"started" doc:"Scope in started states"`
	Percent             int     `json:"percent" doc:"Completed scope as a percentage of the scope, 0 to 100"`
	IssueCount          int     `json:"issue_count" doc:"Issues in the cycle, canceled ones left out"`
	CompletedIssueCount int     `json:"completed_issue_count"`
}

type CycleScopePoint struct {
	Date      string  `json:"date" format:"date" doc:"Day in the team's time zone, YYYY-MM-DD"`
	Scope     float64 `json:"scope"`
	Completed float64 `json:"completed"`
	Started   float64 `json:"started"`
}

type CycleResponse struct {
	ID           uuid.UUID         `json:"id" format:"uuid"`
	TeamID       uuid.UUID         `json:"team_id" format:"uuid"`
	Number       int               `json:"number" doc:"Number of the cycle within its team"`
	Name         string            `json:"name" doc:"Optional name; show Cycle <number> when empty"`
	Description  string            `json:"description"`
	StartsAt     time.Time         `json:"starts_at"`
	EndsAt       time.Time         `json:"ends_at"`
	CompletedAt  *time.Time        `json:"completed_at" nullable:"true"`
	ArchivedAt   *time.Time        `json:"archived_at" nullable:"true"`
	IsActive     bool              `json:"is_active"`
	IsPast       bool              `json:"is_past"`
	IsFuture     bool              `json:"is_future"`
	Progress     CycleProgress     `json:"progress"`
	ScopeHistory []CycleScopePoint `json:"scope_history,omitempty" required:"false" doc:"Daily scope snapshots, oldest first; get-cycle only"`
	CreatedAt    time.Time         `json:"created_at"`
}

func TransformCycleEntity(entity *ent.Cycle, progress CycleProgress, now time.Time) *CycleResponse {
	isPast := entity.CompletedAt != nil || !entity.EndsAt.After(now)
	isFuture := !isPast && entity.StartsAt.After(now)
	return &CycleResponse{
		ID:          entity.ID,
		TeamID:      entity.TeamID,
		Number:      entity.Number,
		Name:        entity.Name,
		Description: entity.Description,
		StartsAt:    entity.StartsAt,
		EndsAt:      entity.EndsAt,
		CompletedAt: entity.CompletedAt,
		ArchivedAt:  entity.ArchivedAt,
		IsActive:    !isPast && !isFuture,
		IsPast:      isPast,
		IsFuture:    isFuture,
		Progress:    progress,
		CreatedAt:   entity.CreatedAt,
	}
}

func TransformCycleScopeHistory(entities []*ent.CycleScopeHistory) []CycleScopePoint {
	points := make([]CycleScopePoint, len(entities))
	for i, entity := range entities {
		points[i] = CycleScopePoint{
			Date:      entity.Date.Format(time.DateOnly),
			Scope:     entity.Scope,
			Completed: entity.Completed,
			Started:   entity.Started,
		}
	}
	return points
}
