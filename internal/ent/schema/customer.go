package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Customer struct {
	ent.Schema
}

func (Customer) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Customer) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").Unique().NotEmpty(),
		field.String("phone").Optional().Default(""),
		field.String("first_name").Optional().Default(""),
		field.String("last_name").Optional().Default(""),
		field.String("password_hash").Optional().Sensitive().Default(""),
		field.Enum("status").Values("active", "disabled", "guest").Default("active"),
		field.String("default_currency").Default("USD").MaxLen(3),
		field.String("default_locale").Default("en").MaxLen(8),
	}
}

func (Customer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("addresses", CustomerAddress.Type),
		edge.To("groups", CustomerGroup.Type),
	}
}

func (Customer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}

type CustomerAddress struct {
	ent.Schema
}

func (CustomerAddress) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (CustomerAddress) Fields() []ent.Field {
	return []ent.Field{
		field.Int("customer_id"),
		field.String("first_name").Default(""),
		field.String("last_name").Default(""),
		field.String("company").Optional().Default(""),
		field.String("address1").Default(""),
		field.String("address2").Optional().Default(""),
		field.String("city").Default(""),
		field.String("province").Optional().Default(""),
		field.String("postal_code").Optional().Default(""),
		field.String("country_code").Default(""),
		field.String("phone").Optional().Default(""),
		field.Bool("is_default_shipping").Default(false),
		field.Bool("is_default_billing").Default(false),
	}
}

func (CustomerAddress) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("customer", Customer.Type).Ref("addresses").Field("customer_id").Unique().Required(),
	}
}

func (CustomerAddress) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("customer_id"),
	}
}

type CustomerGroup struct {
	ent.Schema
}

func (CustomerGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (CustomerGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").Unique(),
		field.String("name").NotEmpty(),
	}
}

func (CustomerGroup) Edges() []ent.Edge {
	return nil
}

func (CustomerGroup) Indexes() []ent.Index {
	return nil
}