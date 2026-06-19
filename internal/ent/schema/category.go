package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Category struct {
	ent.Schema
}

func (Category) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").Unique(),
		field.Int("parent_id").Optional(),
		field.Int("position").Default(0),
	}
}

func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("translations", CategoryTranslation.Type),
	}
}

func (Category) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id"),
	}
}

type CategoryTranslation struct {
	ent.Schema
}

func (CategoryTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (CategoryTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int("category_id"),
		field.String("locale").MaxLen(8),
		field.String("name").NotEmpty(),
		field.Text("description").Optional().Default(""),
	}
}

func (CategoryTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("category", Category.Type).Ref("translations").Field("category_id").Unique().Required(),
	}
}

func (CategoryTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("category_id", "locale").Unique(),
	}
}
