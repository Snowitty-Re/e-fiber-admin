package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Tag struct {
	ent.Schema
}

func (Tag) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").Unique(),
	}
}

func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("translations", TagTranslation.Type),
		edge.To("products", Product.Type),
	}
}

func (Tag) Indexes() []ent.Index {
	return nil
}

type TagTranslation struct {
	ent.Schema
}

func (TagTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (TagTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int("tag_id"),
		field.String("locale").MaxLen(8),
		field.String("name").NotEmpty(),
	}
}

func (TagTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tag", Tag.Type).Ref("translations").Field("tag_id").Unique().Required(),
	}
}

func (TagTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tag_id", "locale").Unique(),
	}
}
