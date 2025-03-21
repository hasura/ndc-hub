#!/bin/bash

curl -X PUT "http://localhost:9200/customers/" -u elastic:elastic -H 'Content-Type: application/json' -d'
{
  "mappings": {
  "properties": {
    "customer_id": {
      "type": "keyword"
    },
    "name": {
      "type": "text",
      "fields": {
        "keyword": {
          "type": "keyword"
        }
      }
    },
    "email": {
      "type": "keyword",
      "index": true
    },
    "location": {
      "type": "geo_point"
    }
  }
  }
}
'

curl -X POST "http://localhost:9200/_bulk" -u elastic:elastic -H 'Content-Type: application/json' -d'
{ "index": { "_index": "customers", "_id": "1" } }
{ "customer_id": "CUST001", "name": "John Doe", "email": "john.doe@example.com", "location": { "lat": 40.7128, "lon": -74.0060 } }
{ "index": { "_index": "customers", "_id": "2" } }
{ "customer_id": "CUST002", "name": "Jane Smith", "email": "jane.smith@example.com", "location": { "lat": 34.0522, "lon": -118.2437 } }
{ "index": { "_index": "customers", "_id": "3" } }
{ "customer_id": "CUST003", "name": "Alice Johnson", "email": "alice.j@example.com", "location": { "lat": 51.5074, "lon": -0.1278 } }
{ "index": { "_index": "customers", "_id": "4" } }
{ "customer_id": "CUST004", "name": "Bob Brown", "email": "bob.brown@example.com", "location": { "lat": 48.8566, "lon": 2.3522 } }
{ "index": { "_index": "customers", "_id": "5" } }
{ "customer_id": "CUST005", "name": "Charlie Davis", "email": "charlie.d@example.com", "location": { "lat": 35.6895, "lon": 139.6917 } }
'