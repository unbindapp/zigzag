package integrations_handler

import (
	"context"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	github_service "github.com/zigzagapp-org/zigzag-api/internal/services/github"
)

type UpdateGithubInput struct {
	server.BaseAuthInput
	Body *github_service.UpdateGithubIntegrationInput `required:"true"`
}

type GithubIntegrationResponse struct {
	Body struct {
		Data *models.GithubIntegrationResponse `json:"data"`
	}
}

func (self *HandlerGroup) UpdateGithub(ctx context.Context, input *UpdateGithubInput) (*GithubIntegrationResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := self.srv.GithubService.UpdateIntegration(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &GithubIntegrationResponse{}
	resp.Body.Data = updated
	return resp, nil
}

type ListGithubRepoSyncsInput struct {
	server.BaseAuthInput
	github_service.ListGithubRepoSyncsInput
}

type ListGithubRepoSyncsResponse struct {
	Body struct {
		Data []*models.GithubRepoSyncResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) ListGithubRepoSyncs(ctx context.Context, input *ListGithubRepoSyncsInput) (*ListGithubRepoSyncsResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	syncs, err := self.srv.GithubService.ListRepoSyncs(ctx, user.ID, &input.ListGithubRepoSyncsInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ListGithubRepoSyncsResponse{}
	resp.Body.Data = syncs
	return resp, nil
}

type GetGithubRepoSyncInput struct {
	server.BaseAuthInput
	github_service.GetGithubRepoSyncInput
}

type GithubRepoSyncResponse struct {
	Body struct {
		Data *models.GithubRepoSyncResponse `json:"data"`
	}
}

func (self *HandlerGroup) GetGithubRepoSync(ctx context.Context, input *GetGithubRepoSyncInput) (*GithubRepoSyncResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	repoSync, err := self.srv.GithubService.GetRepoSync(ctx, user.ID, &input.GetGithubRepoSyncInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &GithubRepoSyncResponse{}
	resp.Body.Data = repoSync
	return resp, nil
}

type AddGithubRepoSyncInput struct {
	server.BaseAuthInput
	Body *github_service.AddGithubRepoSyncInput `required:"true"`
}

func (self *HandlerGroup) AddGithubRepoSync(ctx context.Context, input *AddGithubRepoSyncInput) (*GithubRepoSyncResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	repoSync, err := self.srv.GithubService.AddRepoSync(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &GithubRepoSyncResponse{}
	resp.Body.Data = repoSync
	return resp, nil
}

type BackfillGithubRepoSyncInput struct {
	server.BaseAuthInput
	Body *github_service.BackfillGithubRepoSyncInput `required:"true"`
}

func (self *HandlerGroup) BackfillGithubRepoSync(ctx context.Context, input *BackfillGithubRepoSyncInput) (*GithubRepoSyncResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	repoSync, err := self.srv.GithubService.Backfill(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &GithubRepoSyncResponse{}
	resp.Body.Data = repoSync
	return resp, nil
}

type RemoveGithubRepoSyncInput struct {
	server.BaseAuthInput
	Body *github_service.GithubRepoSyncRefInput `required:"true"`
}

type RemoveGithubRepoSyncResponse struct {
	Body struct {
		Data models.DeletedResponse `json:"data"`
	}
}

func (self *HandlerGroup) RemoveGithubRepoSync(ctx context.Context, input *RemoveGithubRepoSyncInput) (*RemoveGithubRepoSyncResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	id, err := self.srv.GithubService.RemoveRepoSync(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &RemoveGithubRepoSyncResponse{}
	resp.Body.Data = models.DeletedResponse{ID: id, Deleted: true}
	return resp, nil
}
