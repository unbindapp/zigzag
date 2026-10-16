package issue_service

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/api/apictx"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	"github.com/unbindapp/zigzag-api/internal/integrations/github"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	github_sync_service "github.com/unbindapp/zigzag-api/internal/services/githubsync"
	notification_service "github.com/unbindapp/zigzag-api/internal/services/notification"
	webhook_service "github.com/unbindapp/zigzag-api/internal/services/webhook"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i IssueServiceInterface -p issue_service -s IssueService -o issue_service_iface.go
type IssueService struct {
	cfg           *config.Config
	repo          repositories.RepositoriesInterface
	notifications notification_service.NotificationServiceInterface
	webhooks      webhook_service.WebhookServiceInterface
	githubSync    github_sync_service.GithubSyncServiceInterface
	now           func() time.Time
}

func NewIssueService(cfg *config.Config, repo repositories.RepositoriesInterface, notifications notification_service.NotificationServiceInterface, webhooks webhook_service.WebhookServiceInterface, githubSync github_sync_service.GithubSyncServiceInterface) *IssueService {
	return &IssueService{cfg: cfg, repo: repo, notifications: notifications, webhooks: webhooks, githubSync: githubSync, now: time.Now}
}

func issueNotFound() error {
	return errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Issue not found")
}

func invalid(message string) error {
	return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, message)
}

func (self *IssueService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

type VisibleTeams struct {
	WorkspaceID uuid.UUID
	Teams       map[uuid.UUID]*ent.Team
}

func (self *VisibleTeams) Has(teamID uuid.UUID) bool {
	_, ok := self.Teams[teamID]
	return ok
}

func (self *VisibleTeams) IDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(self.Teams))
	for id := range self.Teams {
		ids = append(ids, id)
	}
	slices.SortFunc(ids, func(a, b uuid.UUID) int { return slices.Compare(a[:], b[:]) })
	return ids
}

// VisibleTeams expects the caller to hold viewer on the workspace already.
func (self *IssueService) VisibleTeams(ctx context.Context, userID, workspaceID uuid.UUID) (*VisibleTeams, error) {
	teams, err := self.repo.Team().ListByWorkspace(ctx, workspaceID, true)
	if err != nil {
		return nil, err
	}
	visible := &VisibleTeams{WorkspaceID: workspaceID, Teams: make(map[uuid.UUID]*ent.Team, len(teams))}
	hasPrivate := slices.ContainsFunc(teams, func(team *ent.Team) bool { return team.Private })
	member := map[uuid.UUID]bool{}
	admin := false
	if hasPrivate {
		admin = self.check(ctx, userID, schema.WorkspaceRoleAdmin, workspaceID) == nil
		ids, err := self.repo.Team().TeamIDsOfUser(ctx, userID, workspaceID)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			member[id] = true
		}
	}
	for _, team := range teams {
		if team.Private && !admin && !member[team.ID] {
			continue
		}
		visible.Teams[team.ID] = team
	}
	return visible, nil
}

func (self *IssueService) canSeeTeam(ctx context.Context, userID uuid.UUID, team *ent.Team) (bool, error) {
	if team == nil {
		return false, nil
	}
	if !team.Private || permissions_repo.IsSystemCaller(ctx) {
		return true, nil
	}
	if self.check(ctx, userID, schema.WorkspaceRoleAdmin, team.WorkspaceID) == nil {
		return true, nil
	}
	return self.repo.Team().IsMember(ctx, team.ID, userID)
}

func (self *IssueService) ReadableIssue(ctx context.Context, userID, issueID uuid.UUID) (*ent.Issue, error) {
	issue, err := self.repo.Issue().GetWithEdges(ctx, nil, issueID)
	if ent.IsNotFound(err) {
		return nil, issueNotFound()
	}
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, issue.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Issue not found")
	}
	visible, err := self.canSeeTeam(ctx, userID, issue.Edges.Team)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, issueNotFound()
	}
	return issue, nil
}

func (self *IssueService) WritableIssue(ctx context.Context, userID, issueID uuid.UUID, role schema.WorkspaceRole) (*ent.Issue, error) {
	issue, err := self.ReadableIssue(ctx, userID, issueID)
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, role, issue.WorkspaceID); err != nil {
		return nil, err
	}
	return issue, nil
}

func (self *IssueService) view(userID uuid.UUID, counts models.IssueCounts) models.IssueView {
	return models.IssueView{UIURL: self.cfg.ExternalUIUrl, CallerID: userID, Counts: counts}
}

func (self *IssueService) respond(ctx context.Context, userID uuid.UUID, issues []*ent.Issue) ([]*models.IssueResponse, error) {
	ids := make([]uuid.UUID, len(issues))
	for i, issue := range issues {
		ids[i] = issue.ID
	}
	counts, err := self.repo.Issue().Counts(ctx, ids)
	if err != nil {
		return nil, err
	}
	hiddenTeams, err := self.hiddenParentTeams(ctx, userID, issues)
	if err != nil {
		return nil, err
	}
	responses := make([]*models.IssueResponse, len(issues))
	for i, issue := range issues {
		responses[i] = models.TransformIssueEntity(issue, self.view(userID, counts[issue.ID]))
		if parent := issue.Edges.Parent; parent != nil && hiddenTeams[parent.TeamID] {
			responses[i].Parent = nil
		}
	}
	return responses, nil
}

