#!/bin/bash

# Start the blockchain event auto-sync service
# This script starts the auto-sync service that monitors the indexer database
# and syncs events to the main application database

echo "🚀 Starting Sukuk Blockchain Event Auto-Sync Service"
echo "======================================================"

# Check if the main database is available
echo "📡 Checking database connection..."
if ! nc -z localhost 5432; then
    echo "❌ Database not available on localhost:5432"
    echo "Please start PostgreSQL and ensure the database is running"
    exit 1
fi

# Check if the indexer database is available
echo "📡 Checking indexer database..."
if ! psql postgresql://postgres:postgres@localhost:5432/sukuk_poc_new -c "SELECT 1" > /dev/null 2>&1; then
    echo "❌ Indexer database 'sukuk_poc_new' not available"
    echo "Please ensure the Ponder indexer is running and the database exists"
    exit 1
fi

echo "✅ Database connections verified"

# Start the auto-sync service
echo "🔄 Starting auto-sync service..."
echo "📊 Sync interval: 30 seconds"
echo "🎯 Monitoring events: SukukPurchased, RedemptionRequested"
echo ""
echo "Press Ctrl+C to stop"
echo "======================================================"

# Run the auto-sync service
cd "$(dirname "$0")/.."
go run cmd/sync/main.go