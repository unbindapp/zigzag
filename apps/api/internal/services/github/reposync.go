package github_service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/integrations/github"
	"github.com/unbindapp/zigzag-api/internal/models"
	github_repo "github.com/unbindapp/zigzag-api/internal/repositories/github"
	project_service "github.com/unbindapp/zigzag-api/internal/services/project"
)

const (
	minTokenLength       = 20
	maxTokenLength       = 255
	maxProjectNameLength = 80
	fineGrainedPrefix    = "github_pat_"
)

type UpdateGithubIntegrationInput struct {
	WorkspaceID uuid.UUID `json:"workspace_id" required:"true" format:"uuid" doc:"Workspace whose GitHub connection to change"`
	Token       string    `json:"token" required:"true" maxLength:"255" doc:"Fine-grained GitHub token with Issues read and write on the synced repositories; empty clears it"`
}

func (self *GithubService) UpdateIntegration(ctx context.Context, userID uuid.UUID, input *UpdateGithubIntegrationInput) (*models.GithubIntegrationResponse, error) {
	integration, err := self.adminIntegration(ctx, userID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(input.Token)
	var sealed []byte
	if token != "" {
		if len(token) < minTokenLength || len(token) > maxTokenLength || strings.ContainsFunc(token, func(r rune) bool { return r <= ' ' || r > '~' }) {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "token must be a GitHub token like github_pat_...")
		}
		if sealed, err = self.encryptor.Encrypt([]byte(token)); err != nil {
			return nil, errdefs.NewInternalError(err, "Failed to encrypt the GitHub token")
		}
	}
	updated, err := self.repo.Github().UpdateToken(ctx, integration.ID, sealed, tokenPrefix(token))
	if err != nil {
		return nil, err
	}
	return models.TransformGithubIntegration(updated, self.webhookURL(updated.WorkspaceID)), nil
}

func tokenPrefix(token string) string {
	if token == "" {
		return ""
	}
	if strings.HasPrefix(token, fineGrainedPrefix) {
		return token[:len(fineGrainedPrefix)+4]
	}
	if index := strings.Index(token, "_"); index > 0 && index+5 < len(token) {
		return token[:index+5]
	}
	return token[:4]
}

type ListGithubRepoSyncsInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"Workspace whose synced repositories to list"`
}

func (self *GithubService) ListRepoSyncs(ctx context.Context, userID uuid.UUID, input *ListGithubRepoSyncsInput) ([]*models.GithubRepoSyncResponse, error) {
	if err := self.checkAdmin(ctx, userID, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	integration, err := self.repo.Github().GetByWorkspace(ctx, input.WorkspaceID)
	if ent.IsNotFound(err) {
		return []*models.GithubRepoSyncResponse{}, nil
	}
	if err != nil {
		return nil, err
	}
	syncs, err := self.repo.Github().ListRepoSyncs(ctx, integration.ID)
	if err != nil {
		return nil, err
	}
	return self.respondRepoSyncs(ctx, syncs)
}

type GetGithubRepoSyncInput struct {
	RepoSyncID uuid.UUID `query:"repo_sync_id" required:"true" format:"uuid" doc:"The synced repository"`
}

func (self *GithubService) GetRepoSync(ctx context.Context, userID uuid.UUID, input *GetGithubRepoSyncInput) (*models.GithubRepoSyncResponse, error) {
	repoSync, err := self.adminRepoSync(ctx, userID, input.RepoSyncID)
	if err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Synced repository not found")
	}
	return self.respondRepoSync(ctx, repoSync)
}

type AddGithubRepoSyncInput struct {
	WorkspaceID uuid.UUID `json:"workspace_id" required:"true" format:"uuid" doc:"Workspace to sync the repository into"`
	FullName    string    `json:"full_name" required:"true" minLength:"3" maxLength:"200" doc:"Repository as owner/name, like unbindapp/zigzag"`
	TeamID      uuid.UUID `json:"team_id" required:"true" format:"uuid" doc:"Team new issues from GitHub are created in"`
	SyncClosed  bool      `json:"sync_closed,omitempty" required:"false" doc:"Include closed issues when importing existing issues"`
}

func (self *GithubService) AddRepoSync(ctx context.Context, userID uuid.UUID, input *AddGithubRepoSyncInput) (*models.GithubRepoSyncResponse, error) {
	integration, err := self.adminIntegration(ctx, userID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	fullName, ok := github.NormalizeFullName(input.FullName)
	if !ok {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "full_name must be owner/name, like unbindapp/zigzag")
	}
	team, err := self.repo.Team().GetByID(ctx, input.TeamID)
	if ent.IsNotFound(err) || (err == nil && (team.WorkspaceID != input.WorkspaceID || team.ArchivedAt != nil)) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "team_id must be an active team of the workspace")
	}
	if err != nil {
		return nil, err
	}
	taken, err := self.repo.Github().RepoSyncNameTaken(ctx, fullName)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeConflict, fullName+" is already synced")
	}
	if err := self.verifyRepository(ctx, integration, fullName); err != nil {
		return nil, err
	}
	projectID, err := self.projectFor(ctx, userID, input.WorkspaceID, team.ID, fullName)
	if err != nil {
		return nil, err
	}
	created, err := self.repo.Github().CreateRepoSync(ctx, &github_repo.CreateRepoSyncInput{
		IntegrationID: integration.ID, FullName: fullName, TeamID: team.ID, ProjectID: projectID, SyncClosed: input.SyncClosed,
	})
	if ent.IsConstraintError(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeConflict, fullName+" is already synced")
	}
	if err != nil {
		return nil, err
	}
	return self.respondRepoSync(ctx, created)
}

func (self *GithubService) verifyRepository(ctx context.Context, integration *ent.GithubIntegration, fullName string) error {
	token, err := self.token(integration)
	if err != nil || token == "" {
		return err
	}
	repository, err := self.client.GetRepository(ctx, token, fullName)
	if github.IsNotFound(err) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "GitHub cannot find "+fullName+" with the stored token")
	}
	if err != nil {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Could not reach "+fullName+" on GitHub: "+err.Error())
	}
	if !repository.HasIssues {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, fullName+" has issues turned off on GitHub")
	}
	return nil
}

func (self *GithubService) projectFor(ctx context.Context, userID, workspaceID, teamID uuid.UUID, fullName string) (uuid.UUID, error) {
	name := fullName
	if utf8.RuneCountInString(name) > maxProjectNameLength {
		name = string([]rune(name)[:maxProjectNameLength])
	}
	existing, err := self.repo.Project().GetByName(ctx, workspaceID, name)
	if err != nil && !ent.IsNotFound(err) {
		return uuid.Nil, err
	}
	if existing != nil {
		return existing.ID, self.attachTeam(ctx, userID, existing, teamID)
	}
	created, err := self.projects.Create(ctx, userID, &project_service.CreateProjectInput{
		WorkspaceID: workspaceID, Name: name, TeamIDs: []uuid.UUID{teamID}, Summary: "Issues synced from GitHub",
	})
	if err != nil {
		return uuid.Nil, err
	}
	return created.ID, nil
}

func (self *GithubService) attachTeam(ctx context.Context, userID uuid.UUID, project *ent.Project, teamID uuid.UUID) error {
	teamIDs := []uuid.UUID{teamID}
	for _, team := range project.Edges.Teams {
		if team.ID == teamID {
			return nil
		}
		teamIDs = append(teamIDs, team.ID)
	}
	_, err := self.projects.Update(ctx, userID, &project_service.UpdateProjectInput{ProjectID: project.ID, TeamIDs: teamIDs})
	return err
}

type GithubRepoSyncRefInput struct {
	RepoSyncID uuid.UUID `json:"repo_sync_id" required:"true" format:"uuid" doc:"The synced repository"`
}

func (self *GithubService) RemoveRepoSync(ctx context.Context, userID uuid.UUID, input *GithubRepoSyncRefInput) (uuid.UUID, error) {
	repoSync, err := self.adminRepoSync(ctx, userID, input.RepoSyncID)
	if err != nil {
		return uuid.Nil, err
	}
	return repoSync.ID, self.repo.Github().DeleteRepoSync(ctx, repoSync.ID)
}

func (self *GithubService) adminRepoSync(ctx context.Context, userID, repoSyncID uuid.UUID) (*ent.GithubRepoSync, error) {
	repoSync, err := self.repo.Github().GetRepoSync(ctx, repoSyncID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Synced repository not found")
	}
	if err != nil {
		return nil, err
	}
	if err := self.checkAdmin(ctx, userID, repoSync.Edges.Integration.WorkspaceID); err != nil {
		return nil, err
	}
	return repoSync, nil
}

func (self *GithubService) token(integration *ent.GithubIntegration) (string, error) {
	if len(integration.TokenEncrypted) == 0 {
		return "", nil
	}
	token, err := self.encryptor.Decrypt(integration.TokenEncrypted)
	if err != nil {
		return "", errdefs.NewInternalError(err, "Failed to decrypt the GitHub token")
	}
	return string(token), nil
}

func (self *GithubService) respondRepoSync(ctx context.Context, repoSync *ent.GithubRepoSync) (*models.GithubRepoSyncResponse, error) {
	responses, err := self.respondRepoSyncs(ctx, []*ent.GithubRepoSync{repoSync})
	if err != nil {
		return nil, err
	}
	return responses[0], nil
}

func (self *GithubService) respondRepoSyncs(ctx context.Context, syncs []*ent.GithubRepoSync) ([]*models.GithubRepoSyncResponse, error) {
	ids := make([]uuid.UUID, len(syncs))
	for i, repoSync := range syncs {
		ids[i] = repoSync.ID
	}
	counts, err := self.repo.Github().LinkedIssueCounts(ctx, ids)
	if err != nil {
		return nil, err
	}
	responses := make([]*models.GithubRepoSyncResponse, len(syncs))
	for i, repoSync := range syncs {
		responses[i] = models.TransformGithubRepoSync(repoSync, counts[repoSync.ID])
	}
	return responses, nil
}
