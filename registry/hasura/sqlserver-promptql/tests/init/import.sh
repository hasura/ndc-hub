#!/bin/bash

# Start SQL Server in background
echo "Waiting for SQL Server to start..."
/opt/mssql/bin/sqlservr &
sleep 5

# Wait for SQL Server to be ready to accept connections
echo "Waiting for SQL Server to accept connections..."
until /opt/mssql-tools18/bin/sqlcmd -S localhost,1433 -U sa -P 'Password123' -C -Q "SELECT 1" > /dev/null 2>&1
do
  sleep 2
done

# Run the SQL script
echo "Importing Chinook database..."
/opt/mssql-tools18/bin/sqlcmd -S localhost,1433 -U sa -P 'Password123' -C -i /usr/config/chinook.sql

# Keep the container running
wait