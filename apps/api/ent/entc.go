//go:build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	features := []gen.Feature{gen.FeatureExecQuery, gen.FeatureModifier, gen.FeatureUpsert, gen.FeatureVersionedMigration}
	if err := entc.Generate("./schema", &gen.Config{Features: features}); err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
