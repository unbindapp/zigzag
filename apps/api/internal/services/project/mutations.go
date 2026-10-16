package project_service

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/api/apictx"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	"github.com/unbindapp/zigzag-api/internal/common/utils"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
	project_repo "github.com/unbindapp/zigzag-api/internal/repositories/project"
)

const slugBaseLength = 48

type CreateProjectInput struct {
	WorkspaceID          uuid.UUID             `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	Name                 string                `json:"name" required:"true" minLength:"1" maxLength:"80" doc:"Project name"`
	TeamIDs              []uuid.UUID           `json:"team_ids" format:"uuid" required:"true" nullable:"false" minItems:"1" doc:"Teams working on the project, at least one, all from this workspace"`
	Summary              string                `json:"summary,omitempty" required:"false" maxLength:"255" doc:"One line summary"`
	Description          string                `json:"description,omitempty" required:"false" maxLength:"100000" doc:"Markdown description"`
	Icon                 string                `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name"`
	Color                string                `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^(#[0-9a-fA-F]{6})?$" doc:"Hex color such as #5e6ad2"`
	StatusID             uuid.UUID             `json:"status_id,omitempty" required:"false" format:"uuid" doc:"Project status from list-project-statuses; defaults to the first backlog status"`
	LeadID               uuid.UUID             `json:"lead_id,omitempty" required:"false" format:"uuid" doc:"Workspace member who leads the project"`
	MemberIDs            []uuid.UUID           `json:"member_ids,omitempty" format:"uuid" required:"false" doc:"Workspace members working on the project"`
	Priority             int                   `json:"priority,omitempty" required:"false" minimum:"0" maximum:"4" doc:"0 none, 1 urgent, 2 high, 3 medium, 4 low"`
	StartDate            string                `json:"start_date,omitempty" required:"false" format:"date" doc:"Planned start, YYYY-MM-DD"`
	TargetDate           string                `json:"target_date,omitempty" required:"false" format:"date" doc:"Target completion, YYYY-MM-DD"`
	StartDateResolution  schema.DateResolution `json:"start_date_resolution,omitempty" required:"false" doc:"Precision of start_date: day (default), month, quarter, half_year or year"`
	TargetDateResolution schema.DateResolution `json:"target_date_resolution,omitempty" required:"false" doc:"Precision of target_date: day (default), month, quarter, half_year or year"`
	LabelIDs             []uuid.UUID           `json:"label_ids,omitempty" format:"uuid" required:"false" doc:"Workspace labels (not team labels, not groups)"`
	InitiativeIDs        []uuid.UUID           `json:"initiative_ids,omitempty" format:"uuid" required:"false" doc:"Initiatives the project belongs to"`
}

func (self *ProjectService) Create(ctx context.Context, userID uuid.UUID, input *CreateProjectInput) (*models.ProjectResponse, error) {
	if err := self.checkRole(ctx, userID, schema.WorkspaceRoleMember, input.WorkspaceID, "Workspace not found"); err != nil {
		return nil, err
	}
	name, err := names.Clean(input.Name)
	if err != nil {
		return nil, err
	}
	color, err := names.OptionalColor(input.Color)
	if err != nil {
		return nil, err
	}
	startDate, err := models.ParseDate("start_date", input.StartDate)
	if err != nil {
		return nil, err
	}
	targetDate, err := models.ParseDate("target_date", input.TargetDate)
	if err != nil {
		return nil, err
	}
	if err := validateDates(startDate, targetDate); err != nil {
		return nil, err
	}
	if err := validateResolutions(input.StartDateResolution, input.TargetDateResolution); err != nil {
		return nil, err
	}
	refs := &references{TeamIDs: input.TeamIDs, MemberIDs: input.MemberIDs, LabelIDs: input.LabelIDs, InitiativeIDs: input.InitiativeIDs}
	if input.LeadID != uuid.Nil {
		refs.LeadID = &input.LeadID
	}
	if err := self.validateReferences(ctx, userID, input.WorkspaceID, refs); err != nil {
		return nil, err
	}
	status, err := self.statusForCreate(ctx, input.WorkspaceID, input.StatusID)
	if err != nil {
		return nil, err
	}
	slug, err := utils.SlugWithSuffix(utils.TruncateSlug(utils.Slugify(name), slugBaseLength), 4)
	if err != nil {
		return nil, err
	}
	sortOrder, err := self.repo.Project().MaxSortOrder(ctx, input.WorkspaceID)
	if err != nil {
		return nil, err
	}

	now := self.now()
	create := &project_repo.CreateProjectInput{
		WorkspaceID:          input.WorkspaceID,
		Name:                 name,
		Slug:                 slug,
		StatusID:             status.ID,
		Summary:              strings.TrimSpace(input.Summary),
		Description:          input.Description,
		Icon:                 input.Icon,
		Color:                color,
		LeadID:               refs.LeadID,
		Priority:             input.Priority,
		CreatorID:            &userID,
		StartDate:            startDate,
		TargetDate:           targetDate,
		StartDateResolution:  input.StartDateResolution,
		TargetDateResolution: input.TargetDateResolution,
		SortOrder:            sortOrder + 1,
		TeamIDs:              uniqueIDs(input.TeamIDs),
		MemberIDs:            uniqueIDs(input.MemberIDs),
		LabelIDs:             uniqueIDs(input.LabelIDs),
		InitiativeIDs:        uniqueIDs(input.InitiativeIDs),
	}
	switch status.Type {
	case schema.ProjectStatusStarted:
		create.StartedAt = &now
	case schema.ProjectStatusCompleted:
		create.StartedAt, create.CompletedAt = &now, &now
	case schema.ProjectStatusCanceled:
		create.CanceledAt = &now
	}

	var created *ent.Project
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		created, err = self.repo.Project().Create(ctx, tx, create)
		return err
	})
	if err != nil {
		return nil, err
	}
	response, err := self.reload(ctx, created.ID)
	if err != nil {
		return nil, err
	}
	self.notifyLead(ctx, userID, response, nil)
	self.notifyMembers(ctx, userID, response, create.MemberIDs)
	return response, nil
}

