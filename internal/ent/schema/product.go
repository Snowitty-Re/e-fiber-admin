package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Product struct {
	ent.Schema
}

func (Product) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").Unique(),
		field.Enum("product_type").Values("simple", "variable", "virtual", "grouped").Default("simple"),
		field.Enum("status").Values("draft", "published", "archived").Default("draft"),
		field.Int("category_id").Optional(),
		field.Int("weight_g").Optional(),
		field.Bool("is_virtual").Default(false),
		field.Bool("is_downloadable").Default(false),
		field.Time("published_at").Optional().Nillable(),
	}
}

func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("translations", ProductTranslation.Type),
		edge.To("variants", Variant.Type),
		edge.To("options", ProductOption.Type),
	}
}

func (Product) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("product_type"),
		index.Fields("category_id"),
	}
}
