-- Initialization script for test database
-- This script creates sample tables and data for integration tests

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    stock INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    total NUMERIC(10, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample users
INSERT INTO users (username, email, active) VALUES
    ('john_doe', 'john@example.com', true),
    ('jane_smith', 'jane@example.com', true),
    ('bob_wilson', 'bob@example.com', false),
    ('alice_johnson', 'alice@example.com', true),
    ('charlie_brown', 'charlie@example.com', true);

-- Insert sample products
INSERT INTO products (name, description, price, stock) VALUES
    ('Laptop', 'High-performance laptop', 999.99, 15),
    ('Mouse', 'Wireless mouse', 29.99, 50),
    ('Keyboard', 'Mechanical keyboard', 79.99, 30),
    ('Monitor', '27-inch 4K monitor', 399.99, 10),
    ('Headphones', 'Noise-cancelling headphones', 199.99, 25);

-- Insert sample orders
INSERT INTO orders (user_id, total, status) VALUES
    (1, 1029.98, 'completed'),
    (2, 479.98, 'pending'),
    (1, 199.99, 'completed'),
    (4, 999.99, 'processing'),
    (5, 109.98, 'completed');

-- Create a view for testing
CREATE OR REPLACE VIEW user_order_summary AS
SELECT 
    u.id,
    u.username,
    u.email,
    COUNT(o.id) as order_count,
    COALESCE(SUM(o.total), 0) as total_spent
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.username, u.email;

-- Create a test table with various data types
CREATE TABLE IF NOT EXISTS data_types_test (
    id SERIAL PRIMARY KEY,
    text_col TEXT,
    varchar_col VARCHAR(50),
    int_col INTEGER,
    bigint_col BIGINT,
    numeric_col NUMERIC(10, 2),
    float_col REAL,
    double_col DOUBLE PRECISION,
    bool_col BOOLEAN,
    date_col DATE,
    timestamp_col TIMESTAMP,
    json_col JSON,
    array_col INTEGER[]
);

-- Insert sample data with various types
INSERT INTO data_types_test (
    text_col, varchar_col, int_col, bigint_col, numeric_col,
    float_col, double_col, bool_col, date_col, timestamp_col,
    json_col, array_col
) VALUES
    (
        'Sample text',
        'Short text',
        42,
        9223372036854775807,
        12345.67,
        3.14,
        2.718281828,
        true,
        '2024-01-15',
        '2024-01-15 10:30:00',
        '{"key": "value", "number": 123}',
        ARRAY[1, 2, 3, 4, 5]
    ),
    (
        'Another sample',
        'Text here',
        -100,
        1234567890,
        999.99,
        -1.5,
        0.0,
        false,
        '2023-12-31',
        '2023-12-31 23:59:59',
        '{"active": true, "count": 0}',
        ARRAY[10, 20, 30]
    );

-- Create indexes for better query performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(active);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_products_name ON products(name);

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO testuser;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO testuser;

-- Log completion
DO $$
BEGIN
    RAISE NOTICE 'Test database initialized successfully';
END $$;