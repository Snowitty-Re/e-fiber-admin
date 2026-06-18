package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Role struct {
	ent.Schema
}

func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("slug").Unique(),
		field.String("description").Optional().Default(""),
		field.Bool("is_system").Default(false),
	}
}

func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("admins", AdminUser.Type),
		edge.To("permissions", Permission.Type),
	}
}

func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_system"),
	}
}
