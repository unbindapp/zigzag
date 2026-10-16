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
	var validation *ent.ValidationError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errdefs.ErrInvalidInput):
		return huma.Error400BadRequest("Invalid input", err)
	case errors.As(err, &validation):
		return errdefs.WithCause(huma.Error400BadRequest("Invalid input", errors.New(validation.Name+" is out of range")), err)
	case errors.Is(err, errdefs.ErrUnauthorized):
		return huma.Error403Forbidden("Forbidden")
	case ent.IsNotFound(err) || errors.Is(err, errdefs.ErrNotFound):
		return errdefs.WithCause(huma.Error404NotFound("Not found", clientDetail(err, "Not found")), err)
	case errors.Is(err, errdefs.ErrAlreadyBootstrapped):
		return huma.Error409Conflict("Conflict", errors.New("Setup is already complete"))
	case errors.Is(err, errdefs.ErrConflict) || ent.IsConstraintError(err):
		return errdefs.WithCause(huma.Error409Conflict("Conflict", clientDetail(err, "This conflicts with an existing record")), err)
	case errors.Is(err, context.Canceled):
		return huma.NewError(StatusClientClosedRequest, "Request cancelled")
	case errors.Is(err, context.DeadlineExceeded):
		return errdefs.WithCause(huma.Error504GatewayTimeout("Timed out"), err)
	}

	return errdefs.WithCause(huma.Error500InternalServerError(genericMessage), err)
}

func clientDetail(err error, fallback string) error {
	var custom *errdefs.CustomError
	if errors.As(err, &custom) && custom.Message != "" {
		return custom
	}
	return errors.New(fallback)
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
