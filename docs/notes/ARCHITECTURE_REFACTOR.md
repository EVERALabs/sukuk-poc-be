# Architecture Refactor: From Webhooks to Indexer Integration

## Overview

We've refactored the backend from a webhook-based event processing system to a polling-based indexer integration. This provides better reliability, data consistency, and simplified architecture.

## Old Architecture (Removed)

```
Blockchain → Indexer → Webhook POST → Backend → Database
```

**Problems:**
- Event loss potential
- No idempotency protection
- Race conditions
- Complex retry logic needed
- API key security concerns

## New Architecture (Current)

```
Blockchain → Indexer → Indexed DB (blockchain schema)
                           ↓
                    Backend Sync Service (polling)
                           ↓
                    Business DB (app schema)
                           ↓
                    API Endpoints → Frontend
```

## Key Changes

### 1. Removed Components
- `internal/handlers/events.go` - Webhook event handlers
- `internal/models/event.go` - Event model (using indexer's events table)
- Webhook endpoint `/admin/events/webhook`

### 2. Added Components
- `internal/services/blockchain_sync.go` - Polling sync service
- `internal/models/system_state.go` - System state tracking
- `internal/handlers/blockchain.go` - Combined data endpoints
- New API endpoints for blockchain + business data

### 3. Database Schema Changes
- Added `log_index` to investments table
- Added `distribution_id` to yield_claims table
- Added `system_states` table for sync tracking
- Removed unique constraint on transaction_hash only
- Added unique constraint on (transaction_hash, log_index)

## Data Flow

### Investment Flow (Example)
1. **Blockchain**: Investment transaction is mined
2. **Indexer**: Captures Investment event, stores in `blockchain.events`
3. **Sync Service**: Polls for new events every 10 seconds
4. **Backend**: Processes Investment event, creates `investment` record
5. **API**: Returns investment data to frontend

### Event Processing
The `BlockchainSyncService` polls `blockchain.events` table and processes:
- `SukukDeployed` → Updates sukuk series with token address
- `Investment` → Creates investment records
- `YieldDistributed` → Creates yield claim records
- `YieldClaimed` → Updates yield claim status
- `RedemptionRequested/Approved/Completed/Rejected` → Manages redemption lifecycle

## New API Endpoints

### Blockchain Data Access
- `GET /api/v1/blockchain/events/{txHash}` - Raw blockchain events
- `GET /api/v1/investments/{id}/blockchain` - Investment + blockchain data
- `GET /api/v1/portfolio/{address}/blockchain` - Portfolio + blockchain data
- `GET /api/v1/sukuks/{id}/blockchain-metrics` - Sukuk + blockchain metrics

### Existing Endpoints (Enhanced)
All existing endpoints continue to work but now use data synced from blockchain:
- `GET /api/v1/portfolio/{address}` - Portfolio overview
- `GET /api/v1/investments` - Investment list
- `GET /api/v1/yield-claims` - Yield claims
- `GET /api/v1/redemptions` - Redemptions

## Configuration

### Environment Variables
```env
# Database connection for indexer access
# The sync service needs read access to blockchain.events table
DB_HOST=localhost
DB_NAME=sukuk_db  # Should include both schemas: app and blockchain
```

### Sync Service
- **Polling Interval**: 10 seconds (configurable)
- **Batch Size**: 1000 events per sync
- **State Tracking**: Last processed event ID stored in `system_states`

## Benefits

### 1. Reliability
- No event loss (blockchain is source of truth)
- Automatic retry on sync failures
- Idempotent processing with (tx_hash, log_index) uniqueness

### 2. Data Consistency
- All events processed in order
- Atomic transactions for event processing
- Consistent state across all tables

### 3. Simplicity
- No webhook infrastructure needed
- No API key management for events
- Simple polling logic

### 4. Audit Trail
- Complete blockchain event history available
- Can replay events if needed
- Easy debugging with combined views

## Monitoring

### Health Checks
- Monitor `last_processed_event_id` in system_states
- Check sync service lag (compare with latest blockchain events)
- Monitor database connection health

### Metrics to Track
- Events processed per minute
- Sync service lag time
- Failed event processing count
- Database query performance

## Migration Steps

1. ✅ Remove webhook handlers
2. ✅ Create blockchain sync service  
3. ✅ Update database schema
4. ✅ Add new API endpoints
5. ✅ Update existing handlers with documentation
6. 🔄 Deploy and test
7. 🔄 Monitor sync performance

## Future Enhancements

### 1. Real-time Updates
Add WebSocket endpoints that push updates when new events are processed:
```javascript
// WebSocket connection for real-time portfolio updates
ws://api/v1/portfolio/{address}/live
```

### 2. Event Replay
Add admin endpoint to replay events from a specific block:
```http
POST /api/v1/admin/replay-events
{
  "from_block": 12345,
  "to_block": 12350
}
```

### 3. Multi-chain Support
Extend to support multiple blockchains:
- Add `chain_id` filtering
- Separate sync services per chain
- Chain-specific configuration

## Troubleshooting

### Common Issues

**Sync Service Not Processing Events**
- Check database connection to blockchain schema
- Verify `last_processed_event_id` is updating
- Check service logs for errors

**Missing Investment Records**
- Verify blockchain events exist in indexer database
- Check for processing errors in logs
- Ensure unique constraints are not causing conflicts

**Stale Data**
- Check sync service is running (`ps aux | grep blockchain_sync`)
- Monitor sync lag between indexer and business database
- Verify database transactions are committing properly

### Debug Queries

```sql
-- Check last processed event
SELECT * FROM system_states WHERE key = 'last_processed_event_id';

-- Check latest blockchain events
SELECT * FROM blockchain.events ORDER BY id DESC LIMIT 10;

-- Check sync lag
SELECT 
  (SELECT MAX(id) FROM blockchain.events) as latest_blockchain_event,
  (SELECT value::bigint FROM system_states WHERE key = 'last_processed_event_id') as last_processed,
  (SELECT MAX(id) FROM blockchain.events) - (SELECT value::bigint FROM system_states WHERE key = 'last_processed_event_id') as lag;
```