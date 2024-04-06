
-- +migrate Up
ALTER TABLE tasks ADD COLUMN user_id VARCHAR(255) NULL; -- NULLは許容する

-- +migrate Down
ALTER TABLE tasks DROP COLUMN user_id;
