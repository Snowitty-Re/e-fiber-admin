package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Page struct {
	ent.Schema
}

func (Page) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Page) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").Unique(),
		field.Enum("status").Values("draft", "published", "archived").Default("draft"),
		field.String("template").Default("default"),
		field.Time("published_at").Optional().Nillable(),
	}
}

func (Page) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("translations", PageTranslation.Type),
	}
}

func (Page) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}

type PageTranslation struct {
	ent.Schema
}

func (PageTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (PageTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int("page_id"),
		field.String("locale").MaxLen(8),
		field.String("title").NotEmpty(),
		field.JSON("content", map[string]any{}).Optional(),
		field.String("seo_title").Optional().Default(""),
		field.Text("seo_desc").Optional().Default(""),
	}
}

func (PageTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("page", Page.Type).Ref("translations").Field("page_id").Unique().Required(),
	}
}

func (PageTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("page_id", "locale").Unique(),
	}
}