func (self *IssueService) hiddenParentTeams(ctx context.Context, userID uuid.UUID, issues []*ent.Issue) (map[uuid.UUID]bool, error) {
	hidden := map[uuid.UUID]bool{}
	checked := map[uuid.UUID]bool{}
	for _, issue := range issues {
		parent := issue.Edges.Parent
		if parent == nil || parent.Edges.Team == nil || !parent.Edges.Team.Private || checked[parent.TeamID] {
			continue
		}
		checked[parent.TeamID] = true
		visible, err := self.canSeeTeam(ctx, userID, parent.Edges.Team)
		if err != nil {
			return nil, err
		}
		hidden[parent.TeamID] = !visible
	}
	return hidden, nil
}

func (self *IssueService) respondDetail(ctx context.Context, userID uuid.UUID, issue *ent.Issue) (*models.IssueDetailResponse, error) {
	responses, err := self.respond(ctx, userID, []*ent.Issue{issue})
	if err != nil {
		return nil, err
	}
	response := &models.IssueDetailResponse{IssueResponse: *responses[0]}
	visible, err := self.VisibleTeams(ctx, userID, issue.WorkspaceID)
	if err != nil {
		return nil, err
	}

	children, err := self.repo.Issue().Children(ctx, issue.ID)
	if err != nil {
		return nil, err
	}
	children = slices.DeleteFunc(children, func(child *ent.Issue) bool { return !visible.Has(child.TeamID) })
	if response.Children, err = self.respond(ctx, userID, children); err != nil {
		return nil, err
	}

	outgoing, incoming, err := self.repo.Issue().Relations(ctx, issue.ID)
	if err != nil {
		return nil, err
	}
	outgoing = slices.DeleteFunc(outgoing, func(r *ent.IssueRelation) bool { return !visible.Has(r.Edges.RelatedIssue.TeamID) })
	incoming = slices.DeleteFunc(incoming, func(r *ent.IssueRelation) bool { return !visible.Has(r.Edges.Issue.TeamID) })
	response.Relations = models.TransformRelations(issue.ID, outgoing, incoming)

	attachments, err := self.repo.Issue().Attachments(ctx, issue.ID)
	if err != nil {
		return nil, err
	}
	response.Attachments = models.TransformAttachmentEntities(attachments)

	subscribers, err := self.repo.Issue().Subscribers(ctx, issue.ID)
	if err != nil {
		return nil, err
	}
	response.Subscribers = make([]*models.UserSummary, len(subscribers))
	for i, subscriber := range subscribers {
		response.Subscribers[i] = models.TransformUserSummary(subscriber)
	}

	reactions, err := self.repo.Issue().Reactions(ctx, reactionTargetOfIssue(issue.ID))
	if err != nil {
		return nil, err
	}
	response.Reactions = models.TransformReactionGroups(reactions, userID)

	link, err := self.repo.Github().IssueLinkByIssue(ctx, issue.ID)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	response.GithubIssue = models.TransformGithubIssueLink(link)
	return response, nil
}

func (self *IssueService) reload(ctx context.Context, userID, issueID uuid.UUID) (*models.IssueResponse, error) {
	issue, err := self.repo.Issue().GetWithEdges(ctx, nil, issueID)
	if err != nil {
		return nil, err
	}
	responses, err := self.respond(ctx, userID, []*ent.Issue{issue})
	if err != nil {
		return nil, err
	}
	return responses[0], nil
}

func (self *IssueService) notify(ctx context.Context, event models.NotificationEvent) {
	if len(event.RecipientIDs) == 0 {
		return
	}
	if err := self.notifications.Notify(ctx, event); err != nil {
		log.Error("Failed to record notifications", "type", event.Type, "err", err)
	}
}

func (self *IssueService) emit(ctx context.Context, workspaceID uuid.UUID, event string, payload any) {
	if err := self.webhooks.Emit(ctx, workspaceID, event, payload); err != nil {
		log.Error("Failed to emit webhook", "event", event, "err", err)
	}
}

func actorOf(userID uuid.UUID) *uuid.UUID {
	if userID == uuid.Nil {
		return nil
	}
	return &userID
}

func (self *IssueService) recordHistory(ctx context.Context, tx repository.TxInterface, issueID uuid.UUID, actorID *uuid.UUID, changes []schema.FieldChange) error {
	if inbound, ok := github.InboundFrom(ctx); ok && len(changes) > 0 {
		changes = append(slices.Clone(changes), schema.FieldChange{Field: "github", To: inbound.URL, ToLabel: inbound.Label})
	}
	return self.repo.Issue().CreateHistory(ctx, tx, issueID, actorID, apictx.HistorySource(ctx), changes)
}

func without(ids []uuid.UUID, drop ...*uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if slices.ContainsFunc(drop, func(d *uuid.UUID) bool { return d != nil && *d == id }) || slices.Contains(out, id) {
			continue
		}
		out = append(out, id)
	}
	return out
}
