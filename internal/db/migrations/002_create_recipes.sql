CREATE TABLE IF NOT EXISTS recipes (
    id              CHAR(36) NOT NULL PRIMARY KEY,
    user_id         CHAR(36) NOT NULL,
    name            VARCHAR(200) NOT NULL,
    prep_minutes    SMALLINT NOT NULL,
    servings        SMALLINT NOT NULL DEFAULT 1,
    ingredients     JSON NOT NULL,
    steps           JSON NOT NULL,
    is_rescue       BOOLEAN NOT NULL DEFAULT FALSE,
    times_cooked    SMALLINT NOT NULL DEFAULT 0,
    last_cooked_at  DATE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_recipes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);