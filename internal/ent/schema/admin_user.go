package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type AdminUser struct {
	ent.Schema
}

func (AdminUser) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (AdminUser) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").Unique().NotEmpty(),
		field.String("password_hash").Sensitive().NotEmpty(),
		field.String("first_name").Optional().Default(""),
		field.String("last_name").Optional().Default(""),
		field.Enum("status").Values("active", "disabled").Default("active"),
		field.Time("last_login_at").Optional().Nillable(),
	}
}

func (AdminUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roles", Role.Type).Ref("admins"),
	}
}

func (AdminUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}
