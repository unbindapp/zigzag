package favorite_service

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	favorite_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/favorite"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i FavoriteServiceInterface -p favorite_service -s FavoriteService -o favorite_service_iface.go
type FavoriteService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewFavoriteService(cfg *config.Config, repo repositories.RepositoriesInterface) *FavoriteService {
	return &FavoriteService{cfg: cfg, repo: repo}
}

type ListFavoritesInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
}

func (self *FavoriteService) List(ctx context.Context, userID uuid.UUID, input *ListFavoritesInput) ([]*models.FavoriteResponse, error) {
	if err := self.checkViewer(ctx, userID, input.WorkspaceID); err != nil {
		return nil, err
	}
	favorites, err := self.repo.Favorite().ListByUser(ctx, userID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	idsByKind := map[schema.FavoriteKind][]uuid.UUID{}
	for _, found := range favorites {
		idsByKind[found.Kind] = append(idsByKind[found.Kind], found.EntityID)
	}
	resolved := map[schema.FavoriteKind]map[uuid.UUID]favorite_repo.ResolvedEntity{}
	for kind, ids := range idsByKind {
		if resolved[kind], err = self.repo.Favorite().Resolve(ctx, kind, ids); err != nil {
			return nil, err
		}
	}
	visibility := &teamVisibility{service: self, userID: userID, known: map[uuid.UUID]bool{}}
	responses := make([]*models.FavoriteResponse, 0, len(favorites))
	for _, found := range favorites {
		entity, ok := resolved[found.Kind][found.EntityID]
		if !ok {
			continue
		}
		visible, err := visibility.allows(ctx, input.WorkspaceID, entity)
		if err != nil {
			return nil, err
		}
		if visible {
			responses = append(responses, models.TransformFavoriteEntity(found, entity.Name, entity.Icon, entity.Color, entity.URLParams))
		}
	}
	return responses, nil
}

type FavoriteTargetInput struct {
	WorkspaceID uuid.UUID           `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace the entity belongs to"`
	Kind        schema.FavoriteKind `json:"kind" required:"true" doc:"issue, project, cycle, view, document, label, team or initiative"`
	EntityID    uuid.UUID           `json:"entity_id" required:"true" format:"uuid" doc:"Id of the issue, project, cycle, view, document, label, team or initiative"`
}

func (self *FavoriteService) Add(ctx context.Context, userID uuid.UUID, input *FavoriteTargetInput) (*models.FavoriteResponse, error) {
	if err := self.checkViewer(ctx, userID, input.WorkspaceID); err != nil {
		return nil, err
	}
	if !slices.Contains(input.Kind.Values(), string(input.Kind)) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "kind must be issue, project, cycle, view, document, label, team or initiative")
	}
	resolved, err := self.repo.Favorite().Resolve(ctx, input.Kind, []uuid.UUID{input.EntityID})
	if err != nil {
		return nil, err
	}
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, string(input.Kind)+" not found")
	entity, ok := resolved[input.EntityID]
	if !ok {
		return nil, notFound
	}
	visibility := &teamVisibility{service: self, userID: userID, known: map[uuid.UUID]bool{}}
	visible, err := visibility.allows(ctx, input.WorkspaceID, entity)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, notFound
	}

	var favorite *ent.Favorite
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		existing, err := self.repo.Favorite().Find(ctx, tx, userID, input.Kind, input.EntityID)
		if err != nil || existing != nil {
			favorite = existing
			return err
		}
		sortOrder, err := self.repo.Favorite().NextSortOrder(ctx, tx, userID, input.WorkspaceID)
		if err != nil {
			return err
		}
		favorite, err = self.repo.Favorite().Create(ctx, tx, &favorite_repo.CreateFavoriteInput{
			UserID: userID, WorkspaceID: input.WorkspaceID, Kind: input.Kind, EntityID: input.EntityID, SortOrder: sortOrder,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return models.TransformFavoriteEntity(favorite, entity.Name, entity.Icon, entity.Color, entity.URLParams), nil
}

func (self *FavoriteService) Remove(ctx context.Context, userID uuid.UUID, input *FavoriteTargetInput) error {
	if err := self.checkViewer(ctx, userID, input.WorkspaceID); err != nil {
		return err
	}
	removed, err := self.repo.Favorite().Delete(ctx, userID, input.WorkspaceID, input.Kind, input.EntityID)
	if err != nil {
		return err
	}
	if removed == 0 {
		return errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Favorite not found")
	}
	return nil
}

type ReorderFavoritesInput struct {
	WorkspaceID uuid.UUID   `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	FavoriteIDs []uuid.UUID `json:"favorite_ids" format:"uuid" required:"true" nullable:"false" maxItems:"500" doc:"Favorite ids (not entity ids) in the new order, first on top"`
}

func (self *FavoriteService) Reorder(ctx context.Context, userID uuid.UUID, input *ReorderFavoritesInput) ([]*models.FavoriteResponse, error) {
	if err := self.checkViewer(ctx, userID, input.WorkspaceID); err != nil {
		return nil, err
	}
	if err := self.repo.Favorite().Reorder(ctx, userID, input.WorkspaceID, input.FavoriteIDs); err != nil {
		return nil, err
	}
	return self.List(ctx, userID, &ListFavoritesInput{WorkspaceID: input.WorkspaceID})
}

func (self *FavoriteService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

func (self *FavoriteService) checkViewer(ctx context.Context, userID, workspaceID uuid.UUID) error {
	return errdefs.MaskAsNotFound(self.check(ctx, userID, schema.WorkspaceRoleViewer, workspaceID), "Workspace not found")
}

type teamVisibility struct {
	service *FavoriteService
	userID  uuid.UUID
	isAdmin *bool
	known   map[uuid.UUID]bool
}

func (self *teamVisibility) allows(ctx context.Context, workspaceID uuid.UUID, entity favorite_repo.ResolvedEntity) (bool, error) {
	if entity.WorkspaceID != workspaceID {
		return false, nil
	}
	if entity.ViewOwnerID != nil && *entity.ViewOwnerID != self.userID && !entity.ViewShared {
		return false, nil
	}
	if entity.TeamID == nil {
		return true, nil
	}
	if visible, seen := self.known[*entity.TeamID]; seen {
		return visible, nil
	}
	visible, err := self.teamVisible(ctx, workspaceID, *entity.TeamID)
	if err != nil {
		return false, err
	}
	self.known[*entity.TeamID] = visible
	return visible, nil
}

func (self *teamVisibility) teamVisible(ctx context.Context, workspaceID, teamID uuid.UUID) (bool, error) {
	team, err := self.service.repo.Team().GetByID(ctx, teamID)
	if ent.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !team.Private {
		return true, nil
	}
	if self.isAdmin == nil {
		admin := self.service.check(ctx, self.userID, schema.WorkspaceRoleAdmin, workspaceID) == nil
		self.isAdmin = &admin
	}
	if *self.isAdmin {
		return true, nil
	}
	return self.service.repo.Team().IsMember(ctx, teamID, self.userID)
}
