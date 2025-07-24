-- Add blockchain event tables for SukukPurchased and RedemptionRequested events

-- SukukPurchased events table
CREATE TABLE IF NOT EXISTS sukuk_purchased_events (
    id SERIAL PRIMARY KEY,
    buyer VARCHAR(42) NOT NULL,
    sukuk_address VARCHAR(42) NOT NULL,
    payment_token VARCHAR(42) NOT NULL,
    amount VARCHAR(78) NOT NULL,
    block_number BIGINT NOT NULL,
    tx_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    processed BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes for sukuk_purchased_events
CREATE INDEX idx_sukuk_purchased_buyer ON sukuk_purchased_events(buyer);
CREATE INDEX idx_sukuk_purchased_sukuk_address ON sukuk_purchased_events(sukuk_address);
CREATE INDEX idx_sukuk_purchased_tx_hash ON sukuk_purchased_events(tx_hash);
CREATE INDEX idx_sukuk_purchased_block_number ON sukuk_purchased_events(block_number);
CREATE INDEX idx_sukuk_purchased_timestamp ON sukuk_purchased_events(timestamp);
CREATE INDEX idx_sukuk_purchased_processed ON sukuk_purchased_events(processed);
CREATE INDEX idx_sukuk_purchased_deleted_at ON sukuk_purchased_events(deleted_at);
CREATE UNIQUE INDEX idx_sukuk_purchased_tx_log ON sukuk_purchased_events(tx_hash, log_index);

-- RedemptionRequested events table
CREATE TABLE IF NOT EXISTS redemption_requested_events (
    id SERIAL PRIMARY KEY,
    "user" VARCHAR(42) NOT NULL,
    sukuk_address VARCHAR(42) NOT NULL,
    amount VARCHAR(78) NOT NULL,
    payment_token VARCHAR(42) NOT NULL,
    block_number BIGINT NOT NULL,
    tx_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    processed BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes for redemption_requested_events
CREATE INDEX idx_redemption_requested_user ON redemption_requested_events("user");
CREATE INDEX idx_redemption_requested_sukuk_address ON redemption_requested_events(sukuk_address);
CREATE INDEX idx_redemption_requested_tx_hash ON redemption_requested_events(tx_hash);
CREATE INDEX idx_redemption_requested_block_number ON redemption_requested_events(block_number);
CREATE INDEX idx_redemption_requested_timestamp ON redemption_requested_events(timestamp);
CREATE INDEX idx_redemption_requested_processed ON redemption_requested_events(processed);
CREATE INDEX idx_redemption_requested_deleted_at ON redemption_requested_events(deleted_at);
CREATE UNIQUE INDEX idx_redemption_requested_tx_log ON redemption_requested_events(tx_hash, log_index);

-- Update trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_sukuk_purchased_events_updated_at BEFORE UPDATE
    ON sukuk_purchased_events FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_redemption_requested_events_updated_at BEFORE UPDATE
    ON redemption_requested_events FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();