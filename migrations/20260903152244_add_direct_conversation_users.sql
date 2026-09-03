-- +goose Up

ALTER TABLE conversations
    ADD COLUMN user_low CHAR(36) NULL,
    ADD COLUMN user_high CHAR(36) NULL;

ALTER TABLE conversations
    ADD CONSTRAINT fk_conversations_user_low
        FOREIGN KEY (user_low)
        REFERENCES users(id)
        ON DELETE CASCADE;

ALTER TABLE conversations
    ADD CONSTRAINT fk_conversations_user_high
        FOREIGN KEY (user_high)
        REFERENCES users(id)
        ON DELETE CASCADE;

ALTER TABLE conversations
    ADD CONSTRAINT uq_conversations_direct_users
        UNIQUE (user_low, user_high);


-- +goose Down

ALTER TABLE conversations
    DROP CONSTRAINT uq_conversations_direct_users;

ALTER TABLE conversations
    DROP FOREIGN KEY fk_conversations_user_low;

ALTER TABLE conversations
    DROP FOREIGN KEY fk_conversations_user_high;

ALTER TABLE conversations
    DROP COLUMN user_low,
    DROP COLUMN user_high;