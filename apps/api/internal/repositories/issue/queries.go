package issue_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/issue"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateIssueInput struct {
	WorkspaceID     uuid.UUID
	TeamID          uuid.UUID
	Number          int
	Title           string
	Description     string
	DescriptionText string
	Priority        int
	StateID         uuid.UUID
	AssigneeID      *uuid.UUID
	CreatorID       *uuid.UUID
	SortOrder       float64
}

func (self *IssueRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateIssueInput) (*ent.Issue, error) {
	return self.base.Client(tx).Issue.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetTeamID(input.TeamID).
		SetNumber(input.Number).
		SetTitle(input.Title).
		SetDescription(input.Description).
		SetDescriptionText(input.DescriptionText).
		SetPriority(input.Priority).
		SetStateID(input.StateID).
		SetNillableAssigneeID(input.AssigneeID).
		SetNillableCreatorID(input.CreatorID).
		SetSortOrder(input.SortOrder).
		Save(ctx)
}

func (self *IssueRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Issue, error) {
	return self.base.DB.Issue.Get(ctx, id)
}

func (self *IssueRepository) ResolveWorkspaceOfIssue(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Issue.Query().Where(issue.ID(id)).Select(issue.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
