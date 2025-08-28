-- Phoenix SQL initialization script
-- Create sample tables for testing

-- Create a simple users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(255),
    age INTEGER,
    created_at TIMESTAMP,
    is_active BOOLEAN
);

-- Create an orders table
CREATE TABLE IF NOT EXISTS orders (
    order_id BIGINT NOT NULL PRIMARY KEY,
    user_id BIGINT,
    product_name VARCHAR(255),
    quantity INTEGER,
    price DECIMAL(10,2),
    order_date TIMESTAMP,
    status VARCHAR(50)
);

-- Create a products table
CREATE TABLE IF NOT EXISTS products (
    product_id BIGINT NOT NULL PRIMARY KEY,
    name VARCHAR(255),
    description VARCHAR(1000),
    category VARCHAR(100),
    price DECIMAL(10,2),
    stock_quantity INTEGER,
    created_at TIMESTAMP
);

-- Create an index on users email
CREATE INDEX IF NOT EXISTS users_email_idx ON users (email);

-- Create an index on orders user_id
CREATE INDEX IF NOT EXISTS orders_user_id_idx ON orders (user_id);

-- Create an index on products category
CREATE INDEX IF NOT EXISTS products_category_idx ON products (category);
