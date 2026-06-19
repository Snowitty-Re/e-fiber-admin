package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type TaxRate struct {
	ent.Schema
}

func (TaxRate) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (TaxRate) Fields() []ent.Field {
	return []ent.Field{
		field.Int("region_id"),
		field.String("country_code").Optional().MaxLen(2),
		field.Float("rate").SchemaType(map[string]string{
			"postgres": "numeric(5,4)",
		}),
		field.String("name").NotEmpty(),
		field.Int("priority").Default(0),
	}
}

func (TaxRate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("region", Region.Type).Ref("tax_rates").Field("region_id").Unique().Required(),
	}
}

func (TaxRate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("region_id"),
		index.Fields("country_code"),
	}
}
