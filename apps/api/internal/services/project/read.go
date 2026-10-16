package project_service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	project_repo "github.com/unbindapp/zigzag-api/internal/repositories/project"
)

type ListProjectsInput struct {
	WorkspaceID     uuid.UUID                `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	TeamID          uuid.UUID                `query:"team_id" required:"false" format:"uuid" doc:"Only projects this team takes part in"`
	StatusType      schema.ProjectStatusType `query:"status_type" required:"false" doc:"Only projects whose status has this type: backlog, planned, started, paused, completed or canceled"`
	StatusID        uuid.UUID                `query:"status_id" required:"false" format:"uuid" doc:"Only projects in this status"`
	LeadID          uuid.UUID                `query:"lead_id" required:"false" format:"uuid" doc:"Only projects led by this user"`
	MemberID        uuid.UUID                `query:"member_id" required:"false" format:"uuid" doc:"Only projects this user is a member of"`
	InitiativeID    uuid.UUID                `query:"initiative_id" required:"false" format:"uuid" doc:"Only projects in this initiative"`
	LabelID         uuid.UUID                `query:"label_id" required:"false" format:"uuid" doc:"Only projects with this label"`
	Query           string                   `query:"query" required:"false" maxLength:"200" doc:"Case insensitive text matched against name and summary"`
	IncludeArchived bool                     `query:"include_archived" required:"false" doc:"Include archived projects"`
	IncludeTrashed  bool                     `query:"include_trashed" required:"false" doc:"Include trashed projects"`
	OrderBy         string                   `query:"order_by" required:"false" enum:"sort_order,name,priority,target_date,created_at,updated_at" doc:"Sort order; sort_order (manual) by default, created_at and updated_at newest first"`
}

func optionalID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func (self *ProjectService) List(ctx context.Context, userID uuid.UUID, input *ListProjectsInput) ([]*models.ProjectResponse, error) {
	if err := self.checkVisible(ctx, userID, input.WorkspaceID, "Workspace not found"); err != nil {
		return nil, err
	}
	filter := &project_repo.ListProjectsFilter{
		WorkspaceID:     input.WorkspaceID,
		TeamID:          optionalID(input.TeamID),
		StatusID:        optionalID(input.StatusID),
		LeadID:          optionalID(input.LeadID),
		MemberID:        optionalID(input.MemberID),
		InitiativeID:    optionalID(input.InitiativeID),
		LabelID:         optionalID(input.LabelID),
		Query:           strings.TrimSpace(input.Query),
		IncludeArchived: input.IncludeArchived,
		IncludeTrashed:  input.IncludeTrashed,
		OrderBy:         project_repo.ProjectOrder(input.OrderBy),
	}
	if input.StatusType != "" {
		if !validStatusType(input.StatusType) {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "status_type must be backlog, planned, started, paused, completed or canceled")
		}
		filter.StatusType = &input.StatusType
	}
	projects, err := self.repo.Project().List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return self.respond(ctx, projects)
}

type GetProjectInput struct {
	ProjectID   uuid.UUID `query:"project_id" required:"false" format:"uuid" doc:"Project id; give this, or workspace_id and slug"`
	WorkspaceID uuid.UUID `query:"workspace_id" required:"false" format:"uuid" doc:"Workspace of the project when looking up by slug"`
	Slug        string    `query:"slug" required:"false" maxLength:"80" doc:"Project slug from its URL, such as launch-1e0e8889, with workspace_id"`
}

func (self *ProjectService) Get(ctx context.Context, userID uuid.UUID, input *GetProjectInput) (*models.ProjectResponse, error) {
	project, err := self.findProject(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	response, err := self.respondOne(ctx, project)
	if err != nil {
		return nil, err
	}
	latest, err := self.repo.Project().LatestUpdate(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	response.LastUpdate = models.TransformProjectUpdate(latest, userID)
	return response, nil
}

func (self *ProjectService) findProject(ctx context.Context, userID uuid.UUID, input *GetProjectInput) (*ent.Project, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Project not found")
	if input.ProjectID != uuid.Nil {
		return self.projectWithRole(ctx, userID, input.ProjectID, schema.WorkspaceRoleViewer)
	}
	if input.WorkspaceID == uuid.Nil || input.Slug == "" {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Give project_id, or workspace_id and slug")
	}
	if err := self.checkVisible(ctx, userID, input.WorkspaceID, "Project not found"); err != nil {
		return nil, err
	}
	project, err := self.repo.Project().GetBySlug(ctx, input.WorkspaceID, input.Slug)
	if ent.IsNotFound(err) {
		return nil, notFound
	}
	return project, err
}
