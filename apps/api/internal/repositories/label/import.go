package label_repo

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

type ImportedLabel struct {
	ID                uuid.UUID
	TeamID            *uuid.UUID
	ParentID          *uuid.UUID
	Name              string
	Color             string
	Description       string
	IsGroup           bool
	GroupSingleSelect bool
	SortOrder         float64
	ArchivedAt        *time.Time
}

// ImportLabels expects every group before its children.
func (self *LabelRepository) ImportLabels(ctx context.Context, tx repository.TxInterface, workspaceID uuid.UUID, labels []ImportedLabel) error {
	db := self.base.Client(tx)
	builders := make([]*ent.LabelCreate, len(labels))
	for i, l := range labels {
		builders[i] = db.Label.Create().
			SetID(l.ID).
			SetWorkspaceID(workspaceID).
			SetNillableTeamID(l.TeamID).
			SetNillableParentID(l.ParentID).
			SetName(l.Name).
			SetColor(l.Color).
			SetDescription(l.Description).
			SetIsGroup(l.IsGroup).
			SetGroupSingleSelect(l.GroupSingleSelect).
			SetSortOrder(l.SortOrder).
			SetNillableArchivedAt(l.ArchivedAt)
	}
	for batch := range slices.Chunk(builders, 500) {
		if err := db.Label.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
