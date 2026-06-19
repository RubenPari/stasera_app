CREATE TABLE IF NOT EXISTS staples (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    user_id     CHAR(36) NOT NULL,
    name        VARCHAR(200) NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (user_id, name),
    CONSTRAINT fk_staples_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_preferences (
    user_id              CHAR(36) NOT NULL PRIMARY KEY,
    disliked_ingredients JSON NOT NULL,
    max_prep_minutes    SMALLINT NOT NULL DEFAULT 30,
    preferred_cuisines  JSON NOT NULL,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_preferences_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);