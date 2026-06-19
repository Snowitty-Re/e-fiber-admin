-- Add tsvector GIN index for product full-text search
-- Per docs/03 §11 and docs/08 M1.10
CREATE INDEX IF NOT EXISTS product_translations_tsvector_idx
ON product_translations
USING gin (to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(description, '')));
