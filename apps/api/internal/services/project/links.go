package project_service

import (
	"context"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	project_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/project"
)

type CreateProjectLinkInput struct {
	ProjectID uuid.UUID `json:"project_id" required:"true" format:"uuid" doc:"The project"`
	URL       string    `json:"url" required:"true" minLength:"1" maxLength:"2048" doc:"http or https URL"`
	Label     string    `json:"label" required:"true" minLength:"1" maxLength:"120" doc:"Text shown for the link"`
}

func (self *ProjectService) CreateLink(ctx context.Context, userID uuid.UUID, input *CreateProjectLinkInput) (*models.ProjectLinkResponse, error) {
	project, err := self.projectWithRole(ctx, userID, input.ProjectID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	link, err := parseLinkURL(input.URL)
	if err != nil {
		return nil, err
	}
	label := strings.TrimSpace(input.Label)
	if label == "" {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "A link needs a label")
	}
	sortOrder, err := self.repo.Project().NextLinkSortOrder(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	created, err := self.repo.Project().CreateLink(ctx, &project_repo.CreateLinkInput{
		ProjectID: project.ID, URL: link.String(), Label: label, CreatorID: userID, SortOrder: sortOrder,
	})
	if err != nil {
		return nil, err
	}
	return models.TransformProjectLink(created), nil
}

type UpdateProjectLinkInput struct {
	LinkID    uuid.UUID `json:"link_id" required:"true" format:"uuid" doc:"The link to change"`
	URL       *string   `json:"url,omitempty" required:"false" maxLength:"2048" doc:"New http or https URL"`
	Label     *string   `json:"label,omitempty" required:"false" maxLength:"120" doc:"New label"`
	SortOrder *float64  `json:"sort_order,omitempty" required:"false" doc:"Order among the project's links"`
}

func (self *ProjectService) UpdateLink(ctx context.Context, userID uuid.UUID, input *UpdateProjectLinkInput) (*models.ProjectLinkResponse, error) {
	link, err := self.memberLink(ctx, userID, input.LinkID)
	if err != nil {
		return nil, err
	}
	update := &project_repo.UpdateLinkInput{SortOrder: input.SortOrder}
	if input.URL != nil {
		parsed, err := parseLinkURL(*input.URL)
		if err != nil {
			return nil, err
		}
		raw := parsed.String()
		update.URL = &raw
	}
	if input.Label != nil {
		label := strings.TrimSpace(*input.Label)
		if label == "" {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "A link needs a label")
		}
		update.Label = &label
	}
	updated, err := self.repo.Project().UpdateLink(ctx, link.ID, update)
	if err != nil {
		return nil, err
	}
	return models.TransformProjectLink(updated), nil
}

type ProjectLinkRefInput struct {
	LinkID uuid.UUID `json:"link_id" required:"true" format:"uuid" doc:"The link to delete"`
}

func (self *ProjectService) DeleteLink(ctx context.Context, userID uuid.UUID, input *ProjectLinkRefInput) error {
	link, err := self.memberLink(ctx, userID, input.LinkID)
	if err != nil {
		return err
	}
	return self.repo.Project().DeleteLink(ctx, link.ID)
}

func (self *ProjectService) memberLink(ctx context.Context, userID, linkID uuid.UUID) (*ent.ProjectLink, error) {
	link, err := self.repo.Project().GetLink(ctx, linkID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Link not found")
	}
	if err != nil {
		return nil, err
	}
	if err := self.checkProjectRole(ctx, userID, link.ProjectID, schema.WorkspaceRoleMember, "Link not found"); err != nil {
		return nil, err
	}
	return link, nil
}

func parseLinkURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "url must be an http or https URL")
	}
	return parsed, nil
}
