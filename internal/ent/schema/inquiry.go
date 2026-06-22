package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Inquiry struct {
	ent.Schema
}

func (Inquiry) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Inquiry) Fields() []ent.Field {
	return []ent.Field{
		field.Int("form_id"),
		field.Int("customer_id").Optional(),
		field.String("email").NotEmpty(),
		field.String("phone").Optional().Default(""),
		field.String("name").Default(""),
		field.String("company").Optional().Default(""),
		field.JSON("payload", map[string]any{}).Optional(),
		field.Int("product_id").Optional(),
		field.Enum("status").Values("new", "contacted", "qualified", "converted", "closed").Default("new"),
		field.Int("assigned_admin_id").Optional(),
		field.Int("converted_order_id").Optional(),
	}
}

func (Inquiry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("form", FormDefinition.Type).Ref("inquiries").Field("form_id").Unique().Required(),
	}
}

func (Inquiry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("form_id"),
		index.Fields("customer_id"),
	}
}