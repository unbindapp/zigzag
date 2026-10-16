package user_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/models"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	user_repo "github.com/unbindapp/zigzag-api/internal/repositories/user"
)

const maxUserListLimit = 250

func (self *UserService) Whoami(ctx context.Context, userID uuid.UUID) (*models.WhoamiResponse, error) {
	user, err := self.repo.User().GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	set, err := self.repo.Permissions().GetUserPermissionSet(ctx, userID)
	if err != nil {
		return nil, err
	}
	workspaces, err := self.repo.Workspace().GetByIDs(ctx, set.VisibleWorkspaceIDs())
	if err != nil {
		return nil, err
	}

	resp := &models.WhoamiResponse{
		User:        models.TransformUserEntity(user),
		SystemAdmin: set.IsSystemAdmin,
		Workspaces:  make([]*models.WorkspaceMembership, 0, len(workspaces)),
	}
	for _, workspace := range workspaces {
		role, _ := set.WorkspaceRole(workspace.ID)
		resp.Workspaces = append(resp.Workspaces, models.TransformWorkspaceMembership(workspace, role))
	}

	access, narrowed := permissions_repo.APIKeyAccessFromContext(ctx)
	if !narrowed {
		return resp, nil
	}
	workspaceIDs := access.WorkspaceIDs
	if workspaceIDs == nil {
		workspaceIDs = []uuid.UUID{}
	}
	privileges := access.Privileges
	if privileges == nil {
		privileges = []schema.KeyPrivilege{}
	}
	resp.APIKey = &models.WhoamiAPIKey{Role: access.Role, FullAccess: access.FullAccess, WorkspaceIDs: workspaceIDs, Privileges: privileges}
	return resp, nil
}

type ListUsersInput struct {
	Query string `query:"query" required:"false" maxLength:"200" doc:"Filter by email, name or display name"`
	Limit int    `query:"limit" required:"false" minimum:"1" maximum:"250" default:"100" doc:"Maximum number of users to return"`
}

func (self *UserService) ListUsers(ctx context.Context, requesterUserID uuid.UUID, input *ListUsersInput) ([]*models.UserResponse, error) {
	set, err := self.repo.Permissions().GetUserPermissionSet(ctx, requesterUserID)
	if err != nil {
		return nil, err
	}
	limit := input.Limit
	if limit <= 0 || limit > maxUserListLimit {
		limit = maxUserListLimit
	}
	filter := &user_repo.ListUsersInput{Query: input.Query, Limit: limit}
	if !set.IsSystemAdmin {
		filter.WorkspaceIDs = set.VisibleWorkspaceIDs()
	}
	users, err := self.repo.User().List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return models.TransformUserEntities(users), nil
}

func (self *UserService) GetNotificationSettings(ctx context.Context, userID uuid.UUID) (*models.NotificationSettingsResponse, error) {
	settings, err := self.repo.User().GetNotificationSettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	return models.TransformNotificationSettings(settings), nil
}

func (self *UserService) SetupStatus(ctx context.Context) (isFirstUserCreated bool, err error) {
	count, err := self.repo.User().Count(ctx, nil)
	return count > 0, err
}
