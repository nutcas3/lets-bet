CREATE TABLE IF NOT EXISTS mpesa_deposits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- M-Pesa transaction identifiers
    merchant_request_id VARCHAR(50) NOT NULL,
    checkout_request_id VARCHAR(50) NOT NULL UNIQUE,
    
    -- User and amount details
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone_number VARCHAR(20) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'KES',
    
    -- Transaction status and metadata
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, COMPLETED, FAILED, CANCELLED
    reference VARCHAR(100) NOT NULL, -- Account reference from STK push
    description TEXT,
    
    -- M-Pesa callback data
    mpesa_receipt_number VARCHAR(50), -- When completed
    transaction_date TIMESTAMP, -- When completed
    result_code VARCHAR(10), -- M-Pesa result code
    result_desc TEXT, -- M-Pesa result description
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_mpesa_deposits_user_id ON mpesa_deposits(user_id);
CREATE INDEX IF NOT EXISTS idx_mpesa_deposits_checkout_request_id ON mpesa_deposits(checkout_request_id);
CREATE INDEX IF NOT EXISTS idx_mpesa_deposits_merchant_request_id ON mpesa_deposits(merchant_request_id);
CREATE INDEX IF NOT EXISTS idx_mpesa_deposits_status ON mpesa_deposits(status);
CREATE INDEX IF NOT EXISTS idx_mpesa_deposits_phone_number ON mpesa_deposits(phone_number);
CREATE INDEX IF NOT EXISTS idx_mpesa_deposits_created_at ON mpesa_deposits(created_at);

-- Update trigger
CREATE TRIGGER update_mpesa_deposits_updated_at 
    BEFORE UPDATE ON mpesa_deposits 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
