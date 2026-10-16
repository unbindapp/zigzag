package favorite_repo

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/cycle"
	"github.com/unbindapp/zigzag-api/ent/document"
	"github.com/unbindapp/zigzag-api/ent/initiative"
	"github.com/unbindapp/zigzag-api/ent/issue"
	"github.com/unbindapp/zigzag-api/ent/label"
	"github.com/unbindapp/zigzag-api/ent/project"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/team"
	"github.com/unbindapp/zigzag-api/ent/view"
)

type ResolvedEntity struct {
	WorkspaceID uuid.UUID
	TeamID      *uuid.UUID
	Name        string
	Icon        string
	Color       string
	URLParams   map[string]string
	ViewOwnerID *uuid.UUID
	ViewShared  bool
}

func (self *FavoriteRepository) Resolve(ctx context.Context, kind schema.FavoriteKind, ids []uuid.UUID) (map[uuid.UUID]ResolvedEntity, error) {
	resolved := map[uuid.UUID]ResolvedEntity{}
	if len(ids) == 0 {
		return resolved, nil
	}
	db := self.base.DB
	switch kind {
	case schema.FavoriteKindIssue:
		found, err := db.Issue.Query().Where(issue.IDIn(ids...), issue.TrashedAtIsNil()).WithTeam().All(ctx)
		for _, entity := range found {
			if entity.Edges.Team == nil {
				continue
			}
			identifier := entity.Edges.Team.Key + "-" + strconv.Itoa(entity.Number)
			resolved[entity.ID] = ResolvedEntity{WorkspaceID: entity.WorkspaceID, TeamID: &entity.TeamID, Name: entity.Title, URLParams: map[string]string{"identifier": identifier}}
		}
		return resolved, err
	case schema.FavoriteKindProject:
		found, err := db.Project.Query().Where(project.IDIn(ids...), project.TrashedAtIsNil()).All(ctx)
		for _, entity := range found {
			resolved[entity.ID] = ResolvedEntity{WorkspaceID: entity.WorkspaceID, Name: entity.Name, Icon: entity.Icon, Color: entity.Color, URLParams: map[string]string{"slug": entity.Slug}}
		}
		return resolved, err
	case schema.FavoriteKindCycle:
		found, err := db.Cycle.Query().Where(cycle.IDIn(ids...), cycle.ArchivedAtIsNil()).WithTeam().All(ctx)
		for _, entity := range found {
			if entity.Edges.Team == nil {
				continue
			}
			name := entity.Name
			if name == "" {
				name = "Cycle " + strconv.Itoa(entity.Number)
			}
			resolved[entity.ID] = ResolvedEntity{
				WorkspaceID: entity.Edges.Team.WorkspaceID, TeamID: &entity.TeamID, Name: name,
				URLParams: map[string]string{"key": entity.Edges.Team.Key, "number": strconv.Itoa(entity.Number)},
			}
		}
		return resolved, err
	case schema.FavoriteKindView:
		found, err := db.View.Query().Where(view.IDIn(ids...)).All(ctx)
		for _, entity := range found {
			resolved[entity.ID] = ResolvedEntity{
				WorkspaceID: entity.WorkspaceID, TeamID: entity.TeamID, Name: entity.Name, Icon: entity.Icon, Color: entity.Color,
				URLParams: map[string]string{"view_id": entity.ID.String()}, ViewOwnerID: &entity.OwnerID, ViewShared: entity.Shared,
			}
		}
		return resolved, err
	case schema.FavoriteKindDocument:
		found, err := db.Document.Query().Where(document.IDIn(ids...), document.TrashedAtIsNil()).All(ctx)
		for _, entity := range found {
			resolved[entity.ID] = ResolvedEntity{WorkspaceID: entity.WorkspaceID, Name: entity.Title, Icon: entity.Icon, URLParams: map[string]string{"slug": entity.Slug}}
		}
		return resolved, err
	case schema.FavoriteKindLabel:
		found, err := db.Label.Query().Where(label.IDIn(ids...), label.ArchivedAtIsNil()).WithTeam().All(ctx)
		for _, entity := range found {
			params := map[string]string{"label_id": entity.ID.String()}
			if entity.Edges.Team != nil {
				params["key"] = entity.Edges.Team.Key
			}
			resolved[entity.ID] = ResolvedEntity{WorkspaceID: entity.WorkspaceID, TeamID: entity.TeamID, Name: entity.Name, Color: entity.Color, URLParams: params}
		}
		return resolved, err
	case schema.FavoriteKindTeam:
		found, err := db.Team.Query().Where(team.IDIn(ids...), team.ArchivedAtIsNil()).All(ctx)
		for _, entity := range found {
			resolved[entity.ID] = ResolvedEntity{WorkspaceID: entity.WorkspaceID, TeamID: &entity.ID, Name: entity.Name, Icon: entity.Icon, Color: entity.Color, URLParams: map[string]string{"key": entity.Key}}
		}
		return resolved, err
	case schema.FavoriteKindInitiative:
		found, err := db.Initiative.Query().Where(initiative.IDIn(ids...), initiative.TrashedAtIsNil()).All(ctx)
		for _, entity := range found {
			resolved[entity.ID] = ResolvedEntity{WorkspaceID: entity.WorkspaceID, Name: entity.Name, Icon: entity.Icon, Color: entity.Color, URLParams: map[string]string{"slug": entity.Slug}}
		}
		return resolved, err
	}
	return resolved, nil
}
