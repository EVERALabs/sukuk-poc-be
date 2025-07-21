-- Migration script to update database schema for new model structure

-- 1. Update investments table
-- First, add new columns as nullable
ALTER TABLE investments ADD COLUMN IF NOT EXISTS investment_amount VARCHAR(78);
ALTER TABLE investments ADD COLUMN IF NOT EXISTS token_price VARCHAR(78);
ALTER TABLE investments ADD COLUMN IF NOT EXISTS tx_hash VARCHAR(66);
ALTER TABLE investments ADD COLUMN IF NOT EXISTS log_index INTEGER;

-- Copy data from old columns to new ones
UPDATE investments SET investment_amount = amount WHERE investment_amount IS NULL AND amount IS NOT NULL;
UPDATE investments SET tx_hash = transaction_hash WHERE tx_hash IS NULL AND transaction_hash IS NOT NULL;
UPDATE investments SET log_index = 0 WHERE log_index IS NULL;
UPDATE investments SET token_price = '1000000000000000000' WHERE token_price IS NULL; -- Default 1:1 ratio

-- Now make the columns NOT NULL
ALTER TABLE investments ALTER COLUMN investment_amount SET NOT NULL;
ALTER TABLE investments ALTER COLUMN token_price SET NOT NULL;
ALTER TABLE investments ALTER COLUMN tx_hash SET NOT NULL;
ALTER TABLE investments ALTER COLUMN log_index SET NOT NULL;

-- Drop old columns
ALTER TABLE investments DROP COLUMN IF EXISTS amount;
ALTER TABLE investments DROP COLUMN IF EXISTS transaction_hash;
ALTER TABLE investments DROP COLUMN IF EXISTS block_number;
ALTER TABLE investments DROP COLUMN IF EXISTS investor_email;
ALTER TABLE investments DROP COLUMN IF EXISTS investment_id;

-- Rename sukuk_series_id to sukuk_id using column definition in Gorm
-- This is handled by Gorm's column tag, no SQL change needed

-- 2. Update yield_claims table (if it exists)
-- Check if table exists before altering
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'yield_claims') THEN
        -- Add new columns
        ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS dist_tx_hash VARCHAR(66);
        ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS dist_log_index INTEGER;
        ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS claim_tx_hash VARCHAR(66);
        ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS claim_log_index INTEGER;
        ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS period_start TIMESTAMP WITH TIME ZONE;
        ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS period_end TIMESTAMP WITH TIME ZONE;
        
        -- Copy/set data
        UPDATE yield_claims SET dist_tx_hash = transaction_hash WHERE dist_tx_hash IS NULL AND transaction_hash IS NOT NULL;
        UPDATE yield_claims SET dist_log_index = 0 WHERE dist_log_index IS NULL;
        UPDATE yield_claims SET period_start = distribution_date - INTERVAL '3 months' WHERE period_start IS NULL;
        UPDATE yield_claims SET period_end = distribution_date WHERE period_end IS NULL;
        
        -- Make required columns NOT NULL
        ALTER TABLE yield_claims ALTER COLUMN dist_tx_hash SET NOT NULL;
        ALTER TABLE yield_claims ALTER COLUMN dist_log_index SET NOT NULL;
        ALTER TABLE yield_claims ALTER COLUMN period_start SET NOT NULL;
        ALTER TABLE yield_claims ALTER COLUMN period_end SET NOT NULL;
        
        -- Drop old columns
        ALTER TABLE yield_claims DROP COLUMN IF EXISTS transaction_hash;
        ALTER TABLE yield_claims DROP COLUMN IF EXISTS block_number;
        ALTER TABLE yield_claims DROP COLUMN IF EXISTS investment_id;
        ALTER TABLE yield_claims DROP COLUMN IF EXISTS distribution_id;
        ALTER TABLE yield_claims DROP COLUMN IF EXISTS expires_at;
    END IF;
END $$;

-- 3. Update redemptions table (if it exists)
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'redemptions') THEN
        -- Add new columns
        ALTER TABLE redemptions ADD COLUMN IF NOT EXISTS request_tx_hash VARCHAR(66);
        ALTER TABLE redemptions ADD COLUMN IF NOT EXISTS request_log_index INTEGER;
        ALTER TABLE redemptions ADD COLUMN IF NOT EXISTS complete_tx_hash VARCHAR(66);
        ALTER TABLE redemptions ADD COLUMN IF NOT EXISTS complete_log_index INTEGER;
        ALTER TABLE redemptions ADD COLUMN IF NOT EXISTS request_date TIMESTAMP WITH TIME ZONE;
        
        -- Copy/set data
        UPDATE redemptions SET request_tx_hash = transaction_hash WHERE request_tx_hash IS NULL AND transaction_hash IS NOT NULL;
        UPDATE redemptions SET request_log_index = 0 WHERE request_log_index IS NULL;
        UPDATE redemptions SET request_date = requested_at WHERE request_date IS NULL AND requested_at IS NOT NULL;
        UPDATE redemptions SET request_date = created_at WHERE request_date IS NULL;
        
        -- Make required columns NOT NULL
        ALTER TABLE redemptions ALTER COLUMN request_tx_hash SET NOT NULL;
        ALTER TABLE redemptions ALTER COLUMN request_log_index SET NOT NULL;
        ALTER TABLE redemptions ALTER COLUMN request_date SET NOT NULL;
        
        -- Drop old columns
        ALTER TABLE redemptions DROP COLUMN IF EXISTS transaction_hash;
        ALTER TABLE redemptions DROP COLUMN IF EXISTS block_number;
        ALTER TABLE redemptions DROP COLUMN IF EXISTS investment_id;
        ALTER TABLE redemptions DROP COLUMN IF EXISTS requested_at;
        ALTER TABLE redemptions DROP COLUMN IF EXISTS approved_at;
        ALTER TABLE redemptions DROP COLUMN IF EXISTS request_reason;
        ALTER TABLE redemptions DROP COLUMN IF EXISTS approval_notes;
    END IF;
END $$;

-- 4. Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_investments_tx_hash ON investments(tx_hash);
CREATE INDEX IF NOT EXISTS idx_investments_sukuk_id ON investments(sukuk_series_id);
CREATE INDEX IF NOT EXISTS idx_yield_claims_dist_tx_hash ON yield_claims(dist_tx_hash);
CREATE INDEX IF NOT EXISTS idx_yield_claims_sukuk_id ON yield_claims(sukuk_series_id);
CREATE INDEX IF NOT EXISTS idx_redemptions_request_tx_hash ON redemptions(request_tx_hash);
CREATE INDEX IF NOT EXISTS idx_redemptions_sukuk_id ON redemptions(sukuk_series_id);