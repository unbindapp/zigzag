package oapi

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
)

const StatusClientClosedRequest = 499

const genericMessage = "Internal server error"

func MapError(err error) error {
	var internal *errdefs.InternalError

	summary := ""
	if errors.As(err, &internal) {
		summary = internal.Summary
	}

	return withSummary(classifyError(err), summary)
}

func classifyError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errdefs.ErrInvalidInput):
		return huma.Error400BadRequest("Invalid input", err)
	case errors.Is(err, errdefs.ErrUnauthorized):
		return huma.Error403Forbidden("Forbidden")
	case ent.IsNotFound(err) || errors.Is(err, errdefs.ErrNotFound):
		return huma.Error404NotFound("Not found", err)
	case errors.Is(err, errdefs.ErrConflict) ||
		errors.Is(err, errdefs.ErrAlreadyBootstrapped) ||
		ent.IsConstraintError(err):
		return huma.Error409Conflict("Conflict", err)
	case errors.Is(err, context.Canceled):
		return huma.NewError(StatusClientClosedRequest, "Request cancelled")
	case errors.Is(err, context.DeadlineExceeded):
		return errdefs.WithCause(huma.Error504GatewayTimeout("Timed out"), err)
	}

	return errdefs.WithCause(huma.Error500InternalServerError(genericMessage), err)
}

func withSummary(mapped error, summary string) error {
	resp, ok := mapped.(*errdefs.ResponseError)
	if summary == "" || !ok {
		return mapped
	}

	if len(resp.Details) == 0 && resp.Message != "" && resp.Message != genericMessage {
		resp.Details = append(resp.Details, resp.Message)
	}
	resp.Message = summary
	return resp
}
