package mcpserver

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/internal/api/middleware"
	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/router"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
)

func realAPI() huma.API {
	cfg := &config.Config{}
	api := humachi.New(chi.NewRouter(), router.NewHumaConfig("Zigzag API", "test", cfg.CookieSecure))
	router.RegisterRoutes(api, &server.Server{}, middleware.NewMiddleware(cfg, nil, api, nil, nil))
	return api
}

func TestEveryOperationDecidesOnMCP(t *testing.T) {
	for _, op := range Operations(realAPI()) {
		choice, decided := oapi.MCPChoiceOf(op)
		unreachable := len(op.Security) == 0 || oapi.IsSessionOnly(op)

		if unreachable && choice.Exposed {
			t.Errorf("%s is public or session only, it cannot be registered with oapi.MCP", op.OperationID)
		}
		if unreachable {
			continue
		}
		if !decided {
			t.Errorf("%s must be registered with oapi.MCP or oapi.NoMCP(reason)", op.OperationID)
			continue
		}
		if !choice.Exposed && choice.Reason == "" {
			t.Errorf("%s opts out of MCP without a reason", op.OperationID)
		}
	}
}

func TestRealOperationsBuildValidTools(t *testing.T) {
	tools, err := buildTools(realAPI())
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) == 0 {
		t.Fatal("no operation is registered with oapi.MCP")
	}

	for _, tool := range tools {
		raw, err := json.Marshal(tool.definition.InputSchema)
		if err != nil {
			t.Fatalf("%s: %v", tool.definition.Name, err)
		}
		var schema jsonschema.Schema
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatalf("%s: input schema is not valid JSON Schema: %v", tool.definition.Name, err)
		}
		if _, err := schema.Resolve(nil); err != nil {
			t.Errorf("%s: input schema does not resolve: %v", tool.definition.Name, err)
		}
		if tool.definition.Description == "" {
			t.Errorf("%s: a tool needs a description", tool.definition.Name)
		}
	}
}

func TestPublicSpecListsOnlyWhatAnAPIKeyCanCall(t *testing.T) {
	api := realAPI()
	full := api.OpenAPI()
	before := len(Operations(api))
	public := oapi.APIKeySpec(full)

	reachable := 0
	for _, op := range Operations(api) {
		if len(op.Security) > 0 && !oapi.IsSessionOnly(op) {
			reachable++
		}
	}
	kept := 0
	for path, item := range public.Paths {
		for _, op := range []*huma.Operation{item.Get, item.Post, item.Put, item.Patch, item.Delete} {
			if op == nil {
				continue
			}
			kept++
			if oapi.IsSessionOnly(op) || len(op.Security) != 1 || op.Security[0]["bearerAuth"] == nil {
				t.Errorf("%s %s: the public spec keeps only bearer operations an API key can call", path, op.OperationID)
			}
		}
	}
	if kept != reachable || kept == 0 {
		t.Errorf("public spec has %d operations, want the %d an API key can reach", kept, reachable)
	}
	if _, ok := public.Components.SecuritySchemes["cookieAuth"]; ok {
		t.Error("the public spec must not describe cookie auth")
	}
	if _, ok := full.Components.SecuritySchemes["cookieAuth"]; !ok || len(Operations(api)) != before {
		t.Error("building the public spec must leave the full spec intact")
	}
	if full.Paths["/users/whoami"].Get.Security[0]["cookieAuth"] == nil {
		t.Error("the full spec keeps cookie auth on its operations")
	}
}

func TestActionsFollowTheHTTPVerb(t *testing.T) {
	for _, op := range Operations(realAPI()) {
		action, ok := oapi.ActionOf(op)
		if !ok {
			t.Errorf("%s is not registered through oapi.Register", op.OperationID)
			continue
		}
		var want []oapi.Action
		switch op.Method {
		case http.MethodGet:
			want = []oapi.Action{oapi.Read}
		case http.MethodPut:
			want = []oapi.Action{oapi.Update}
		case http.MethodDelete:
			want = []oapi.Action{oapi.Delete}
		case http.MethodPost:
			want = []oapi.Action{oapi.Invoke}
			if strings.HasSuffix(op.Path, "/create") || strings.HasSuffix(op.Path, "/add") || strings.HasSuffix(op.Path, "/create-user") {
				want = []oapi.Action{oapi.Create}
			}
		}
		if !slices.Contains(want, action) {
			t.Errorf("%s %s %s uses action %d, want %v", op.OperationID, op.Method, op.Path, action, want)
		}
	}
}

func TestEveryInputIsDocumentedAndComesOneWay(t *testing.T) {
	api := realAPI()
	registry := api.OpenAPI().Components.Schemas
	for _, op := range Operations(api) {
		query := 0
		for _, param := range op.Parameters {
			if param.In == "header" || param.In == "cookie" {
				continue
			}
			query++
			if param.Description == "" {
				t.Errorf("%s: query parameter %s needs a doc tag", op.OperationID, param.Name)
			}
		}
		if op.RequestBody == nil {
			continue
		}
		if _, raw := op.RequestBody.Content["application/octet-stream"]; query > 0 && !raw {
			t.Errorf("%s takes both query parameters and a body", op.OperationID)
		}
		if _, multipart := op.RequestBody.Content["multipart/form-data"]; !multipart && !op.RequestBody.Required {
			t.Errorf("%s: the request body must be required", op.OperationID)
		}
		for _, media := range op.RequestBody.Content {
			schema := media.Schema
			if schema.Ref != "" {
				schema = registry.SchemaFromRef(schema.Ref)
			}
			for name, property := range schema.Properties {
				if name == "$schema" || property.Description != "" {
					continue
				}
				if property.Ref != "" && registry.SchemaFromRef(property.Ref).Description != "" {
					continue
				}
				t.Errorf("%s: body field %s needs a doc tag", op.OperationID, name)
			}
		}
	}
}
