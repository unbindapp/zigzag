package github_sync_service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/crypto"
	"github.com/unbindapp/zigzag-api/internal/integrations/github"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

var errNoToken = errors.New("no GitHub token; add one to send changes made in Zigzag back to GitHub")

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i GithubSyncServiceInterface -p github_sync_service -s GithubSyncService -o github_sync_service_iface.go
type GithubSyncService struct {
	repo      repositories.RepositoriesInterface
	encryptor *crypto.Encryptor
	client    github.ClientInterface
	now       func() time.Time
	run       func(func())
}

func NewGithubSyncService(repo repositories.RepositoriesInterface, encryptor *crypto.Encryptor, client github.ClientInterface) *GithubSyncService {
	return &GithubSyncService{repo: repo, encryptor: encryptor, client: client, now: time.Now, run: func(job func()) { go job() }}
}

func (self *GithubSyncService) AfterIssueUpdate(ctx context.Context, issueID uuid.UUID, changes []schema.FieldChange) {
	if _, inbound := github.InboundFrom(ctx); inbound || len(changes) == 0 {
		return
	}
	detached := context.WithoutCancel(ctx)
	self.run(func() { self.pushIssue(detached, issueID, changes) })
}

func (self *GithubSyncService) AfterCommentCreate(ctx context.Context, commentID uuid.UUID) {
	self.afterComment(ctx, commentID, self.pushNewComment)
}

func (self *GithubSyncService) AfterCommentUpdate(ctx context.Context, commentID uuid.UUID) {
	self.afterComment(ctx, commentID, self.pushEditedComment)
}

func (self *GithubSyncService) AfterCommentDelete(ctx context.Context, commentID uuid.UUID) {
	self.afterComment(ctx, commentID, self.pushDeletedComment)
}

func (self *GithubSyncService) afterComment(ctx context.Context, commentID uuid.UUID, push func(context.Context, uuid.UUID)) {
	if _, inbound := github.InboundFrom(ctx); inbound {
		return
	}
	detached := context.WithoutCancel(ctx)
	self.run(func() { push(detached, commentID) })
}

type target struct {
	link     *ent.GithubIssueLink
	repoSync *ent.GithubRepoSync
	token    string
}

func (self *GithubSyncService) targetOf(ctx context.Context, link *ent.GithubIssueLink) (*target, error) {
	repoSync := link.Edges.RepoSync
	if repoSync == nil || repoSync.Edges.Integration == nil {
		return nil, errors.New("the GitHub link has no repository")
	}
	sealed := repoSync.Edges.Integration.TokenEncrypted
	if len(sealed) == 0 {
		return &target{link: link, repoSync: repoSync}, errNoToken
	}
	token, err := self.encryptor.Decrypt(sealed)
	if err != nil {
		return &target{link: link, repoSync: repoSync}, err
	}
	return &target{link: link, repoSync: repoSync, token: string(token)}, nil
}

func (self *GithubSyncService) issueTarget(ctx context.Context, issueID uuid.UUID) (*target, error) {
	link, err := self.repo.Github().IssueLinkByIssue(ctx, issueID)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return self.targetOf(ctx, link)
}

func (self *GithubSyncService) record(ctx context.Context, target *target, err error) {
	if target == nil && err == nil {
		return
	}
	if err != nil {
		log.Error("Failed to sync a change to GitHub", "err", err)
	}
	if target == nil {
		return
	}
	if recordErr := self.repo.Github().RecordSyncResult(ctx, target.repoSync.ID, errdefs.DetailMessage(err), self.now()); recordErr != nil {
		log.Error("Failed to record the GitHub sync result", "repo", target.repoSync.FullName, "err", recordErr)
	}
}
