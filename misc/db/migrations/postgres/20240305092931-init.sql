
-- +migrate Up
CREATE TABLE tasks (
  id SERIAL PRIMARY KEY,
  task_id VARCHAR(255) NOT NULL,
  task_status VARCHAR(255) NOT NULL,
  title TEXT NOT NULL,
  page_url TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
drop table tasks;

