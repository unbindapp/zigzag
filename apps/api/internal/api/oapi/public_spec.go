package oapi

import (
	"maps"

	"github.com/danielgtaylor/huma/v2"
)

func APIKeySpec(spec *huma.OpenAPI) *huma.OpenAPI {
	bearerOnly := []map[string][]string{{"bearerAuth": {}}}
	public := *spec
	public.Paths = map[string]*huma.PathItem{}
	for path, item := range spec.Paths {
		kept := *item
		operations := []**huma.Operation{&kept.Get, &kept.Post, &kept.Put, &kept.Patch, &kept.Delete}
		count := 0
		for _, operation := range operations {
			if *operation == nil {
				continue
			}
			if len((*operation).Security) == 0 || IsSessionOnly(*operation) {
				*operation = nil
				continue
			}
			copied := **operation
			copied.Security = bearerOnly
			*operation = &copied
			count++
		}
		if count > 0 {
			public.Paths[path] = &kept
		}
	}
	if spec.Components != nil {
		components := *spec.Components
		components.SecuritySchemes = maps.Clone(spec.Components.SecuritySchemes)
		delete(components.SecuritySchemes, "cookieAuth")
		public.Components = &components
	}
	return &public
}
