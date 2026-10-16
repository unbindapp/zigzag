package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type JWTKey struct {
	ent.Schema
}

func (JWTKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("label"),
		field.Bytes("private_key"),
	}
}

func (JWTKey) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "jwt_keys",
		},
	}
}
