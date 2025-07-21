-- Fix yield_claims table migration

-- Add missing columns for yield_claims
ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS dist_tx_hash VARCHAR(66);
ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS dist_log_index INTEGER;
ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS claim_tx_hash VARCHAR(66);
ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS claim_log_index INTEGER;

-- Copy data from existing columns
UPDATE yield_claims SET dist_tx_hash = transaction_hash WHERE dist_tx_hash IS NULL AND transaction_hash IS NOT NULL;
UPDATE yield_claims SET dist_log_index = block_number WHERE dist_log_index IS NULL AND block_number IS NOT NULL;
UPDATE yield_claims SET dist_log_index = 0 WHERE dist_log_index IS NULL;

-- Set NOT NULL constraints where appropriate (only for fields that should always have values)
-- Note: claim_tx_hash and claim_log_index can be NULL until claim is processed

-- Drop old columns that are no longer needed
ALTER TABLE yield_claims DROP COLUMN IF EXISTS block_number;
-- Keep transaction_hash as dist_tx_hash is the replacement

-- Create missing indexes
CREATE INDEX IF NOT EXISTS idx_yield_claims_dist_tx_hash ON yield_claims(dist_tx_hash);