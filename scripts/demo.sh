#!/bin/bash

# TechCart Failover API Demo
# This script demonstrates the failover scenario

set -e

echo "================================================"
echo "   TechCart Failover Intelligence API Demo"
echo "================================================"
echo ""

# Check if server is running
if ! curl -s http://localhost:8080/api/v1/processors > /dev/null 2>&1; then
    echo "❌ Server not running. Starting server..."
    echo "   Run: go run cmd/server/main.go"
    echo ""
    echo "Then run this script again."
    exit 1
fi

echo "✅ Server is running"
echo ""

# Run the data generator
echo "🚀 Running data generator with outage simulation..."
echo ""
go run scripts/generate_data.go

echo ""
echo "================================================"
echo "   Demo Complete!"
echo "================================================"
echo ""
echo "Try these commands to explore:"
echo ""
echo "  # Get all processor health"
echo "  curl -s localhost:8080/api/v1/health | jq"
echo ""
echo "  # Get routing recommendation for PIX in Brazil"
echo "  curl -s 'localhost:8080/api/v1/routing/recommend?payment_method=PIX&country=BR' | jq"
echo ""
echo "  # Get alerts (status transitions)"
echo "  curl -s localhost:8080/api/v1/alerts | jq"
echo ""
echo "  # Send a transaction"
echo '  curl -X POST localhost:8080/api/v1/transactions -H "Content-Type: application/json" -d '"'"'{"processor_id":"processor_a","result":"approved","payment_method":"PIX","country":"BR","amount":100}'"'"' | jq'
echo ""
