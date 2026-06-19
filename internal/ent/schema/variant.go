package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Variant struct {
	ent.Schema
}

func (Variant) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Variant) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.String("sku").Unique(),
		field.String("barcode").Optional().Default(""),
		field.Int("weight_g").Optional(),
		field.Bool("allow_backorder").Default(false),
		field.Int("inventory").Default(0),
		field.Int("position").Default(0),
	}
}

func (Variant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).Ref("variants").Field("product_id").Unique().Required(),
		edge.To("prices", VariantPrice.Type),
		edge.To("option_values", VariantOptionValue.Type),
	}
}

func (Variant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
		index.Fields("sku"),
	}
}
