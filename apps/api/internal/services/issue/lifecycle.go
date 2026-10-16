package issue_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
)

type IssueRefInput struct {
	IssueID uuid.UUID `json:"issue_id" required:"true" format:"uuid" doc:"The issue"`
}

func (self *IssueService) Delete(ctx context.Context, userID uuid.UUID, input *IssueRefInput) error {
	issue, err := self.WritableIssue(ctx, userID, input.IssueID, schema.WorkspaceRoleAdmin)
	if err != nil {
		return err
	}
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		return self.repo.Issue().Delete(ctx, tx, issue.ID)
	})
	if err != nil {
		return err
	}
	self.emit(ctx, issue.WorkspaceID, "issue.deleted", map[string]any{
		"id":         issue.ID,
		"identifier": models.Identifier(issue.Edges.Team.Key, issue.Number),
		"team_id":    issue.TeamID,
	})
	return nil
}

func (self *IssueService) Trash(ctx context.Context, userID uuid.UUID, input *IssueRefInput) (*models.IssueResponse, error) {
	return self.setLifecycle(ctx, userID, input.IssueID, "trashed", func(fields *issue_repo.IssueFields, issue *ent.Issue) bool {
		if issue.TrashedAt != nil {
			return false
		}
		now := self.now()
		fields.TrashedAt = &now
		return true
	})
}

func (self *IssueService) Restore(ctx context.Context, userID uuid.UUID, input *IssueRefInput) (*models.IssueResponse, error) {
	return self.setLifecycle(ctx, userID, input.IssueID, "restored", func(fields *issue_repo.IssueFields, issue *ent.Issue) bool {
		if issue.TrashedAt == nil {
			return false
		}
		fields.TrashedAt = nil
		return true
	})
}

func (self *IssueService) Archive(ctx context.Context, userID uuid.UUID, input *IssueRefInput) (*models.IssueResponse, error) {
	return self.setLifecycle(ctx, userID, input.IssueID, "archived", func(fields *issue_repo.IssueFields, issue *ent.Issue) bool {
		if issue.ArchivedAt != nil {
			return false
		}
		now := self.now()
		fields.ArchivedAt = &now
		return true
	})
}

func (self *IssueService) Unarchive(ctx context.Context, userID uuid.UUID, input *IssueRefInput) (*models.IssueResponse, error) {
	return self.setLifecycle(ctx, userID, input.IssueID, "unarchived", func(fields *issue_repo.IssueFields, issue *ent.Issue) bool {
		if issue.ArchivedAt == nil {
			return false
		}
		fields.ArchivedAt, fields.AutoArchivedAt = nil, nil
		return true
	})
}

func (self *IssueService) setLifecycle(ctx context.Context, userID, issueID uuid.UUID, field string, apply func(*issue_repo.IssueFields, *ent.Issue) bool) (*models.IssueResponse, error) {
	issue, err := self.WritableIssue(ctx, userID, issueID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	fields := issue_repo.FieldsOf(issue)
	changed := apply(&fields, issue)
	if changed {
		changes := []schema.FieldChange{{Field: field}}
		err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
			if err := self.repo.Issue().Update(ctx, tx, issue.ID, &fields); err != nil {
				return err
			}
			return self.repo.Issue().CreateHistory(ctx, tx, issue.ID, &userID, historySource(ctx), changes)
		})
		if err != nil {
			return nil, err
		}
	}
	response, err := self.reload(ctx, userID, issue.ID)
	if err != nil {
		return nil, err
	}
	if changed {
		self.emit(ctx, issue.WorkspaceID, "issue.updated", map[string]any{"issue": response, "changes": []schema.FieldChange{{Field: field}}})
	}
	return response, nil
}
