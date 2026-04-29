-- Create Flutterwave deposits table
CREATE TABLE IF NOT EXISTS flutterwave_deposits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    deposit_id VARCHAR(255) UNIQUE NOT NULL,
    
    -- Transaction details
    transaction_id VARCHAR(255) NOT NULL,
    reference VARCHAR(255) NOT NULL,
    payment_link VARCHAR(500),
    
    -- Amount and currency
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'NGN',
    
    -- Customer information
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    
    -- Status and metadata
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'CANCELLED')),
    payment_method VARCHAR(50),
    provider_name VARCHAR(50) NOT NULL DEFAULT 'flutterwave',
    provider_txn_id VARCHAR(255) NOT NULL,
    
    -- Flutterwave specific fields
    flutterwave_ref VARCHAR(255),
    network VARCHAR(50), -- For mobile money (MTN, VODAFONE, TIGO, AIRTEL)
    
    -- Processing information
    processed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    failure_reason TEXT,
    
    -- Fees
    app_fee DECIMAL(15,2) DEFAULT 0,
    merchant_fee DECIMAL(15,2) DEFAULT 0,
    total_fees DECIMAL(15,2) GENERATED ALWAYS AS (app_fee + merchant_fee) STORED,
    
    -- Metadata
    meta JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_flutterwave_deposits_user_id ON flutterwave_deposits(user_id);
CREATE INDEX IF NOT EXISTS idx_flutterwave_deposits_status ON flutterwave_deposits(status);
CREATE INDEX IF NOT EXISTS idx_flutterwave_deposits_transaction_id ON flutterwave_deposits(transaction_id);
CREATE INDEX IF NOT EXISTS idx_flutterwave_deposits_reference ON flutterwave_deposits(reference);
CREATE INDEX IF NOT EXISTS idx_flutterwave_deposits_provider_txn_id ON flutterwave_deposits(provider_txn_id);
CREATE INDEX IF NOT EXISTS idx_flutterwave_deposits_created_at ON flutterwave_deposits(created_at);

-- Create index for Flutterwave reference
CREATE INDEX IF NOT EXISTS idx_flutterwave_deposits_flutterwave_ref ON flutterwave_deposits(flutterwave_ref) WHERE flutterwave_ref IS NOT NULL;

-- Create index for mobile money network
CREATE INDEX IF NOT EXISTS idx_flutterwave_deposits_network ON flutterwave_deposits(network) WHERE network IS NOT NULL;

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_flutterwave_deposits_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_flutterwave_deposits_updated_at
    BEFORE UPDATE ON flutterwave_deposits
    FOR EACH ROW
    EXECUTE FUNCTION update_flutterwave_deposits_updated_at();

-- Add comments
COMMENT ON TABLE flutterwave_deposits IS 'Stores Flutterwave deposit transactions for Nigeria and Ghana markets';
COMMENT ON COLUMN flutterwave_deposits.id IS 'Primary key';
COMMENT ON COLUMN flutterwave_deposits.user_id IS 'Foreign key to users table';
COMMENT ON COLUMN flutterwave_deposits.deposit_id IS 'Internal deposit identifier';
COMMENT ON COLUMN flutterwave_deposits.transaction_id IS 'Flutterwave transaction ID (flw_ref)';
COMMENT ON COLUMN flutterwave_deposits.reference IS 'Flutterwave transaction reference (tx_ref)';
COMMENT ON COLUMN flutterwave_deposits.payment_link IS 'Flutterwave payment link for card/bank transfers';
COMMENT ON COLUMN flutterwave_deposits.amount IS 'Deposit amount';
COMMENT ON COLUMN flutterwave_deposits.currency IS 'Deposit currency (NGN for Nigeria, GHS for Ghana)';
COMMENT ON COLUMN flutterwave_deposits.email IS 'Customer email';
COMMENT ON COLUMN flutterwave_deposits.phone_number IS 'Customer phone number';
COMMENT ON COLUMN flutterwave_deposits.status IS 'Deposit status';
COMMENT ON COLUMN flutterwave_deposits.payment_method IS 'Payment method used';
COMMENT ON COLUMN flutterwave_deposits.provider_name IS 'Payment provider name';
COMMENT ON COLUMN flutterwave_deposits.provider_txn_id IS 'Provider transaction ID';
COMMENT ON COLUMN flutterwave_deposits.flutterwave_ref IS 'Flutterwave reference';
COMMENT ON COLUMN flutterwave_deposits.network IS 'Mobile money network (for Ghana)';
COMMENT ON COLUMN flutterwave_deposits.processed_at IS 'When the deposit was processed';
COMMENT ON COLUMN flutterwave_deposits.completed_at IS 'When the deposit was completed';
COMMENT ON COLUMN flutterwave_deposits.failure_reason IS 'Reason for failure if applicable';
COMMENT ON COLUMN flutterwave_deposits.app_fee IS 'Flutterwave application fee';
COMMENT ON COLUMN flutterwave_deposits.merchant_fee IS 'Flutterwave merchant fee';
COMMENT ON COLUMN flutterwave_deposits.total_fees IS 'Total fees (app_fee + merchant_fee)';
COMMENT ON COLUMN flutterwave_deposits.meta IS 'Additional metadata in JSON format';
COMMENT ON COLUMN flutterwave_deposits.created_at IS 'When the deposit was created';
COMMENT ON COLUMN flutterwave_deposits.updated_at IS 'When the deposit was last updated';
