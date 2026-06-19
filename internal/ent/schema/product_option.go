package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ProductOption struct {
	ent.Schema
}

func (ProductOption) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (ProductOption) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.String("name").NotEmpty(),
		field.Int("position").Default(0),
	}
}

func (ProductOption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).Ref("options").Field("product_id").Unique().Required(),
		edge.To("values", ProductOptionValue.Type),
	}
}

func (ProductOption) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
	}
}

type ProductOptionValue struct {
	ent.Schema
}

func (ProductOptionValue) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (ProductOptionValue) Fields() []ent.Field {
	return []ent.Field{
		field.Int("option_id"),
		field.String("value").NotEmpty(),
		field.Int("position").Default(0),
	}
}

func (ProductOptionValue) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("option", ProductOption.Type).Ref("values").Field("option_id").Unique().Required(),
	}
}

func (ProductOptionValue) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("option_id"),
	}
}

type VariantOptionValue struct {
	ent.Schema
}

func (VariantOptionValue) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (VariantOptionValue) Fields() []ent.Field {
	return []ent.Field{
		field.Int("variant_id"),
		field.Int("option_value_id"),
	}
}

func (VariantOptionValue) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("variant", Variant.Type).Ref("option_values").Field("variant_id").Unique().Required(),
	}
}

func (VariantOptionValue) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("variant_id", "option_value_id").Unique(),
	}
}
