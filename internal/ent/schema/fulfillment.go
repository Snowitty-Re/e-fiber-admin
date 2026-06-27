package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Fulfillment struct {
	ent.Schema
}

func (Fulfillment) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Fulfillment) Fields() []ent.Field {
	return []ent.Field{
		field.Int("order_id"),
		field.Int("shipping_option_id").Optional(),
		field.String("tracking_number").Optional().Default(""),
		field.Enum("status").Values("pending", "fulfilled", "canceled").Default("pending"),
		field.JSON("metadata", map[string]any{}).Optional(),
	}
}

func (Fulfillment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).Ref("fulfillments").Field("order_id").Unique().Required(),
		edge.To("items", FulfillmentItem.Type),
	}
}

func (Fulfillment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
	}
}

type FulfillmentItem struct {
	ent.Schema
}

func (FulfillmentItem) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (FulfillmentItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int("fulfillment_id"),
		field.Int("order_item_id"),
		field.Int("quantity").Default(1),
	}
}

func (FulfillmentItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("fulfillment", Fulfillment.Type).Ref("items").Field("fulfillment_id").Unique().Required(),
	}
}

func (FulfillmentItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("fulfillment_id"),
		index.Fields("order_item_id"),
	}
}

type OrderReturn struct {
	ent.Schema
}

func (OrderReturn) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (OrderReturn) Fields() []ent.Field {
	return []ent.Field{
		field.Int("order_id"),
		field.Enum("status").Values("pending", "approved", "rejected", "completed").Default("pending"),
		field.String("reason").Optional().Default(""),
		field.Int64("refund_amount").Default(0),
		field.String("currency_code").Default("USD").MaxLen(3),
		field.JSON("totals", map[string]any{}).Optional(),
	}
}

func (OrderReturn) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).Ref("returns").Field("order_id").Unique().Required(),
		edge.To("items", ReturnItem.Type),
	}
}

func (OrderReturn) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("status"),
	}
}

type ReturnItem struct {
	ent.Schema
}

func (ReturnItem) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (ReturnItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int("return_id"),
		field.Int("order_item_id"),
		field.Int("quantity").Default(1),
		field.String("reason").Optional().Default(""),
	}
}

func (ReturnItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("return", OrderReturn.Type).Ref("items").Field("return_id").Unique().Required(),
	}
}

func (ReturnItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("return_id"),
	}
}