#!/bin/bash

echo "Initializing Phoenix database with sample data..."

# Wait a bit more for Phoenix to be fully ready
sleep 30

# Run the Phoenix SQL initialization
echo "Creating tables..."
/opt/phoenix-server/bin/sqlline.py localhost:2181 /opt/init-phoenix.sql

echo "Inserting sample data..."
/opt/phoenix-server/bin/sqlline.py localhost:2181 /opt/sample-data.sql

echo "Phoenix initialization complete!"

# Test the setup
echo "Testing Phoenix setup..."
/opt/phoenix-server/bin/sqlline.py localhost:2181 <<EOF
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM products;
SELECT COUNT(*) FROM orders;
!quit
EOF

echo "Phoenix database is ready!"
