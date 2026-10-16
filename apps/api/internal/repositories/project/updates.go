package project_repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/projectpost"
	"github.com/zigzagapp-org/zigzag-api/ent/reaction"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

func withUpdateEdges(q *ent.ProjectPostQuery) *ent.ProjectPostQuery {
	return q.WithAuthor().WithReactions(func(q *ent.ReactionQuery) { q.WithUser().Order(reaction.ByCreatedAt()) })
}

func (self *ProjectRepository) ListUpdates(ctx context.Context, projectID uuid.UUID) ([]*ent.ProjectPost, error) {
	return withUpdateEdges(self.base.DB.ProjectPost.Query().Where(projectpost.ProjectID(projectID))).
		Order(projectpost.ByCreatedAt(sql.OrderDesc())).
		All(ctx)
}

func (self *ProjectRepository) GetUpdate(ctx context.Context, updateID uuid.UUID) (*ent.ProjectPost, error) {
	return withUpdateEdges(self.base.DB.ProjectPost.Query().Where(projectpost.ID(updateID))).Only(ctx)
}

type CreateUpdateInput struct {
	ProjectID uuid.UUID
	AuthorID  uuid.UUID
	Body      string
	Health    schema.Health
}

func (self *ProjectRepository) CreateUpdate(ctx context.Context, tx repository.TxInterface, input *CreateUpdateInput) (*ent.ProjectPost, error) {
	return self.base.Client(tx).ProjectPost.Create().
		SetProjectID(input.ProjectID).
		SetAuthorID(input.AuthorID).
		SetBody(input.Body).
		SetHealth(input.Health).
		Save(ctx)
}

type EditUpdateInput struct {
	Body     *string
	Health   *schema.Health
	EditedAt time.Time
}

func (self *ProjectRepository) EditUpdate(ctx context.Context, tx repository.TxInterface, updateID uuid.UUID, input *EditUpdateInput) error {
	m := self.base.Client(tx).ProjectPost.UpdateOneID(updateID).SetEditedAt(input.EditedAt)
	if input.Body != nil {
		m.SetBody(*input.Body)
	}
	if input.Health != nil {
		m.SetHealth(*input.Health)
	}
	return m.Exec(ctx)
}

func (self *ProjectRepository) DeleteUpdate(ctx context.Context, tx repository.TxInterface, updateID uuid.UUID) error {
	return self.base.Client(tx).ProjectPost.DeleteOneID(updateID).Exec(ctx)
}

func (self *ProjectRepository) LatestHealth(ctx context.Context, tx repository.TxInterface, projectID uuid.UUID) (*schema.Health, error) {
	latest, err := self.base.Client(tx).ProjectPost.Query().
		Where(projectpost.ProjectID(projectID)).
		Order(projectpost.ByCreatedAt(sql.OrderDesc())).
		Select(projectpost.FieldHealth).
		First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &latest.Health, nil
}
