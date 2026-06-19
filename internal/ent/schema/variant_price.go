package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type VariantPrice struct {
	ent.Schema
}

func (VariantPrice) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (VariantPrice) Fields() []ent.Field {
	return []ent.Field{
		field.Int("variant_id"),
		field.String("currency_code").MaxLen(3),
		field.Int64("amount"),
		field.Int64("compare_at_amount").Optional(),
	}
}

func (VariantPrice) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("variant", Variant.Type).Ref("prices").Field("variant_id").Unique().Required(),
	}
}

func (VariantPrice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("variant_id", "currency_code").Unique(),
	}
}
