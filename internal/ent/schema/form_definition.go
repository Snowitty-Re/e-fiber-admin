package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type FormDefinition struct {
	ent.Schema
}

func (FormDefinition) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (FormDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").Unique(),
		field.JSON("fields", []map[string]any{}).Optional(),
		field.JSON("notify_emails", []string{}).Optional(),
		field.Bool("is_active").Default(true),
	}
}

func (FormDefinition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("translations", FormDefinitionTranslation.Type),
		edge.To("inquiries", Inquiry.Type),
	}
}

func (FormDefinition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
	}
}

type FormDefinitionTranslation struct {
	ent.Schema
}

func (FormDefinitionTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (FormDefinitionTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int("form_definition_id"),
		field.String("locale").MaxLen(8),
		field.String("title").NotEmpty(),
		field.JSON("field_labels", map[string]string{}).Optional(),
	}
}

func (FormDefinitionTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("form_definition", FormDefinition.Type).Ref("translations").Field("form_definition_id").Unique().Required(),
	}
}

func (FormDefinitionTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("form_definition_id", "locale").Unique(),
	}
}