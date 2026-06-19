package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Collection struct {
	ent.Schema
}

func (Collection) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Collection) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").Unique(),
		field.Enum("status").Values("draft", "published", "archived").Default("draft"),
	}
}

func (Collection) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("translations", CollectionTranslation.Type),
		edge.To("products", Product.Type),
	}
}

func (Collection) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}

type CollectionTranslation struct {
	ent.Schema
}

func (CollectionTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (CollectionTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int("collection_id"),
		field.String("locale").MaxLen(8),
		field.String("name").NotEmpty(),
		field.Text("description").Optional().Default(""),
	}
}

func (CollectionTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("collection", Collection.Type).Ref("translations").Field("collection_id").Unique().Required(),
	}
}

func (CollectionTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collection_id", "locale").Unique(),
	}
}
