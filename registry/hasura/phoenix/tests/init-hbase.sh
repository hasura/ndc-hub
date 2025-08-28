#!/bin/bash

# HBase initialization script
# This script sets up HBase for Phoenix integration

echo "Starting HBase initialization..."

# Wait for HBase to be fully started
sleep 30

# Create Phoenix system tables if they don't exist
echo "Creating Phoenix system tables..."
/hbase/bin/hbase shell <<EOF
# Enable Phoenix system tables
create_namespace 'SYSTEM' if not exists
EOF

echo "HBase initialization complete!"
