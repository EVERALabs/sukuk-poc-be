-- Add distribution_date column to yield_claims table

-- First add the column as nullable
ALTER TABLE yield_claims ADD COLUMN IF NOT EXISTS distribution_date TIMESTAMP WITH TIME ZONE;

-- Set distribution_date based on existing data
-- Use period_start as a default approximation, or created_at if period_start is null
UPDATE yield_claims 
SET distribution_date = COALESCE(period_start, created_at, NOW())
WHERE distribution_date IS NULL;

-- Now make it NOT NULL
ALTER TABLE yield_claims ALTER COLUMN distribution_date SET NOT NULL;

-- Create index for better performance
CREATE INDEX IF NOT EXISTS idx_yield_claims_distribution_date ON yield_claims(distribution_date);