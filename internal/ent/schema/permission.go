package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Permission struct {
	ent.Schema
}

func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("resource").NotEmpty(),
		field.String("action").NotEmpty(),
		field.String("description").Optional().Default(""),
	}
}

func (Permission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roles", Role.Type).Ref("permissions"),
	}
}

func (Permission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource", "action").Unique(),
	}
}
