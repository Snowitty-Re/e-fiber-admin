package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Currency struct {
	ent.Schema
}

func (Currency) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Currency) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").Unique().MaxLen(3),
		field.String("name").NotEmpty(),
		field.String("symbol").Default(""),
		field.Int("precision").Default(2),
		field.Bool("is_active").Default(true),
	}
}

func (Currency) Edges() []ent.Edge {
	return nil
}

func (Currency) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
	}
}
