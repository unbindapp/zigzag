package project_service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/markdown"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	project_repo "github.com/unbindapp/zigzag-api/internal/repositories/project"
)

const snippetLength = 200

func (self *ProjectService) ListUpdates(ctx context.Context, userID uuid.UUID, input *ProjectQueryInput) ([]*models.ProjectUpdateResponse, error) {
	project, err := self.projectWithRole(ctx, userID, input.ProjectID, schema.WorkspaceRoleViewer)
	if err != nil {
		return nil, err
	}
	updates, err := self.repo.Project().ListUpdates(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	return models.TransformProjectUpdates(updates, userID), nil
}

type CreateProjectUpdateInput struct {
	ProjectID uuid.UUID     `json:"project_id" required:"true" format:"uuid" doc:"The project"`
	Body      string        `json:"body" required:"true" minLength:"1" maxLength:"50000" doc:"Markdown text of the update"`
	Health    schema.Health `json:"health" required:"true" doc:"on_track, at_risk or off_track; becomes the project's health"`
}

func (self *ProjectService) CreateUpdate(ctx context.Context, userID uuid.UUID, input *CreateProjectUpdateInput) (*models.ProjectUpdateResponse, error) {
	project, err := self.projectWithRole(ctx, userID, input.ProjectID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	body, err := updateBody(input.Body)
	if err != nil {
		return nil, err
	}
	if !validHealth(input.Health) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "health must be on_track, at_risk or off_track")
	}
	var created *ent.ProjectPost
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		created, err = self.repo.Project().CreateUpdate(ctx, tx, &project_repo.CreateUpdateInput{
			ProjectID: project.ID, AuthorID: userID, Body: body, Health: input.Health,
		})
		if err != nil {
			return err
		}
		return self.repo.Project().SetHealth(ctx, tx, project.ID, &input.Health)
	})
	if err != nil {
		return nil, err
	}
	update, err := self.repo.Project().GetUpdate(ctx, created.ID)
	if err != nil {
		return nil, err
	}
	self.notifyUpdate(ctx, userID, project, update)
	return models.TransformProjectUpdate(update, userID), nil
}

type EditProjectUpdateInput struct {
	UpdateID uuid.UUID      `json:"update_id" required:"true" format:"uuid" doc:"The project update to change"`
	Body     *string        `json:"body,omitempty" required:"false" maxLength:"50000" doc:"New markdown text"`
	Health   *schema.Health `json:"health,omitempty" required:"false" doc:"on_track, at_risk or off_track"`
}

func (self *ProjectService) EditUpdate(ctx context.Context, userID uuid.UUID, input *EditProjectUpdateInput) (*models.ProjectUpdateResponse, error) {
	update, err := self.authorOrAdminUpdate(ctx, userID, input.UpdateID)
	if err != nil {
		return nil, err
	}
	edit := &project_repo.EditUpdateInput{Health: input.Health, EditedAt: self.now()}
	if input.Body != nil {
		body, err := updateBody(*input.Body)
		if err != nil {
			return nil, err
		}
		edit.Body = &body
	}
	if input.Health != nil && !validHealth(*input.Health) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "health must be on_track, at_risk or off_track")
	}
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.repo.Project().EditUpdate(ctx, tx, update.ID, edit); err != nil {
			return err
		}
		return self.refreshHealth(ctx, tx, update.ProjectID)
	})
	if err != nil {
		return nil, err
	}
	edited, err := self.repo.Project().GetUpdate(ctx, update.ID)
	if err != nil {
		return nil, err
	}
	return models.TransformProjectUpdate(edited, userID), nil
}

type ProjectUpdateRefInput struct {
	UpdateID uuid.UUID `json:"update_id" required:"true" format:"uuid" doc:"The project update to delete"`
}

func (self *ProjectService) DeleteUpdate(ctx context.Context, userID uuid.UUID, input *ProjectUpdateRefInput) error {
	update, err := self.authorOrAdminUpdate(ctx, userID, input.UpdateID)
	if err != nil {
		return err
	}
	return self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.repo.Project().DeleteUpdate(ctx, tx, update.ID); err != nil {
			return err
		}
		return self.refreshHealth(ctx, tx, update.ProjectID)
	})
}

func (self *ProjectService) refreshHealth(ctx context.Context, tx repository.TxInterface, projectID uuid.UUID) error {
	health, err := self.repo.Project().LatestHealth(ctx, tx, projectID)
	if err != nil {
		return err
	}
	return self.repo.Project().SetHealth(ctx, tx, projectID, health)
}

func (self *ProjectService) authorOrAdminUpdate(ctx context.Context, userID, updateID uuid.UUID) (*ent.ProjectPost, error) {
	update, err := self.repo.Project().GetUpdate(ctx, updateID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Project update not found")
	}
	if err != nil {
		return nil, err
	}
	role := schema.WorkspaceRoleAdmin
	if update.AuthorID != nil && *update.AuthorID == userID {
		role = schema.WorkspaceRoleMember
	}
	if err := self.checkProjectRole(ctx, userID, update.ProjectID, role, "Project update not found"); err != nil {
		return nil, err
	}
	return update, nil
}

func (self *ProjectService) notifyUpdate(ctx context.Context, actorID uuid.UUID, project *ent.Project, update *ent.ProjectPost) {
	audience, err := self.repo.Project().AudienceIDs(ctx, project.ID)
	if err != nil {
		return
	}
	self.notify(ctx, models.NotificationEvent{
		Type:            schema.NotificationProjectUpdate,
		WorkspaceID:     project.WorkspaceID,
		ActorID:         &actorID,
		RecipientIDs:    without(audience, actorID),
		ProjectID:       &project.ID,
		ProjectUpdateID: &update.ID,
		Title:           "New update on " + project.Name + ": " + healthLabel(update.Health),
		Body:            snippet(update.Body),
	})
}

func updateBody(raw string) (string, error) {
	body := strings.TrimSpace(raw)
	if body == "" {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "An update needs a body")
	}
	return body, nil
}

func healthLabel(health schema.Health) string {
	switch health {
	case schema.HealthOnTrack:
		return "On track"
	case schema.HealthAtRisk:
		return "At risk"
	case schema.HealthOffTrack:
		return "Off track"
	default:
		return string(health)
	}
}

func snippet(body string) string {
	text := []rune(markdown.PlainText(body))
	if len(text) <= snippetLength {
		return string(text)
	}
	return strings.TrimSpace(string(text[:snippetLength])) + "…"
}

func without(ids []uuid.UUID, excluded uuid.UUID) []uuid.UUID {
	kept := []uuid.UUID{}
	for _, id := range ids {
		if id != excluded {
			kept = append(kept, id)
		}
	}
	return kept
}
