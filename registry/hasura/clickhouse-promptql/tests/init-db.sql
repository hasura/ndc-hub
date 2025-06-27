-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS testdb;

-- Use the database
USE testdb;

-- Create a sample table
CREATE TABLE IF NOT EXISTS users (
    id UInt32,
    name String,
    email String,
    age UInt8,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY id;

-- Insert sample data
INSERT INTO users (id, name, email, age) VALUES
    (1, 'John Doe', 'john@example.com', 25),
    (2, 'Jane Smith', 'jane@example.com', 30),
    (3, 'Bob Johnson', 'bob@example.com', 35),
    (4, 'Alice Brown', 'alice@example.com', 28),
    (5, 'Charlie Wilson', 'charlie@example.com', 32);

-- Create another sample table for analytics
CREATE TABLE IF NOT EXISTS page_views (
    user_id UInt32,
    page String,
    timestamp DateTime DEFAULT now(),
    duration UInt32
) ENGINE = MergeTree()
ORDER BY (user_id, timestamp);

-- Insert sample page view data
INSERT INTO page_views (user_id, page, duration) VALUES
    (1, '/home', 45),
    (1, '/products', 120),
    (2, '/home', 30),
    (2, '/about', 60),
    (3, '/products', 90),
    (3, '/contact', 25),
    (4, '/home', 55),
    (5, '/products', 180);
