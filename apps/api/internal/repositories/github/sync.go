package github_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/githubcommentlink"
	"github.com/zigzagapp-org/zigzag-api/ent/githubintegration"
	"github.com/zigzagapp-org/zigzag-api/ent/githubissuelink"
	"github.com/zigzagapp-org/zigzag-api/ent/githubreposync"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

func (self *GithubRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.GithubIntegration, error) {
	return self.base.DB.GithubIntegration.Query().Where(githubintegration.ID(id)).Only(ctx)
}

func (self *GithubRepository) UpdateToken(ctx context.Context, id uuid.UUID, tokenEncrypted []byte, tokenPrefix string) (*ent.GithubIntegration, error) {
	update := self.base.DB.GithubIntegration.UpdateOneID(id).SetTokenPrefix(tokenPrefix)
	if tokenEncrypted == nil {
		update.ClearTokenEncrypted()
	} else {
		update.SetTokenEncrypted(tokenEncrypted)
	}
	updated, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}
	return self.GetByWorkspace(ctx, updated.WorkspaceID)
}

func withRepoSyncEdges(q *ent.GithubRepoSyncQuery) *ent.GithubRepoSyncQuery {
	return q.WithTeam().WithProject().WithIntegration()
}

func (self *GithubRepository) ListRepoSyncs(ctx context.Context, integrationID uuid.UUID) ([]*ent.GithubRepoSync, error) {
	return withRepoSyncEdges(self.base.DB.GithubRepoSync.Query()).
		Where(githubreposync.IntegrationID(integrationID)).
		Order(ent.Asc(githubreposync.FieldFullName)).
		All(ctx)
}

func (self *GithubRepository) GetRepoSync(ctx context.Context, id uuid.UUID) (*ent.GithubRepoSync, error) {
	return withRepoSyncEdges(self.base.DB.GithubRepoSync.Query()).Where(githubreposync.ID(id)).Only(ctx)
}

func (self *GithubRepository) RepoSyncByName(ctx context.Context, integrationID uuid.UUID, fullName string) (*ent.GithubRepoSync, error) {
	return withRepoSyncEdges(self.base.DB.GithubRepoSync.Query()).
		Where(githubreposync.IntegrationID(integrationID), githubreposync.FullName(fullName)).
		Only(ctx)
}

func (self *GithubRepository) RepoSyncNameTaken(ctx context.Context, fullName string) (bool, error) {
	return self.base.DB.GithubRepoSync.Query().Where(githubreposync.FullName(fullName)).Exist(ctx)
}

type CreateRepoSyncInput struct {
	IntegrationID uuid.UUID
	FullName      string
	TeamID        uuid.UUID
	ProjectID     uuid.UUID
	SyncClosed    bool
}

func (self *GithubRepository) CreateRepoSync(ctx context.Context, input *CreateRepoSyncInput) (*ent.GithubRepoSync, error) {
	created, err := self.base.DB.GithubRepoSync.Create().
		SetIntegrationID(input.IntegrationID).
		SetFullName(input.FullName).
		SetTeamID(input.TeamID).
		SetProjectID(input.ProjectID).
		SetSyncClosed(input.SyncClosed).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return self.GetRepoSync(ctx, created.ID)
}

func (self *GithubRepository) DeleteRepoSync(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.GithubRepoSync.DeleteOneID(id).Exec(ctx)
}

func (self *GithubRepository) StartBackfill(ctx context.Context, id uuid.UUID, syncClosed bool) (bool, error) {
	updated, err := self.base.DB.GithubRepoSync.Update().
		Where(githubreposync.ID(id), githubreposync.BackfillStatusNEQ(schema.GithubBackfillRunning)).
		SetBackfillStatus(schema.GithubBackfillRunning).
		SetSyncClosed(syncClosed).
		SetBackfillReport(&schema.GithubBackfillReport{}).
		SetLastError("").
		Save(ctx)
	return updated == 1, err
}

func (self *GithubRepository) RecordBackfillProgress(ctx context.Context, id uuid.UUID, report *schema.GithubBackfillReport) error {
	return self.base.DB.GithubRepoSync.UpdateOneID(id).SetBackfillReport(report).Exec(ctx)
}

func (self *GithubRepository) FinishBackfill(ctx context.Context, id uuid.UUID, status schema.GithubBackfillStatus, report *schema.GithubBackfillReport, lastError string, at time.Time) error {
	return self.base.DB.GithubRepoSync.UpdateOneID(id).
		SetBackfillStatus(status).
		SetBackfillReport(report).
		SetLastError(lastError).
		SetLastSyncedAt(at).
		Exec(ctx)
}

func (self *GithubRepository) FailInterruptedBackfills(ctx context.Context) error {
	return self.base.DB.GithubRepoSync.Update().
		Where(githubreposync.BackfillStatusEQ(schema.GithubBackfillRunning)).
		SetBackfillStatus(schema.GithubBackfillFailed).
		SetLastError("The import stopped when the server restarted; run it again").
		Exec(ctx)
}

