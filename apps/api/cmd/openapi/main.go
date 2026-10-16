package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/api/middleware"
	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/router"
	"github.com/unbindapp/zigzag-api/internal/api/server"
)

func main() {
	apiKeyOnly := flag.Bool("api-key-only", false, "only include operations an API key can call")
	flag.Parse()

	cfg := &config.Config{}

	r := chi.NewRouter()
	humaCfg := router.NewHumaConfig("Zigzag API", "1.0.0", cfg.CookieSecure)
	api := humachi.New(r, humaCfg)

	mw := middleware.NewMiddleware(cfg, nil, api, nil, nil)
	router.RegisterRoutes(api, &server.Server{}, mw)

	if *apiKeyOnly {
		keepAPIKeyOperations(api.OpenAPI())
	}

	spec, err := api.OpenAPI().YAML()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to render OpenAPI spec:", err)
		os.Exit(1)
	}

	if _, err := os.Stdout.Write(spec); err != nil {
		os.Exit(1)
	}
}

func keepAPIKeyOperations(spec *huma.OpenAPI) {
	bearerOnly := []map[string][]string{{"bearerAuth": {}}}

	for path, item := range spec.Paths {
		operations := []**huma.Operation{&item.Get, &item.Post, &item.Put, &item.Patch, &item.Delete}
		kept := 0
		for _, operation := range operations {
			if *operation == nil {
				continue
			}
			if len((*operation).Security) == 0 || oapi.IsSessionOnly(*operation) {
				*operation = nil
				continue
			}
			(*operation).Security = bearerOnly
			kept++
		}
		if kept == 0 {
			delete(spec.Paths, path)
		}
	}

	delete(spec.Components.SecuritySchemes, "cookieAuth")
	spec.Servers = []*huma.Server{{
		URL: "https://{domain}/api/go",
		Variables: map[string]*huma.ServerVariable{
			"domain": {Default: "zigzag.example.com", Description: "The domain of your Zigzag instance."},
		},
	}}
}