type UpdateProjectInput struct {
	ProjectID            uuid.UUID              `json:"project_id" required:"true" format:"uuid" doc:"The project to change"`
	Name                 *string                `json:"name,omitempty" required:"false" maxLength:"80" doc:"New name"`
	TeamIDs              []uuid.UUID            `json:"team_ids,omitempty" format:"uuid" required:"false" doc:"Replaces the teams; at least one"`
	Summary              *string                `json:"summary,omitempty" required:"false" maxLength:"255" doc:"One line summary, empty to clear"`
	Description          *string                `json:"description,omitempty" required:"false" maxLength:"100000" doc:"Markdown description, empty to clear"`
	Icon                 *string                `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name, empty to clear"`
	Color                *string                `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^(#[0-9a-fA-F]{6})?$" doc:"Hex color, empty to clear"`
	StatusID             *uuid.UUID             `json:"status_id,omitempty" required:"false" format:"uuid" doc:"New project status; moving to a started, completed or canceled status stamps started_at, completed_at or canceled_at"`
	LeadID               *uuid.UUID             `json:"lead_id,omitempty" required:"false" format:"uuid" doc:"New lead; the nil uuid 00000000-0000-0000-0000-000000000000 removes the lead"`
	MemberIDs            []uuid.UUID            `json:"member_ids,omitempty" format:"uuid" required:"false" doc:"Replaces the members; an empty list removes everyone"`
	Priority             *int                   `json:"priority,omitempty" required:"false" minimum:"0" maximum:"4" doc:"0 none, 1 urgent, 2 high, 3 medium, 4 low"`
	StartDate            *string                `json:"start_date,omitempty" required:"false" doc:"Planned start YYYY-MM-DD, empty to clear"`
	TargetDate           *string                `json:"target_date,omitempty" required:"false" doc:"Target completion YYYY-MM-DD, empty to clear"`
	StartDateResolution  *schema.DateResolution `json:"start_date_resolution,omitempty" required:"false" doc:"Precision of start_date: day, month, quarter, half_year or year"`
	TargetDateResolution *schema.DateResolution `json:"target_date_resolution,omitempty" required:"false" doc:"Precision of target_date: day, month, quarter, half_year or year"`
	LabelIDs             []uuid.UUID            `json:"label_ids,omitempty" format:"uuid" required:"false" doc:"Replaces the labels; an empty list removes all"`
	InitiativeIDs        []uuid.UUID            `json:"initiative_ids,omitempty" format:"uuid" required:"false" doc:"Replaces the initiatives; an empty list removes all"`
	SortOrder            *float64               `json:"sort_order,omitempty" required:"false" doc:"Manual order in lists and boards"`
}

func (self *ProjectService) Update(ctx context.Context, userID uuid.UUID, input *UpdateProjectInput) (*models.ProjectResponse, error) {
	project, err := self.projectWithRole(ctx, userID, input.ProjectID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	update, refs, err := self.buildUpdate(ctx, project, input)
	if err != nil {
		return nil, err
	}
	if err := self.validateReferences(ctx, userID, project.WorkspaceID, refs); err != nil {
		return nil, err
	}
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		return self.repo.Project().Update(ctx, tx, project.ID, update)
	})
	if err != nil {
		return nil, err
	}
	response, err := self.reload(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	if update.LeadID != nil {
		self.notifyLead(ctx, userID, response, project.LeadID)
	}
	if update.ReplaceMembers {
		self.notifyMembers(ctx, userID, response, addedIDs(memberIDs(project), update.MemberIDs))
	}
	self.emitUpdated(ctx, response)
	return response, nil
}

