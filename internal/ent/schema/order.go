package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Order struct {
	ent.Schema
}

func (Order) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.String("number").Unique(),
		field.Int("customer_id").Optional(),
		field.String("email").NotEmpty(),
		field.String("currency_code").Default("USD").MaxLen(3),
		field.String("locale").Default("en").MaxLen(8),
		field.Enum("status").Values("pending", "paid", "fulfilled", "cancelled", "refunded").Default("pending"),
		field.Enum("fulfillment_status").Values("not_fulfilled", "partial", "fulfilled").Default("not_fulfilled"),
		field.Enum("payment_status").Values("awaiting", "paid", "partial", "refunded").Default("awaiting"),
		field.JSON("shipping_address", map[string]any{}).Optional(),
		field.JSON("billing_address", map[string]any{}).Optional(),
		field.JSON("totals", map[string]any{}).Optional(),
		field.Int("shipping_option_id").Optional(),
		field.Time("placed_at").Optional().Nillable(),
		field.Time("cancelled_at").Optional().Nillable(),
	}
}

func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("items", OrderItem.Type),
		edge.To("fulfillments", Fulfillment.Type),
		edge.To("returns", OrderReturn.Type),
	}
}

func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("customer_id"),
		index.Fields("status"),
		index.Fields("email"),
	}
}

type OrderItem struct {
	ent.Schema
}

func (OrderItem) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (OrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int("order_id"),
		field.Int("variant_id").Optional(),
		field.String("sku").Default(""),
		field.String("title").NotEmpty(),
		field.Int("quantity").Default(1),
		field.Int64("unit_amount").Default(0),
		field.Int64("total_amount").Default(0),
		field.JSON("metadata", map[string]any{}).Optional(),
	}
}

func (OrderItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).Ref("items").Field("order_id").Unique().Required(),
	}
}

func (OrderItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
	}
}