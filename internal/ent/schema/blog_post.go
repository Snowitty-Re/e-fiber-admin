package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type BlogPost struct {
	ent.Schema
}

func (BlogPost) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (BlogPost) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").Unique(),
		field.Int("author_admin_id"),
		field.Enum("status").Values("draft", "published", "archived").Default("draft"),
		field.Time("published_at").Optional().Nillable(),
	}
}

func (BlogPost) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("translations", BlogPostTranslation.Type),
	}
}

func (BlogPost) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("author_admin_id"),
	}
}

type BlogPostTranslation struct {
	ent.Schema
}

func (BlogPostTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (BlogPostTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int("blog_post_id"),
		field.String("locale").MaxLen(8),
		field.String("title").NotEmpty(),
		field.Text("excerpt").Optional().Default(""),
		field.Text("content").Optional().Default(""),
		field.String("seo_title").Optional().Default(""),
		field.Text("seo_desc").Optional().Default(""),
	}
}

func (BlogPostTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("blog_post", BlogPost.Type).Ref("translations").Field("blog_post_id").Unique().Required(),
	}
}

func (BlogPostTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("blog_post_id", "locale").Unique(),
	}
}
