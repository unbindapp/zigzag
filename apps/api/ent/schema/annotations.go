package schema

import (
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

func table(name string) []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: name}}
}

func cascade() *entsql.Annotation { return entsql.OnDelete(entsql.Cascade) }

func setNull() *entsql.Annotation { return entsql.OnDelete(entsql.SetNull) }

func noAction() *entsql.Annotation { return entsql.OnDelete(entsql.NoAction) }

var dateType = map[string]string{dialect.Postgres: "date"}
