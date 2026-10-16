package mcpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
)

func (suite *EndToEndSuite) send(method, path string, query url.Values, body []byte) (int, string) {
	target := suite.url + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, target, reader)
	suite.Require().NoError(err)
	req.Header.Set("Authorization", "Bearer "+suite.admin)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	text, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(text)
}

func (suite *EndToEndSuite) bodySchema(op *huma.Operation) *huma.Schema {
	if op.RequestBody == nil || op.RequestBody.Content["application/json"] == nil {
		return nil
	}
	schema := op.RequestBody.Content["application/json"].Schema
	if schema.Ref != "" {
		schema = suite.api.OpenAPI().Components.Schemas.SchemaFromRef(schema.Ref)
	}
	return schema
}

func (suite *EndToEndSuite) plausibleValue(name string, property *huma.Schema) any {
	if property.Ref != "" {
		property = suite.api.OpenAPI().Components.Schemas.SchemaFromRef(property.Ref)
	}
	if len(property.Enum) > 0 {
		return property.Enum[0]
	}
	switch {
	case property.Format == "uuid":
		return uuid.NewString()
	case property.Format == "date":
		return "2026-10-16"
	case property.Format == "date-time":
		return "2026-10-16T12:00:00Z"
	case property.Format == "email":
		return "nobody@example.com"
	case strings.Contains(name, "url"):
		return "https://example.com/hook"
	case strings.Contains(name, "color"):
		return "#5e6ad2"
	}
	switch property.Type {
	case "string":
		return "Plausible"
	case "integer", "number":
		return 1
	case "boolean":
		return false
	case "array":
		if property.Items == nil {
			return []any{}
		}
		return []any{suite.plausibleValue(name, property.Items)}
	default:
		return map[string]any{}
	}
}

func wrongValue(property *huma.Schema) any {
	if property.Type == "string" || property.Ref != "" {
		return 12345
	}
	return "wrong"
}

func (suite *EndToEndSuite) TestMalformedInputNeverAnswers500() {
	for _, op := range Operations(suite.api) {
		if len(op.Security) == 0 || oapi.IsSessionOnly(op) {
			continue
		}
		var attempts []struct {
			name  string
			query url.Values
			body  []byte
		}
		add := func(name string, query url.Values, body []byte) {
			attempts = append(attempts, struct {
				name  string
				query url.Values
				body  []byte
			}{name, query, body})
		}

		if schema := suite.bodySchema(op); schema != nil {
			wrong, unknownIDs := map[string]any{}, map[string]any{}
			for name, property := range schema.Properties {
				wrong[name] = wrongValue(property)
				unknownIDs[name] = suite.plausibleValue(name, property)
			}
			wrongJSON, _ := json.Marshal(wrong)
			unknownJSON, _ := json.Marshal(unknownIDs)
			add("no body", nil, nil)
			add("empty object", nil, []byte(`{}`))
			add("array", nil, []byte(`[]`))
			add("truncated", nil, []byte(`{"a":`))
			add("wrong types", nil, wrongJSON)
			add("unknown ids", nil, unknownJSON)
		}
		if op.RequestBody == nil {
			wrong, unknownIDs := url.Values{}, url.Values{}
			for _, param := range op.Parameters {
				if param.In != "query" {
					continue
				}
				wrong.Set(param.Name, "not-valid-%%")
				if param.Schema != nil && param.Schema.Format == "uuid" {
					unknownIDs.Set(param.Name, uuid.NewString())
				}
			}
			add("no query", nil, nil)
			add("wrong query", wrong, nil)
			add("unknown ids", unknownIDs, nil)
		}

		for _, attempt := range attempts {
			status, text := suite.send(op.Method, op.Path, attempt.query, attempt.body)
			if status >= 500 {
				suite.Failf("server error on malformed input", "%s %s with %s answered %d: %s", op.Method, op.OperationID, attempt.name, status, strings.TrimSpace(text))
			}
		}
	}
}
