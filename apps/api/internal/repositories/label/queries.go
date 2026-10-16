package label_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/label"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateLabelInput struct {
	WorkspaceID       uuid.UUID
	TeamID            *uuid.UUID
	Name              string
	Color             string
	Description       string
	ParentID          *uuid.UUID
	IsGroup           bool
	GroupSingleSelect bool
}

func (self *LabelRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateLabelInput) (*ent.Label, error) {
	return self.base.Client(tx).Label.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetNillableTeamID(input.TeamID).
		SetName(input.Name).
		SetColor(input.Color).
		SetDescription(input.Description).
		SetNillableParentID(input.ParentID).
		SetIsGroup(input.IsGroup).
		SetGroupSingleSelect(input.GroupSingleSelect).
		Save(ctx)
}

func (self *LabelRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Label, error) {
	return self.base.DB.Label.Get(ctx, id)
}

func (self *LabelRepository) ResolveWorkspaceOfLabel(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Label.Query().Where(label.ID(id)).Select(label.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
