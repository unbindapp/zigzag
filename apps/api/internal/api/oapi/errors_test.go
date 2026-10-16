package oapi

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
)

func init() {
	huma.NewError = errdefs.HumaErrorFunc
}

func mapped(t *testing.T, err error) *errdefs.ResponseError {
	t.Helper()
	resp, ok := MapError(err).(*errdefs.ResponseError)
	if !ok {
		t.Fatalf("MapError(%v) did not return a ResponseError", err)
	}
	return resp
}

func TestMapErrorStatuses(t *testing.T) {
	tests := map[string]struct {
		err     error
		status  int
		message string
	}{
		"invalid input":     {errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "bad key"), 400, "Invalid input"},
		"unauthorized":      {errdefs.ErrUnauthorized, 403, "Forbidden"},
		"not found":         {errdefs.NewCustomError(errdefs.ErrTypeNotFound, "no team"), 404, "Not found"},
		"ent not found":     {fmt.Errorf("loading: %w", &ent.NotFoundError{}), 404, "Not found"},
		"conflict":          {errdefs.NewCustomError(errdefs.ErrTypeConflict, "taken"), 409, "Conflict"},
		"bootstrapped":      {errdefs.ErrAlreadyBootstrapped, 409, "Conflict"},
		"ent constraint":    {&ent.ConstraintError{}, 409, "Conflict"},
		"cancelled":         {fmt.Errorf("query: %w", context.Canceled), StatusClientClosedRequest, "Request cancelled"},
		"deadline exceeded": {fmt.Errorf("query: %w", context.DeadlineExceeded), 504, "Timed out"},
		"unknown":           {errors.New("boom"), 500, "Internal server error"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp := mapped(t, tt.err)
			if resp.Status != tt.status || resp.Message != tt.message {
				t.Fatalf("got %d %q, want %d %q", resp.Status, resp.Message, tt.status, tt.message)
			}
		})
	}
}

func TestMapErrorNil(t *testing.T) {
	if MapError(nil) != nil {
		t.Fatal("nil must stay nil")
	}
}

func TestCustomErrorMessageIsTheDetail(t *testing.T) {
	resp := mapped(t, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Key must be uppercase"))
	if len(resp.Details) != 1 || resp.Details[0] != "Key must be uppercase" {
		t.Fatalf("details = %v", resp.Details)
	}
}

func TestInternalErrorSummaryReplacesGenericMessage(t *testing.T) {
	cause := errors.New("connection reset")
	resp := mapped(t, errdefs.NewInternalError(cause, "Failed to store the session"))
	if resp.Status != 500 || resp.Message != "Failed to store the session" {
		t.Fatalf("got %d %q", resp.Status, resp.Message)
	}
	if len(resp.Details) != 0 {
		t.Fatalf("a generic classification adds no detail, got %v", resp.Details)
	}
	if !errors.Is(resp.Cause(), cause) {
		t.Fatal("the cause is kept for the logs")
	}
}

func TestInternalErrorKeepsSpecificClassification(t *testing.T) {
	resp := mapped(t, errdefs.NewInternalError(&ent.NotFoundError{}, "Failed to load the team"))
	if resp.Status != 404 || resp.Message != "Failed to load the team" {
		t.Fatalf("got %d %q", resp.Status, resp.Message)
	}
	if len(resp.Details) != 1 {
		t.Fatalf("the underlying failure stays as a detail, got %v", resp.Details)
	}
}

func TestDatabaseTextNeverReachesTheClient(t *testing.T) {
	tests := map[string]struct {
		err    error
		status int
		detail string
	}{
		"constraint": {
			errors.Join(errors.New("committing transaction"), &ent.ConstraintError{}), 409, "This conflicts with an existing record",
		},
		"ent not found": {fmt.Errorf("error fetching user: %w", &ent.NotFoundError{}), 404, "Not found"},
		"validator":     {&ent.ValidationError{Name: "title"}, 400, "title is out of range"},
		"custom kept":   {errdefs.NewCustomError(errdefs.ErrTypeConflict, "The key ENG is taken"), 409, "The key ENG is taken"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp := mapped(t, tt.err)
			if resp.Status != tt.status || len(resp.Details) != 1 || resp.Details[0] != tt.detail {
				t.Fatalf("got %d %v, want %d [%s]", resp.Status, resp.Details, tt.status, tt.detail)
			}
			if resp.Cause() == nil && name != "custom kept" {
				t.Fatal("the cause is kept for the logs")
			}
		})
	}
}
