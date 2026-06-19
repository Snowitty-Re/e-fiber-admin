package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ProductTranslation struct {
	ent.Schema
}

func (ProductTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (ProductTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.String("locale").MaxLen(8),
		field.String("title").NotEmpty(),
		field.String("subtitle").Optional().Default(""),
		field.Text("description").Optional().Default(""),
		field.Text("material").Optional().Default(""),
		field.Text("origin").Optional().Default(""),
		field.Text("packing").Optional().Default(""),
		field.String("seo_title").Optional().Default(""),
		field.Text("seo_desc").Optional().Default(""),
	}
}

func (ProductTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).Ref("translations").Field("product_id").Unique().Required(),
	}
}

func (ProductTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id", "locale").Unique(),
	}
}