func (self *GithubRepository) RecordSyncResult(ctx context.Context, id uuid.UUID, lastError string, at time.Time) error {
	return self.base.DB.GithubRepoSync.UpdateOneID(id).SetLastError(lastError).SetLastSyncedAt(at).Exec(ctx)
}

func (self *GithubRepository) LinkedIssueCounts(ctx context.Context, repoSyncIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	var rows []struct {
		RepoSyncID uuid.UUID `json:"repo_sync_id"`
		Count      int       `json:"count"`
	}
	err := self.base.DB.GithubIssueLink.Query().
		Where(githubissuelink.RepoSyncIDIn(repoSyncIDs...)).
		GroupBy(githubissuelink.FieldRepoSyncID).
		Aggregate(ent.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	counts := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		counts[row.RepoSyncID] = row.Count
	}
	return counts, nil
}

func withIssueLinkEdges(q *ent.GithubIssueLinkQuery) *ent.GithubIssueLinkQuery {
	return q.WithRepoSync(func(q *ent.GithubRepoSyncQuery) { q.WithIntegration() })
}

func (self *GithubRepository) IssueLinkByNodeID(ctx context.Context, nodeID string) (*ent.GithubIssueLink, error) {
	return withIssueLinkEdges(self.base.DB.GithubIssueLink.Query()).Where(githubissuelink.NodeID(nodeID)).Only(ctx)
}

func (self *GithubRepository) IssueLinkByIssue(ctx context.Context, issueID uuid.UUID) (*ent.GithubIssueLink, error) {
	return withIssueLinkEdges(self.base.DB.GithubIssueLink.Query()).Where(githubissuelink.IssueID(issueID)).Only(ctx)
}

type IssueLinkInput struct {
	IssueID         uuid.UUID
	RepoSyncID      uuid.UUID
	Number          int
	NodeID          string
	HTMLURL         string
	GithubUpdatedAt time.Time
}

func (self *GithubRepository) CreateIssueLink(ctx context.Context, tx repository.TxInterface, input *IssueLinkInput) (*ent.GithubIssueLink, error) {
	return self.base.Client(tx).GithubIssueLink.Create().
		SetIssueID(input.IssueID).
		SetRepoSyncID(input.RepoSyncID).
		SetNumber(input.Number).
		SetNodeID(input.NodeID).
		SetHTMLURL(input.HTMLURL).
		SetGithubUpdatedAt(input.GithubUpdatedAt).
		Save(ctx)
}

func (self *GithubRepository) Relink(ctx context.Context, id uuid.UUID, input *IssueLinkInput) error {
	return self.base.DB.GithubIssueLink.UpdateOneID(id).
		SetRepoSyncID(input.RepoSyncID).
		SetNumber(input.Number).
		SetNodeID(input.NodeID).
		SetHTMLURL(input.HTMLURL).
		SetGithubUpdatedAt(input.GithubUpdatedAt).
		Exec(ctx)
}

func (self *GithubRepository) RecordInbound(ctx context.Context, id uuid.UUID, githubUpdatedAt time.Time) error {
	return self.base.DB.GithubIssueLink.Update().
		Where(githubissuelink.ID(id), githubissuelink.GithubUpdatedAtLT(githubUpdatedAt)).
		SetGithubUpdatedAt(githubUpdatedAt).
		Exec(ctx)
}

func (self *GithubRepository) RecordOutbound(ctx context.Context, id uuid.UUID, githubUpdatedAt *time.Time, at time.Time) error {
	update := self.base.DB.GithubIssueLink.UpdateOneID(id).SetLastOutboundAt(at)
	if githubUpdatedAt != nil {
		update.SetGithubUpdatedAt(*githubUpdatedAt)
	}
	return update.Exec(ctx)
}

func (self *GithubRepository) DeleteIssueLink(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.GithubIssueLink.DeleteOneID(id).Exec(ctx)
}

func (self *GithubRepository) CommentLinkByGithubID(ctx context.Context, githubCommentID int64) (*ent.GithubCommentLink, error) {
	return self.base.DB.GithubCommentLink.Query().Where(githubcommentlink.GithubCommentID(githubCommentID)).Only(ctx)
}

func (self *GithubRepository) CommentLinkByComment(ctx context.Context, commentID uuid.UUID) (*ent.GithubCommentLink, error) {
	return self.base.DB.GithubCommentLink.Query().
		Where(githubcommentlink.CommentID(commentID)).
		WithIssueLink(func(q *ent.GithubIssueLinkQuery) { withIssueLinkEdges(q) }).
		Only(ctx)
}

func (self *GithubRepository) CreateCommentLink(ctx context.Context, tx repository.TxInterface, commentID, issueLinkID uuid.UUID, githubCommentID int64) error {
	return self.base.Client(tx).GithubCommentLink.Create().
		SetCommentID(commentID).
		SetIssueLinkID(issueLinkID).
		SetGithubCommentID(githubCommentID).
		Exec(ctx)
}

func (self *GithubRepository) DeleteCommentLink(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.GithubCommentLink.DeleteOneID(id).Exec(ctx)
}
