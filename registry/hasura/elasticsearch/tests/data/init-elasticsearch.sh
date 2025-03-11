#!/bin/bash
echo "Waiting for Elasticsearch to be ready..."

until curl -s -u elastic:elastic http://elasticsearch:9200/_cluster/health | grep '"status":"green"\|"status":"yellow"'; do
    sleep 5
done

echo "Elasticsearch is ready. Adding mappings and data..."

# Add customers data
echo "Adding mapping::customers"
sh /usr/share/elasticsearch/config/customers.sh

echo "Mappings and data loaded successfully!"
