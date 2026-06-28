package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PaymentProvider struct {
	ent.Schema
}

func (PaymentProvider) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (PaymentProvider) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").Unique(),
		field.String("name").NotEmpty(),
		field.Bool("is_active").Default(false),
		field.JSON("config", map[string]any{}).Optional(),
	}
}

func (PaymentProvider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sessions", PaymentSession.Type),
	}
}

func (PaymentProvider) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
	}
}

type PaymentSession struct {
	ent.Schema
}

func (PaymentSession) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (PaymentSession) Fields() []ent.Field {
	return []ent.Field{
		field.Int("order_id"),
		field.String("provider_code").NotEmpty(),
		field.Enum("status").Values("pending", "authorized", "captured", "canceled", "failed").Default("pending"),
		field.Int64("amount").Default(0),
		field.String("currency_code").Default("USD").MaxLen(3),
		field.JSON("provider_data", map[string]any{}).Optional(),
	}
}

func (PaymentSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("transactions", Transaction.Type),
	}
}

func (PaymentSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("status"),
	}
}

type Transaction struct {
	ent.Schema
}

func (Transaction) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		field.Int("payment_session_id"),
		field.Int64("amount").Default(0),
		field.String("currency_code").Default("USD").MaxLen(3),
		field.Enum("type").Values("authorize", "capture", "refund").Default("capture"),
		field.Enum("status").Values("pending", "succeeded", "failed").Default("pending"),
		field.String("reference").Optional().Default(""),
	}
}

func (Transaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", PaymentSession.Type).Ref("transactions").Field("payment_session_id").Unique().Required(),
	}
}

func (Transaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("payment_session_id"),
	}
}