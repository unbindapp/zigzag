package project_repo

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/projectlink"
)

func (self *ProjectRepository) GetLink(ctx context.Context, linkID uuid.UUID) (*ent.ProjectLink, error) {
	return self.base.DB.ProjectLink.Get(ctx, linkID)
}

func (self *ProjectRepository) NextLinkSortOrder(ctx context.Context, projectID uuid.UUID) (float64, error) {
	last, err := self.base.DB.ProjectLink.Query().
		Where(projectlink.ProjectID(projectID)).
		Order(projectlink.BySortOrder(sql.OrderDesc())).
		First(ctx)
	if ent.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return last.SortOrder + 1, nil
}

type CreateLinkInput struct {
	ProjectID uuid.UUID
	URL       string
	Label     string
	CreatorID uuid.UUID
	SortOrder float64
}

func (self *ProjectRepository) CreateLink(ctx context.Context, input *CreateLinkInput) (*ent.ProjectLink, error) {
	return self.base.DB.ProjectLink.Create().
		SetProjectID(input.ProjectID).
		SetURL(input.URL).
		SetLabel(input.Label).
		SetCreatorID(input.CreatorID).
		SetSortOrder(input.SortOrder).
		Save(ctx)
}

type UpdateLinkInput struct {
	URL       *string
	Label     *string
	SortOrder *float64
}

func (self *ProjectRepository) UpdateLink(ctx context.Context, linkID uuid.UUID, input *UpdateLinkInput) (*ent.ProjectLink, error) {
	m := self.base.DB.ProjectLink.UpdateOneID(linkID)
	if input.URL != nil {
		m.SetURL(*input.URL)
	}
	if input.Label != nil {
		m.SetLabel(*input.Label)
	}
	if input.SortOrder != nil {
		m.SetSortOrder(*input.SortOrder)
	}
	return m.Save(ctx)
}

func (self *ProjectRepository) DeleteLink(ctx context.Context, linkID uuid.UUID) error {
	return self.base.DB.ProjectLink.DeleteOneID(linkID).Exec(ctx)
}
