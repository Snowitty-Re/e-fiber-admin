package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ShippingProfile struct {
	ent.Schema
}

func (ShippingProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (ShippingProfile) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("product_type").Default("physical"),
	}
}

func (ShippingProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("options", ShippingOption.Type),
	}
}

func (ShippingProfile) Indexes() []ent.Index {
	return nil
}

type ShippingOption struct {
	ent.Schema
}

func (ShippingOption) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (ShippingOption) Fields() []ent.Field {
	return []ent.Field{
		field.Int("profile_id"),
		field.String("name").NotEmpty(),
		field.Int64("price_amount").Default(0),
		field.String("price_currency").Default("USD").MaxLen(3),
		field.Int("estimated_days").Default(7),
		field.JSON("requirements", map[string]any{}).Optional(),
		field.Bool("is_active").Default(true),
	}
}

func (ShippingOption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("profile", ShippingProfile.Type).Ref("options").Field("profile_id").Unique().Required(),
	}
}

func (ShippingOption) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("profile_id"),
		index.Fields("is_active"),
	}
}

type ProductShippingProfile struct {
	ent.Schema
}

func (ProductShippingProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (ProductShippingProfile) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.Int("profile_id"),
	}
}

func (ProductShippingProfile) Edges() []ent.Edge {
	return nil
}

func (ProductShippingProfile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
		index.Fields("profile_id"),
	}
}