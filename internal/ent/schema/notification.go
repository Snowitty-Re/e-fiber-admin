package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type EmailTemplate struct {
	ent.Schema
}

func (EmailTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (EmailTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").Unique(),
		field.JSON("variables", []string{}).Optional(),
	}
}

func (EmailTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("translations", EmailTemplateTranslation.Type),
	}
}

func (EmailTemplate) Indexes() []ent.Index {
	return nil
}

type EmailTemplateTranslation struct {
	ent.Schema
}

func (EmailTemplateTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (EmailTemplateTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int("email_template_id"),
		field.String("locale").MaxLen(8),
		field.String("subject").NotEmpty(),
		field.Text("body_html").NotEmpty(),
	}
}

func (EmailTemplateTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("email_template", EmailTemplate.Type).Ref("translations").Field("email_template_id").Unique().Required(),
	}
}

func (EmailTemplateTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email_template_id", "locale").Unique(),
	}
}

type Notification struct {
	ent.Schema
}

func (Notification) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("channel").Values("email", "webhook", "inapp").Default("email"),
		field.String("recipient").NotEmpty(),
		field.String("template_code").Default(""),
		field.JSON("payload", map[string]any{}).Optional(),
		field.Enum("status").Values("pending", "sent", "failed").Default("pending"),
		field.Int("attempts").Default(0),
		field.Time("sent_at").Optional().Nillable(),
	}
}

func (Notification) Edges() []ent.Edge {
	return nil
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("template_code"),
	}
}