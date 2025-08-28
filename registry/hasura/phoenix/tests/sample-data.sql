BEGIN;

-- Insert sample users
UPSERT INTO users (id, name, email, age, created_at, is_active) VALUES
(1, 'John Doe', 'john.doe@example.com', 30, CURRENT_TIME(), true),
(2, 'Jane Smith', 'jane.smith@example.com', 25, CURRENT_TIME(), true),
(3, 'Bob Johnson', 'bob.johnson@example.com', 35, CURRENT_TIME(), false),
(4, 'Alice Brown', 'alice.brown@example.com', 28, CURRENT_TIME(), true),
(5, 'Charlie Wilson', 'charlie.wilson@example.com', 42, CURRENT_TIME(), true);

-- Commit the changes
COMMIT;
