CREATE TABLE IF NOT EXISTS meal_plans (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    user_id     CHAR(36) NOT NULL,
    week_start  DATE NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, week_start),
    CONSTRAINT fk_meal_plans_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS meal_plan_days (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    plan_id     CHAR(36) NOT NULL,
    day_of_week SMALLINT NOT NULL,
    recipe_id   CHAR(36) NOT NULL,
    UNIQUE (plan_id, day_of_week),
    CONSTRAINT fk_meal_plan_days_plan FOREIGN KEY (plan_id) REFERENCES meal_plans(id) ON DELETE CASCADE,
    CONSTRAINT fk_meal_plan_days_recipe FOREIGN KEY (recipe_id) REFERENCES recipes(id)
);