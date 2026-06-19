package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Locale struct {
	ent.Schema
}

func (Locale) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Locale) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").Unique().MaxLen(8),
		field.String("name").NotEmpty(),
		field.Bool("is_active").Default(true),
	}
}

func (Locale) Edges() []ent.Edge {
	return nil
}

func (Locale) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
	}
}
