package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Store struct {
	ent.Schema
}

func (Store) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Store) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("slug").Unique(),
		field.Enum("site_type").Values("corporate", "store", "hybrid").Default("store"),
		field.String("default_locale").Default("en").MaxLen(8),
		field.String("default_currency").Default("USD").MaxLen(3),
		field.JSON("feature_flags", map[string]bool{}).Optional(),
		field.String("timezone").Default("UTC"),
		field.Enum("status").Values("active", "maintenance").Default("active"),
	}
}

func (Store) Edges() []ent.Edge {
	return nil
}