func (self *ProjectService) buildUpdate(ctx context.Context, project *ent.Project, input *UpdateProjectInput) (*project_repo.UpdateProjectInput, *references, error) {
	update := &project_repo.UpdateProjectInput{
		Description: input.Description, Icon: input.Icon, Priority: input.Priority, SortOrder: input.SortOrder,
		StartDateResolution: input.StartDateResolution, TargetDateResolution: input.TargetDateResolution,
	}
	refs := &references{}
	if input.Name != nil {
		name, err := names.Clean(*input.Name)
		if err != nil {
			return nil, nil, err
		}
		update.Name = &name
	}
	if input.Summary != nil {
		summary := strings.TrimSpace(*input.Summary)
		update.Summary = &summary
	}
	if input.Color != nil {
		color, err := names.OptionalColor(*input.Color)
		if err != nil {
			return nil, nil, err
		}
		update.Color = &color
	}
	if err := validateOptionalResolutions(input.StartDateResolution, input.TargetDateResolution); err != nil {
		return nil, nil, err
	}
	if err := applyDates(project, input, update); err != nil {
		return nil, nil, err
	}
	if input.LeadID != nil {
		update.ClearLead = *input.LeadID == uuid.Nil
		if !update.ClearLead {
			update.LeadID = input.LeadID
			refs.LeadID = input.LeadID
		}
	}
	if input.TeamIDs != nil {
		if len(input.TeamIDs) == 0 {
			return nil, nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "A project needs at least one team")
		}
		update.ReplaceTeams, update.TeamIDs, refs.TeamIDs = true, uniqueIDs(input.TeamIDs), input.TeamIDs
	}
	if input.MemberIDs != nil {
		update.ReplaceMembers, update.MemberIDs, refs.MemberIDs = true, uniqueIDs(input.MemberIDs), input.MemberIDs
	}
	if input.LabelIDs != nil {
		update.ReplaceLabels, update.LabelIDs, refs.LabelIDs = true, uniqueIDs(input.LabelIDs), input.LabelIDs
	}
	if input.InitiativeIDs != nil {
		update.ReplaceInitiatives, update.InitiativeIDs, refs.InitiativeIDs = true, uniqueIDs(input.InitiativeIDs), input.InitiativeIDs
	}
	if input.StatusID != nil && *input.StatusID != project.StatusID {
		status, err := self.workspaceStatus(ctx, project.WorkspaceID, *input.StatusID)
		if err != nil {
			return nil, nil, err
		}
		update.StatusID = &status.ID
		self.stampStatus(project, status.Type, update)
	}
	return update, refs, nil
}

func applyDates(project *ent.Project, input *UpdateProjectInput, update *project_repo.UpdateProjectInput) error {
	startDate, targetDate := project.StartDate, project.TargetDate
	if input.StartDate != nil {
		parsed, err := models.ParseDate("start_date", *input.StartDate)
		if err != nil {
			return err
		}
		startDate, update.StartDate, update.ClearStartDate = parsed, parsed, parsed == nil
	}
	if input.TargetDate != nil {
		parsed, err := models.ParseDate("target_date", *input.TargetDate)
		if err != nil {
			return err
		}
		targetDate, update.TargetDate, update.ClearTargetDate = parsed, parsed, parsed == nil
	}
	return validateDates(startDate, targetDate)
}

func (self *ProjectService) stampStatus(project *ent.Project, statusType schema.ProjectStatusType, update *project_repo.UpdateProjectInput) {
	now := self.now()
	if project.StartedAt == nil && (statusType == schema.ProjectStatusStarted || statusType == schema.ProjectStatusCompleted) {
		update.StartedAt = &now
	}
	update.ClearCompletedAt = project.CompletedAt != nil
	update.ClearCanceledAt = project.CanceledAt != nil
	switch statusType {
	case schema.ProjectStatusCompleted:
		update.CompletedAt, update.ClearCompletedAt = &now, false
	case schema.ProjectStatusCanceled:
		update.CanceledAt, update.ClearCanceledAt = &now, false
	}
}

type ProjectRefInput struct {
	ProjectID uuid.UUID `json:"project_id" required:"true" format:"uuid" doc:"The project"`
}

