package search_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	issue_service "github.com/unbindapp/zigzag-api/internal/services/issue"
)

const defaultLimit = 10

type TeamVisibility interface {
	VisibleTeams(ctx context.Context, userID, workspaceID uuid.UUID) (*issue_service.VisibleTeams, error)
}

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i SearchServiceInterface -p search_service -s SearchService -o search_service_iface.go
type SearchService struct {
	cfg   *config.Config
	repo  repositories.RepositoriesInterface
	teams TeamVisibility
}

func NewSearchService(cfg *config.Config, repo repositories.RepositoriesInterface, teams TeamVisibility) *SearchService {
	return &SearchService{cfg: cfg, repo: repo, teams: teams}
}

type SearchInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"Workspace to search in"`
	Query       string    `query:"query" required:"true" minLength:"1" maxLength:"200" doc:"Words to look for in issue titles and descriptions, project names and summaries, document titles and content"`
	Limit       int       `query:"limit" required:"false" minimum:"1" maximum:"50" default:"10" doc:"Most results per kind, 1 to 50"`
}

func (self *SearchService) Search(ctx context.Context, userID uuid.UUID, input *SearchInput) (*models.SearchResponse, error) {
	check := []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleViewer, WorkspaceID: input.WorkspaceID}}
	if err := self.repo.Permissions().Check(ctx, userID, check); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	limit := input.Limit
	if limit <= 0 {
		limit = defaultLimit
	}

	visible, err := self.teams.VisibleTeams(ctx, userID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	issueIDs, err := self.repo.Issue().Search(ctx, input.WorkspaceID, visible.IDs(), input.Query, limit)
	if err != nil {
		return nil, err
	}
	issues, err := self.repo.Issue().GetManyWithEdges(ctx, nil, issueIDs)
	if err != nil {
		return nil, err
	}
	workspaces := []uuid.UUID{input.WorkspaceID}
	projects, err := self.repo.Project().Search(ctx, workspaces, input.Query, limit)
	if err != nil {
		return nil, err
	}
	documents, err := self.repo.Document().Search(ctx, workspaces, input.Query, limit)
	if err != nil {
		return nil, err
	}

	response := &models.SearchResponse{
		Issues:    make([]*models.IssueSummary, len(issues)),
		Projects:  make([]*models.ProjectSummary, len(projects)),
		Documents: models.TransformDocuments(documents),
	}
	for i, issue := range issues {
		response.Issues[i] = models.TransformIssueSummary(issue)
	}
	for i, project := range projects {
		response.Projects[i] = models.TransformProjectSummary(project)
	}
	return response, nil
}
