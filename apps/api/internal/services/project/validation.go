package project_service

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
)

type references struct {
	LeadID        *uuid.UUID
	TeamIDs       []uuid.UUID
	MemberIDs     []uuid.UUID
	LabelIDs      []uuid.UUID
	InitiativeIDs []uuid.UUID
}

func (self *ProjectService) validateReferences(ctx context.Context, userID, workspaceID uuid.UUID, refs *references) error {
	repo := self.repo.Project()
	people := uniqueIDs(refs.MemberIDs)
	if refs.LeadID != nil {
		people = uniqueIDs(append(people, *refs.LeadID))
	}
	checks := []struct {
		ids     []uuid.UUID
		count   func(context.Context, uuid.UUID, []uuid.UUID) (int, error)
		message string
	}{
		{uniqueIDs(refs.TeamIDs), repo.CountTeamsInWorkspace, "Every team must belong to this workspace"},
		{people, repo.CountWorkspaceMembers, "The lead and every member must be members of this workspace"},
		{uniqueIDs(refs.LabelIDs), repo.CountApplicableLabels, "Labels must be workspace labels that are not groups"},
		{uniqueIDs(refs.InitiativeIDs), repo.CountInitiativesInWorkspace, "Every initiative must belong to this workspace"},
	}
	for _, check := range checks {
		if len(check.ids) == 0 {
			continue
		}
		found, err := check.count(ctx, workspaceID, check.ids)
		if err != nil {
			return err
		}
		if found != len(check.ids) {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, check.message)
		}
	}
	return self.ensureTeamsVisible(ctx, userID, workspaceID, uniqueIDs(refs.TeamIDs))
}

func (self *ProjectService) ensureTeamsVisible(ctx context.Context, userID, workspaceID uuid.UUID, teamIDs []uuid.UUID) error {
	if len(teamIDs) == 0 || self.check(ctx, userID, schema.WorkspaceRoleAdmin, workspaceID) == nil {
		return nil
	}
	for _, teamID := range teamIDs {
		team, err := self.repo.Team().GetByID(ctx, teamID)
		if err != nil {
			return err
		}
		if !team.Private {
			continue
		}
		member, err := self.repo.Team().IsMember(ctx, teamID, userID)
		if err != nil {
			return err
		}
		if !member {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Every team must belong to this workspace")
		}
	}
	return nil
}

func (self *ProjectService) statusForCreate(ctx context.Context, workspaceID, statusID uuid.UUID) (*ent.ProjectStatus, error) {
	if statusID != uuid.Nil {
		return self.workspaceStatus(ctx, workspaceID, statusID)
	}
	status, err := self.repo.Project().DefaultStatus(ctx, workspaceID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "The workspace has no project statuses")
	}
	return status, err
}

func (self *ProjectService) workspaceStatus(ctx context.Context, workspaceID, statusID uuid.UUID) (*ent.ProjectStatus, error) {
	status, err := self.repo.Project().GetStatus(ctx, statusID)
	if ent.IsNotFound(err) || (err == nil && status.WorkspaceID != workspaceID) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "The status must be a project status of this workspace")
	}
	return status, err
}

func validateDates(start, target *time.Time) error {
	if start == nil || target == nil || !target.Before(*start) {
		return nil
	}
	return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "target_date cannot be before start_date")
}

func validateResolutions(resolutions ...schema.DateResolution) error {
	for _, resolution := range resolutions {
		if resolution == "" {
			continue
		}
		if !slices.Contains(resolution.Values(), string(resolution)) {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Date resolution must be day, month, quarter, half_year or year")
		}
	}
	return nil
}

func validateOptionalResolutions(resolutions ...*schema.DateResolution) error {
	for _, resolution := range resolutions {
		if resolution == nil {
			continue
		}
		if *resolution == "" {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Date resolution must be day, month, quarter, half_year or year")
		}
		if err := validateResolutions(*resolution); err != nil {
			return err
		}
	}
	return nil
}

func validStatusType(statusType schema.ProjectStatusType) bool {
	return slices.Contains(statusType.Values(), string(statusType))
}

func validHealth(health schema.Health) bool {
	return slices.Contains(health.Values(), string(health))
}