func (self *ProjectService) Delete(ctx context.Context, userID uuid.UUID, input *ProjectRefInput) error {
	project, err := self.projectWithRole(ctx, userID, input.ProjectID, schema.WorkspaceRoleAdmin)
	if err != nil {
		return err
	}
	return self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		issues, err := self.repo.Project().IssuesInProject(ctx, tx, project.ID)
		if err != nil {
			return err
		}
		entries := make([]issue_repo.HistoryEntry, len(issues))
		for i, issue := range issues {
			changes := []schema.FieldChange{{Field: "project", From: project.ID.String(), FromLabel: project.Name}}
			if milestone := issue.Edges.Milestone; milestone != nil {
				changes = append(changes, schema.FieldChange{Field: "milestone", From: milestone.ID.String(), FromLabel: milestone.Name})
			}
			entries[i] = issue_repo.HistoryEntry{IssueID: issue.ID, ActorID: &userID, Source: apictx.HistorySource(ctx), Changes: changes}
		}
		if err := self.repo.Issue().CreateHistoryMany(ctx, tx, entries); err != nil {
			return err
		}
		return self.repo.Project().Delete(ctx, tx, project.ID)
	})
}

func (self *ProjectService) Archive(ctx context.Context, userID uuid.UUID, input *ProjectRefInput) (*models.ProjectResponse, error) {
	now := self.now()
	return self.setLifecycle(ctx, userID, input.ProjectID, self.repo.Project().SetArchivedAt, &now)
}

func (self *ProjectService) Unarchive(ctx context.Context, userID uuid.UUID, input *ProjectRefInput) (*models.ProjectResponse, error) {
	return self.setLifecycle(ctx, userID, input.ProjectID, self.repo.Project().SetArchivedAt, nil)
}

func (self *ProjectService) Trash(ctx context.Context, userID uuid.UUID, input *ProjectRefInput) (*models.ProjectResponse, error) {
	now := self.now()
	return self.setLifecycle(ctx, userID, input.ProjectID, self.repo.Project().SetTrashedAt, &now)
}

func (self *ProjectService) Restore(ctx context.Context, userID uuid.UUID, input *ProjectRefInput) (*models.ProjectResponse, error) {
	return self.setLifecycle(ctx, userID, input.ProjectID, self.repo.Project().SetTrashedAt, nil)
}

func (self *ProjectService) setLifecycle(ctx context.Context, userID, projectID uuid.UUID, set func(context.Context, uuid.UUID, *time.Time) error, at *time.Time) (*models.ProjectResponse, error) {
	project, err := self.projectWithRole(ctx, userID, projectID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	if err := set(ctx, project.ID, at); err != nil {
		return nil, err
	}
	response, err := self.reload(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	self.emitUpdated(ctx, response)
	return response, nil
}

func (self *ProjectService) notifyLead(ctx context.Context, actorID uuid.UUID, project *models.ProjectResponse, previousLead *uuid.UUID) {
	if project.Lead == nil || project.Lead.ID == actorID {
		return
	}
	if previousLead != nil && *previousLead == project.Lead.ID {
		return
	}
	self.notify(ctx, models.NotificationEvent{
		Type:         schema.NotificationProjectLead,
		WorkspaceID:  project.WorkspaceID,
		ActorID:      &actorID,
		RecipientIDs: []uuid.UUID{project.Lead.ID},
		ProjectID:    &project.ID,
		Title:        "You were made lead of " + project.Name,
	})
}

func (self *ProjectService) notifyMembers(ctx context.Context, actorID uuid.UUID, project *models.ProjectResponse, added []uuid.UUID) {
	recipients := slices.DeleteFunc(slices.Clone(added), func(id uuid.UUID) bool { return id == actorID })
	self.notify(ctx, models.NotificationEvent{
		Type:         schema.NotificationProjectMember,
		WorkspaceID:  project.WorkspaceID,
		ActorID:      &actorID,
		RecipientIDs: recipients,
		ProjectID:    &project.ID,
		Title:        "You were added to " + project.Name,
	})
}

func memberIDs(project *ent.Project) []uuid.UUID {
	ids := make([]uuid.UUID, len(project.Edges.Members))
	for i, member := range project.Edges.Members {
		ids[i] = member.ID
	}
	return ids
}

func addedIDs(before, after []uuid.UUID) []uuid.UUID {
	added := []uuid.UUID{}
	for _, id := range after {
		if !slices.Contains(before, id) {
			added = append(added, id)
		}
	}
	return added
}

func uniqueIDs(ids []uuid.UUID) []uuid.UUID {
	unique := []uuid.UUID{}
	for _, id := range ids {
		if !slices.Contains(unique, id) {
			unique = append(unique, id)
		}
	}
	return unique
}
