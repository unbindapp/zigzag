package issue_service

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	issue_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/issue"
)

type GithubLink struct {
	IssueID    uuid.UUID
	Label      string
	MovedLabel string
	Attachment issue_repo.LinkedAttachmentInput
	ReplaceURL string
	State      *ent.WorkflowState
	MoveFrom   []schema.StateType
}

func (self *IssueService) ApplyGithubLink(ctx context.Context, link *GithubLink) error {
	var outcome *issueOutcome
	err := self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		current, err := self.repo.Issue().LockWithEdges(ctx, tx, link.IssueID)
		if err != nil {
			return err
		}
		outcome, err = self.applyGithubLink(ctx, tx, current, link)
		return err
	})
	if err != nil || len(outcome.changes) == 0 {
		return err
	}
	response, err := self.reload(ctx, uuid.Nil, link.IssueID)
	if err != nil {
		return err
	}
	self.announce(ctx, nil, outcome, response)
	return nil
}

func (self *IssueService) applyGithubLink(ctx context.Context, tx repository.TxInterface, current *ent.Issue, link *GithubLink) (*issueOutcome, error) {
	outcome := &issueOutcome{issueID: current.ID, workspaceID: current.WorkspaceID}
	linked, err := self.syncLinkedAttachment(ctx, tx, current.ID, link)
	if err != nil {
		return nil, err
	}
	moves := link.State != nil && current.Edges.State != nil && link.State.ID != current.StateID && slices.Contains(link.MoveFrom, current.Edges.State.Type)
	if !linked && !moves {
		return outcome, nil
	}
	label := link.Label
	if moves && link.MovedLabel != "" {
		label = link.MovedLabel
	}
	outcome.changes = append(outcome.changes, schema.FieldChange{Field: "github", To: link.Attachment.URL, ToLabel: label})
	if moves {
		if err := self.moveForGithub(ctx, tx, current, link.State, outcome); err != nil {
			return nil, err
		}
	}
	return outcome, self.repo.Issue().CreateHistory(ctx, tx, current.ID, nil, schema.HistorySourceSystem, outcome.changes)
}

func (self *IssueService) syncLinkedAttachment(ctx context.Context, tx repository.TxInterface, issueID uuid.UUID, link *GithubLink) (bool, error) {
	wanted := &link.Attachment
	urls := []string{wanted.URL}
	if link.ReplaceURL != "" && link.ReplaceURL != wanted.URL {
		urls = append(urls, link.ReplaceURL)
	}
	found, err := self.repo.Issue().AttachmentsByURL(ctx, tx, issueID, urls)
	if err != nil {
		return false, err
	}
	existing := firstByURL(found, urls)
	if existing == nil {
		_, err := self.repo.Issue().CreateAttachment(ctx, tx, &issue_repo.CreateAttachmentInput{
			IssueID: issueID, Title: wanted.Title, URL: wanted.URL, SourceType: wanted.SourceType, Metadata: wanted.Metadata,
		})
		return err == nil, err
	}
	visible := existing.Title != wanted.Title || existing.URL != wanted.URL || existing.SourceType != wanted.SourceType || existing.Metadata["state"] != wanted.Metadata["state"]
	if !visible && sameJSON(existing.Metadata, wanted.Metadata) {
		return false, nil
	}
	if _, err := self.repo.Issue().UpdateLinkedAttachment(ctx, tx, existing.ID, wanted); err != nil {
		return false, err
	}
	return visible, nil
}

func firstByURL(attachments []*ent.Attachment, urls []string) *ent.Attachment {
	for _, url := range urls {
		for _, attachment := range attachments {
			if attachment.URL == url {
				return attachment
			}
		}
	}
	return nil
}

func sameJSON(a, b map[string]any) bool {
	encodedA, errA := json.Marshal(a)
	encodedB, errB := json.Marshal(b)
	return errA == nil && errB == nil && string(encodedA) == string(encodedB)
}

func (self *IssueService) moveForGithub(ctx context.Context, tx repository.TxInterface, current *ent.Issue, state *ent.WorkflowState, outcome *issueOutcome) error {
	fields := issue_repo.FieldsOf(current)
	fields.StateID = state.ID
	applyStateTimestamps(&fields, current.Edges.State, state, self.now())
	outcome.changes = append(outcome.changes, schema.FieldChange{Field: "state", From: current.StateID.String(), To: state.ID.String(), FromLabel: current.Edges.State.Name, ToLabel: state.Name})
	outcome.stateChanged = state
	if state.Type == schema.StateTypeStarted && fields.CycleID == nil && current.Edges.Team != nil {
		cycleID, err := self.activeCycle(ctx, current.Edges.Team)
		if err != nil {
			return err
		}
		if cycleID != nil {
			label, err := self.cycleLabelOf(ctx, cycleID)
			if err != nil {
				return err
			}
			fields.CycleID = cycleID
			outcome.changes = append(outcome.changes, schema.FieldChange{Field: "cycle", To: cycleID.String(), ToLabel: label})
		}
	}
	return self.repo.Issue().Update(ctx, tx, current.ID, &fields)
}
