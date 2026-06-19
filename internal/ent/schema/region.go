package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Region struct {
	ent.Schema
}

func (Region) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Region) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("locale").MaxLen(8),
		field.String("currency_code").MaxLen(3),
		field.Bool("tax_inclusive").Default(false),
		field.JSON("countries", []string{}).Optional(),
		field.Enum("status").Values("active", "inactive").Default("active"),
	}
}

func (Region) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("tax_rates", TaxRate.Type),
	}
}

func (Region) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}
