package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ProductMedia struct {
	ent.Schema
}

func (ProductMedia) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (ProductMedia) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.Int("media_id"),
		field.Int("position").Default(0),
		field.Enum("role").Values("image", "gallery", "document").Default("gallery"),
	}
}

func (ProductMedia) Edges() []ent.Edge {
	return nil
}

func (ProductMedia) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
		index.Fields("media_id"),
	}
}
