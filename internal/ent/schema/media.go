package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Media struct {
	ent.Schema
}

func (Media) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Media) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").NotEmpty(),
		field.String("url").NotEmpty(),
		field.String("mime_type").NotEmpty(),
		field.Int64("size_bytes"),
		field.Int("width").Optional(),
		field.Int("height").Optional(),
		field.Enum("kind").Values("image", "document", "video").Default("image"),
	}
}

func (Media) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("translations", MediaTranslation.Type),
	}
}

func (Media) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kind"),
	}
}

type MediaTranslation struct {
	ent.Schema
}

func (MediaTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (MediaTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int("media_id"),
		field.String("locale").MaxLen(8),
		field.String("alt").Optional().Default(""),
	}
}

func (MediaTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).Ref("translations").Field("media_id").Unique().Required(),
	}
}

func (MediaTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("media_id", "locale").Unique(),
	}
}
