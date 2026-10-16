package issue_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/api/apictx"
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
		return self.deleteIssue(ctx, tx, issue, &userID)
	})
	if err != nil {
		return err
	}
	self.emitDeleted(ctx, issue)
	return nil
}

func (self *IssueService) emitDeleted(ctx context.Context, issue *ent.Issue) {
	self.emit(ctx, issue.WorkspaceID, "issue.deleted", map[string]any{
		"id":         issue.ID,
		"identifier": models.Identifier(issue.Edges.Team.Key, issue.Number),
		"team_id":    issue.TeamID,
	})
}

func (self *IssueService) deleteIssue(ctx context.Context, tx repository.TxInterface, issue *ent.Issue, actorID *uuid.UUID) error {
	children, err := self.repo.Issue().ChildIDs(ctx, tx, issue.ID)
	if err != nil {
		return err
	}
	change := schema.FieldChange{Field: "parent", From: issue.ID.String(), FromLabel: models.Identifier(issue.Edges.Team.Key, issue.Number)}
	entries := make([]issue_repo.HistoryEntry, len(children))
	for i, child := range children {
		entries[i] = issue_repo.HistoryEntry{IssueID: child, ActorID: actorID, Source: apictx.HistorySource(ctx), Changes: []schema.FieldChange{change}}
	}
	if err := self.repo.Issue().CreateHistoryMany(ctx, tx, entries); err != nil {
		return err
	}
	return self.repo.Issue().Delete(ctx, tx, issue.ID)
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
	changed := false
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		locked, err := self.repo.Issue().LockWithEdges(ctx, tx, issue.ID)
		if err != nil {
			return err
		}
		fields := issue_repo.FieldsOf(locked)
		if changed = apply(&fields, locked); !changed {
			return nil
		}
		if err := self.repo.Issue().Update(ctx, tx, issue.ID, &fields); err != nil {
			return err
		}
		return self.recordHistory(ctx, tx, issue.ID, actorOf(userID), []schema.FieldChange{{Field: field}})
	})
	if err != nil {
		return nil, err
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
