CREATE TABLE IF NOT EXISTS shopping_lists (
    id           CHAR(36) NOT NULL PRIMARY KEY,
    user_id      CHAR(36) NOT NULL,
    plan_id      CHAR(36),
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    CONSTRAINT fk_shopping_lists_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_shopping_lists_plan FOREIGN KEY (plan_id) REFERENCES meal_plans(id)
);

CREATE TABLE IF NOT EXISTS shopping_items (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    list_id     CHAR(36) NOT NULL,
    name        VARCHAR(200) NOT NULL,
    quantity    VARCHAR(50) NOT NULL,
    aisle       VARCHAR(50) NOT NULL,
    is_checked  BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order  SMALLINT NOT NULL DEFAULT 0,
    CONSTRAINT fk_shopping_items_list FOREIGN KEY (list_id) REFERENCES shopping_lists(id) ON DELETE CASCADE
);