package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Menu struct {
	ent.Schema
}

func (Menu) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Menu) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").Unique(),
		field.String("location").Default("header"),
	}
}

func (Menu) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("items", MenuItem.Type),
	}
}

func (Menu) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("location"),
	}
}

type MenuItem struct {
	ent.Schema
}

func (MenuItem) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (MenuItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int("menu_id"),
		field.Int("parent_id").Optional(),
		field.Enum("target_type").Values("page", "category", "url", "post").Default("url"),
		field.Int("target_id").Optional(),
		field.String("url").Optional().Default(""),
		field.Int("position").Default(0),
	}

}

func (MenuItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("menu", Menu.Type).Ref("items").Field("menu_id").Unique().Required(),
		edge.To("translations", MenuItemTranslation.Type),
	}
}

func (MenuItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("menu_id"),
		index.Fields("parent_id"),
	}
}

type MenuItemTranslation struct {
	ent.Schema
}

func (MenuItemTranslation) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (MenuItemTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int("menu_item_id"),
		field.String("locale").MaxLen(8),
		field.String("title").NotEmpty(),
	}
}

func (MenuItemTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("menu_item", MenuItem.Type).Ref("translations").Field("menu_item_id").Unique().Required(),
	}
}

func (MenuItemTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("menu_item_id", "locale").Unique(),
	}
}
