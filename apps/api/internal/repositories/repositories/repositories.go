package repositories

import (
	"context"

	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	apikey_repo "github.com/unbindapp/zigzag-api/internal/repositories/apikey"
	comment_repo "github.com/unbindapp/zigzag-api/internal/repositories/comment"
	cycle_repo "github.com/unbindapp/zigzag-api/internal/repositories/cycle"
	document_repo "github.com/unbindapp/zigzag-api/internal/repositories/document"
	favorite_repo "github.com/unbindapp/zigzag-api/internal/repositories/favorite"
	importjob_repo "github.com/unbindapp/zigzag-api/internal/repositories/importjob"
	initiative_repo "github.com/unbindapp/zigzag-api/internal/repositories/initiative"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
	label_repo "github.com/unbindapp/zigzag-api/internal/repositories/label"
	notification_repo "github.com/unbindapp/zigzag-api/internal/repositories/notification"
	oauth_repo "github.com/unbindapp/zigzag-api/internal/repositories/oauth"
	oauthserver_repo "github.com/unbindapp/zigzag-api/internal/repositories/oauthserver"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	project_repo "github.com/unbindapp/zigzag-api/internal/repositories/project"
	system_repo "github.com/unbindapp/zigzag-api/internal/repositories/system"
	team_repo "github.com/unbindapp/zigzag-api/internal/repositories/team"
	template_repo "github.com/unbindapp/zigzag-api/internal/repositories/template"
	upload_repo "github.com/unbindapp/zigzag-api/internal/repositories/upload"
	user_repo "github.com/unbindapp/zigzag-api/internal/repositories/user"
	view_repo "github.com/unbindapp/zigzag-api/internal/repositories/view"
	webhook_repo "github.com/unbindapp/zigzag-api/internal/repositories/webhook"
	workspace_repo "github.com/unbindapp/zigzag-api/internal/repositories/workspace"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i RepositoriesInterface -p repositories -s Repositories -o repositories_iface.go
type Repositories struct {
	db           *ent.Client
	base         *repository.BaseRepository
	user         user_repo.UserRepositoryInterface
	oauth        oauth_repo.OauthRepositoryInterface
	oauthServer  oauthserver_repo.OAuthServerRepositoryInterface
	apiKey       apikey_repo.APIKeyRepositoryInterface
	system       system_repo.SystemRepositoryInterface
	workspace    workspace_repo.WorkspaceRepositoryInterface
	team         team_repo.TeamRepositoryInterface
	label        label_repo.LabelRepositoryInterface
	project      project_repo.ProjectRepositoryInterface
	initiative   initiative_repo.InitiativeRepositoryInterface
	cycle        cycle_repo.CycleRepositoryInterface
	issue        issue_repo.IssueRepositoryInterface
	comment      comment_repo.CommentRepositoryInterface
	notification notification_repo.NotificationRepositoryInterface
	document     document_repo.DocumentRepositoryInterface
	view         view_repo.ViewRepositoryInterface
	template     template_repo.TemplateRepositoryInterface
	favorite     favorite_repo.FavoriteRepositoryInterface
	webhook      webhook_repo.WebhookRepositoryInterface
	upload       upload_repo.UploadRepositoryInterface
	importJob    importjob_repo.ImportJobRepositoryInterface
	permissions  permissions_repo.PermissionsRepositoryInterface
}

func NewRepositories(db *ent.Client) *Repositories {
	return &Repositories{
		db:           db,
		base:         repository.NewBaseRepository(db),
		user:         user_repo.NewUserRepository(db),
		oauth:        oauth_repo.NewOauthRepository(db),
		oauthServer:  oauthserver_repo.NewOAuthServerRepository(db),
		apiKey:       apikey_repo.NewAPIKeyRepository(db),
		system:       system_repo.NewSystemRepository(db),
		workspace:    workspace_repo.NewWorkspaceRepository(db),
		team:         team_repo.NewTeamRepository(db),
		label:        label_repo.NewLabelRepository(db),
		project:      project_repo.NewProjectRepository(db),
		initiative:   initiative_repo.NewInitiativeRepository(db),
		cycle:        cycle_repo.NewCycleRepository(db),
		issue:        issue_repo.NewIssueRepository(db),
		comment:      comment_repo.NewCommentRepository(db),
		notification: notification_repo.NewNotificationRepository(db),
		document:     document_repo.NewDocumentRepository(db),
		view:         view_repo.NewViewRepository(db),
		template:     template_repo.NewTemplateRepository(db),
		favorite:     favorite_repo.NewFavoriteRepository(db),
		webhook:      webhook_repo.NewWebhookRepository(db),
		upload:       upload_repo.NewUploadRepository(db),
		importJob:    importjob_repo.NewImportJobRepository(db),
		permissions:  permissions_repo.NewPermissionsRepository(db),
	}
}

func (r *Repositories) Ent() *ent.Client {
	return r.db
}

func (r *Repositories) WithTx(ctx context.Context, fn func(tx repository.TxInterface) error) error {
	return r.base.WithTx(ctx, fn)
}

func (r *Repositories) User() user_repo.UserRepositoryInterface {
	return r.user
}

func (r *Repositories) Oauth() oauth_repo.OauthRepositoryInterface {
	return r.oauth
}

func (r *Repositories) OAuthServer() oauthserver_repo.OAuthServerRepositoryInterface {
	return r.oauthServer
}

func (r *Repositories) APIKey() apikey_repo.APIKeyRepositoryInterface {
	return r.apiKey
}

func (r *Repositories) System() system_repo.SystemRepositoryInterface {
	return r.system
}

func (r *Repositories) Workspace() workspace_repo.WorkspaceRepositoryInterface {
	return r.workspace
}

func (r *Repositories) Team() team_repo.TeamRepositoryInterface {
	return r.team
}

func (r *Repositories) Label() label_repo.LabelRepositoryInterface {
	return r.label
}

func (r *Repositories) Project() project_repo.ProjectRepositoryInterface {
	return r.project
}

func (r *Repositories) Initiative() initiative_repo.InitiativeRepositoryInterface {
	return r.initiative
}

func (r *Repositories) Cycle() cycle_repo.CycleRepositoryInterface {
	return r.cycle
}

func (r *Repositories) Issue() issue_repo.IssueRepositoryInterface {
	return r.issue
}

func (r *Repositories) Comment() comment_repo.CommentRepositoryInterface {
	return r.comment
}

func (r *Repositories) Notification() notification_repo.NotificationRepositoryInterface {
	return r.notification
}

func (r *Repositories) Document() document_repo.DocumentRepositoryInterface {
	return r.document
}

func (r *Repositories) View() view_repo.ViewRepositoryInterface {
	return r.view
}

func (r *Repositories) Template() template_repo.TemplateRepositoryInterface {
	return r.template
}

func (r *Repositories) Favorite() favorite_repo.FavoriteRepositoryInterface {
	return r.favorite
}

func (r *Repositories) Webhook() webhook_repo.WebhookRepositoryInterface {
	return r.webhook
}

func (r *Repositories) Upload() upload_repo.UploadRepositoryInterface {
	return r.upload
}

func (r *Repositories) ImportJob() importjob_repo.ImportJobRepositoryInterface {
	return r.importJob
}

func (r *Repositories) Permissions() permissions_repo.PermissionsRepositoryInterface {
	return r.permissions
}
