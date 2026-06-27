-- 006_add_indexes.sql
-- Adds secondary indexes on foreign-key columns used in WHERE/JOIN clauses.
-- These columns are declared as FK targets but MySQL does not automatically
-- create an index on the referencing side.

CREATE INDEX idx_recipes_user_id ON recipes(user_id);
CREATE INDEX idx_meal_plan_days_recipe_id ON meal_plan_days(recipe_id);
CREATE INDEX idx_shopping_lists_user_id ON shopping_lists(user_id);
CREATE INDEX idx_shopping_items_list_id ON shopping_items(list_id);