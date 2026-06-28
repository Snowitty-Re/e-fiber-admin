package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Discount struct {
	ent.Schema
}

func (Discount) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Discount) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").Optional(),
		field.String("name").NotEmpty(),
		field.Bool("is_dynamic").Default(false),
		field.Time("starts_at").Optional().Nillable(),
		field.Time("ends_at").Optional().Nillable(),
		field.Int("usage_limit").Optional(),
		field.Int("usage_count").Default(0),
		field.Enum("status").Values("active", "expired", "disabled").Default("active"),
	}
}

func (Discount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("rules", DiscountRule.Type),
		edge.To("conditions", DiscountCondition.Type),
	}
}

func (Discount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code"),
		index.Fields("status"),
	}
}

type DiscountRule struct {
	ent.Schema
}

func (DiscountRule) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (DiscountRule) Fields() []ent.Field {
		return []ent.Field{
		field.Int("discount_id"),
		field.Enum("type").Values("percentage", "fixed", "shipping", "free_item").Default("percentage"),
		field.Int64("value").Default(0),
		field.Enum("allocation").Values("all", "items").Default("all"),
		field.JSON("target", map[string]any{}).Optional(),
	}
}

func (DiscountRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("discount", Discount.Type).Ref("rules").Field("discount_id").Unique().Required(),
	}
}

func (DiscountRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("discount_id"),
	}
}

type DiscountCondition struct {
	ent.Schema
}

func (DiscountCondition) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (DiscountCondition) Fields() []ent.Field {
	return []ent.Field{
		field.Int("discount_id"),
		field.Enum("type").Values("products", "collections", "categories", "customer_groups"),
		field.JSON("values", []int{}).Optional(),
	}
}

func (DiscountCondition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("discount", Discount.Type).Ref("conditions").Field("discount_id").Unique().Required(),
	}
}

func (DiscountCondition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("discount_id"),
	}
}