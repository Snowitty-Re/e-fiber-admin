package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Cart struct {
	ent.Schema
}

func (Cart) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Cart) Fields() []ent.Field {
	return []ent.Field{
		field.Int("customer_id").Optional(),
		field.String("email").Optional().Default(""),
		field.String("currency_code").Default("USD").MaxLen(3),
		field.String("locale").Default("en").MaxLen(8),
		field.Enum("status").Values("active", "converted", "abandoned").Default("active"),
		field.Time("expires_at").Optional().Nillable(),
	}
}

func (Cart) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("items", CartItem.Type),
	}
}

func (Cart) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("customer_id"),
		index.Fields("status"),
	}
}

type CartItem struct {
	ent.Schema
}

func (CartItem) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (CartItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int("cart_id"),
		field.Int("variant_id"),
		field.Int("product_id"),
		field.Int("quantity").Default(1),
		field.Int64("unit_amount").Default(0),
		field.String("currency_code").Default("USD").MaxLen(3),
		field.String("sku").Default(""),
		field.JSON("metadata", map[string]any{}).Optional(),
	}
}

func (CartItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cart", Cart.Type).Ref("items").Field("cart_id").Unique().Required(),
	}
}

func (CartItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cart_id"),
		index.Fields("variant_id"),
	}
}