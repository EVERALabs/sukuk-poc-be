-- Migration: Add blockchain-related fields for new indexer architecture
-- Date: 2024-01-15

-- Add log_index to investments table for proper event deduplication
ALTER TABLE investments ADD COLUMN IF NOT EXISTS log_index INTEGER NOT NULL DEFAULT 0;

-- Add token_price to investments table to track purchase price
ALTER TABLE investments ADD COLUMN IF NOT EXISTS token_price VARCHAR(78) NOT NULL DEFAULT '0';

-- Add distribution_id to yield_claims table to track yield distributions
ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS distribution_id BIGINT NOT NULL DEFAULT 0;

-- Add external_id to redemptions table for blockchain redemption ID tracking
ALTER TABLE redemptions ADD COLUMN IF NOT EXISTS external_id VARCHAR(66) UNIQUE;

-- Create system_states table to track sync progress
CREATE TABLE IF NOT EXISTS system_states (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create unique index on (transaction_hash, log_index) for investments
-- First remove the old unique constraint on just transaction_hash
ALTER TABLE investments DROP CONSTRAINT IF EXISTS investments_transaction_hash_key;
ALTER TABLE investments DROP CONSTRAINT IF EXISTS idx_investments_transaction_hash;

-- Add new unique constraint on (transaction_hash, log_index)
CREATE UNIQUE INDEX IF NOT EXISTS idx_investments_tx_hash_log_index 
ON investments(transaction_hash, log_index) WHERE deleted_at IS NULL;

-- Add index on distribution_id for yield_claims
CREATE INDEX IF NOT EXISTS idx_yield_claims_distribution_id ON yield_claims(distribution_id);

-- Add index on log_index for investments
CREATE INDEX IF NOT EXISTS idx_investments_log_index ON investments(log_index);

-- Insert initial system state for last processed event ID
INSERT INTO system_states (key, value) 
VALUES ('last_processed_event_id', '0') 
ON CONFLICT (key) DO NOTHING;

-- Create comments for documentation
COMMENT ON COLUMN investments.log_index IS 'Log index from blockchain event for deduplication';
COMMENT ON COLUMN investments.token_price IS 'Token price at time of investment purchase';
COMMENT ON COLUMN yield_claims.distribution_id IS 'Links to YieldDistributed event from blockchain';
COMMENT ON COLUMN redemptions.external_id IS 'Blockchain redemption ID (bytes32) for tracking';
COMMENT ON TABLE system_states IS 'System configuration and state tracking';