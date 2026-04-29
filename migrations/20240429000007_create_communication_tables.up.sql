-- Create communication tables for SMS/OTP tracking

-- Create sms_logs table for tracking SMS messages
CREATE TABLE IF NOT EXISTS sms_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id VARCHAR(255) UNIQUE NOT NULL, -- External message ID from Africa's Talking
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    phone_number VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    sender_name VARCHAR(50) NOT NULL DEFAULT 'BettingPlatform',
    message_type VARCHAR(50) NOT NULL, -- 'welcome', 'verification', 'deposit', 'withdrawal', 'bet', 'win', 'promotional', etc.
    
    -- Status information
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SENT', 'DELIVERED', 'FAILED', 'REJECTED')),
    status_code VARCHAR(10),
    error_message TEXT,
    
    -- Cost information
    cost DECIMAL(10,4) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Provider information
    provider VARCHAR(50) NOT NULL DEFAULT 'africastalking',
    provider_data JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT sms_logs_status_check CHECK (status IN ('PENDING', 'SENT', 'DELIVERED', 'FAILED', 'REJECTED'))
);

-- Create otp_logs table for tracking OTP requests and verifications
CREATE TABLE IF NOT EXISTS otp_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id VARCHAR(255) UNIQUE NOT NULL, -- External transaction ID
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    phone_number VARCHAR(20) NOT NULL,
    brand_name VARCHAR(50) NOT NULL DEFAULT 'BettingPlatform',
    otp_code VARCHAR(10), -- Only stored for debugging, should be encrypted in production
    otp_length INTEGER NOT NULL DEFAULT 6,
    time_to_live INTEGER NOT NULL DEFAULT 300, -- TTL in seconds
    
    -- Status information
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SENT', 'VERIFIED', 'EXPIRED', 'FAILED')),
    verification_attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    
    -- Provider information
    provider VARCHAR(50) NOT NULL DEFAULT 'africastalking',
    provider_data JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT otp_logs_status_check CHECK (status IN ('PENDING', 'SENT', 'VERIFIED', 'EXPIRED', 'FAILED'))
);

-- Create ussd_logs table for tracking USSD sessions
CREATE TABLE IF NOT EXISTS ussd_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    phone_number VARCHAR(20) NOT NULL,
    input_text TEXT NOT NULL,
    response_text TEXT NOT NULL,
    session_status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (session_status IN ('ACTIVE', 'COMPLETED', 'TIMEOUT', 'ERROR')),
    
    -- Provider information
    provider VARCHAR(50) NOT NULL DEFAULT 'africastalking',
    provider_data JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT ussd_logs_status_check CHECK (session_status IN ('ACTIVE', 'COMPLETED', 'TIMEOUT', 'ERROR'))
);

-- Create voice_logs table for tracking voice calls
CREATE TABLE IF NOT EXISTS voice_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id VARCHAR(255) UNIQUE NOT NULL, -- External call ID
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    phone_number VARCHAR(20) NOT NULL,
    from_number VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    call_status VARCHAR(20) NOT NULL DEFAULT 'INITIATED' CHECK (call_status IN ('INITIATED', 'CONNECTED', 'COMPLETED', 'FAILED', 'BUSY', 'NO_ANSWER')),
    
    -- Call details
    duration INTEGER DEFAULT 0, -- Call duration in seconds
    recorded BOOLEAN DEFAULT FALSE,
    recording_url TEXT,
    
    -- Cost information
    cost DECIMAL(10,4) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Provider information
    provider VARCHAR(50) NOT NULL DEFAULT 'africastalking',
    provider_data JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    connected_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT voice_logs_status_check CHECK (call_status IN ('INITIATED', 'CONNECTED', 'COMPLETED', 'FAILED', 'BUSY', 'NO_ANSWER'))
);

-- Create communication_preferences table for user communication settings
CREATE TABLE IF NOT EXISTS communication_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- SMS preferences
    sms_enabled BOOLEAN DEFAULT TRUE,
    sms_welcome BOOLEAN DEFAULT TRUE,
    sms_verification BOOLEAN DEFAULT TRUE,
    sms_deposit BOOLEAN DEFAULT TRUE,
    sms_withdrawal BOOLEAN DEFAULT TRUE,
    sms_bet_confirmation BOOLEAN DEFAULT TRUE,
    sms_win_notifications BOOLEAN DEFAULT TRUE,
    sms_promotional BOOLEAN DEFAULT FALSE,
    
    -- OTP preferences
    otp_enabled BOOLEAN DEFAULT TRUE,
    otp_via_sms BOOLEAN DEFAULT TRUE,
    otp_via_voice BOOLEAN DEFAULT FALSE,
    
    -- Voice preferences
    voice_enabled BOOLEAN DEFAULT FALSE,
    voice_notifications BOOLEAN DEFAULT FALSE,
    
    -- USSD preferences
    ussd_enabled BOOLEAN DEFAULT TRUE,
    
    -- Rate limiting preferences
    max_sms_per_hour INTEGER DEFAULT 10,
    max_otp_per_hour INTEGER DEFAULT 5,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create communication_templates table for message templates
CREATE TABLE IF NOT EXISTS communication_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_name VARCHAR(100) UNIQUE NOT NULL,
    template_type VARCHAR(50) NOT NULL, -- 'sms', 'otp', 'voice', 'ussd'
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    subject VARCHAR(255),
    content TEXT NOT NULL,
    variables JSONB DEFAULT '{}', -- Template variables and their descriptions
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    
    -- Metadata
    description TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for sms_logs
CREATE INDEX IF NOT EXISTS idx_sms_logs_message_id ON sms_logs(message_id);
CREATE INDEX IF NOT EXISTS idx_sms_logs_user_id ON sms_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_sms_logs_phone_number ON sms_logs(phone_number);
CREATE INDEX IF NOT EXISTS idx_sms_logs_message_type ON sms_logs(message_type);
CREATE INDEX IF NOT EXISTS idx_sms_logs_status ON sms_logs(status);
CREATE INDEX IF NOT EXISTS idx_sms_logs_provider ON sms_logs(provider);
CREATE INDEX IF NOT EXISTS idx_sms_logs_created_at ON sms_logs(created_at);

-- Create indexes for otp_logs
CREATE INDEX IF NOT EXISTS idx_otp_logs_transaction_id ON otp_logs(transaction_id);
CREATE INDEX IF NOT EXISTS idx_otp_logs_user_id ON otp_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_otp_logs_phone_number ON otp_logs(phone_number);
CREATE INDEX IF NOT EXISTS idx_otp_logs_status ON otp_logs(status);
CREATE INDEX IF NOT EXISTS idx_otp_logs_expires_at ON otp_logs(expires_at);
CREATE INDEX IF NOT EXISTS idx_otp_logs_provider ON otp_logs(provider);
CREATE INDEX IF NOT EXISTS idx_otp_logs_created_at ON otp_logs(created_at);

-- Create indexes for ussd_logs
CREATE INDEX IF NOT EXISTS idx_ussd_logs_session_id ON ussd_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_ussd_logs_user_id ON ussd_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_ussd_logs_phone_number ON ussd_logs(phone_number);
CREATE INDEX IF NOT EXISTS idx_ussd_logs_session_status ON ussd_logs(session_status);
CREATE INDEX IF NOT EXISTS idx_ussd_logs_provider ON ussd_logs(provider);
CREATE INDEX IF NOT EXISTS idx_ussd_logs_created_at ON ussd_logs(created_at);

-- Create indexes for voice_logs
CREATE INDEX IF NOT EXISTS idx_voice_logs_call_id ON voice_logs(call_id);
CREATE INDEX IF NOT EXISTS idx_voice_logs_user_id ON voice_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_voice_logs_phone_number ON voice_logs(phone_number);
CREATE INDEX IF NOT EXISTS idx_voice_logs_call_status ON voice_logs(call_status);
CREATE INDEX IF NOT EXISTS idx_voice_logs_provider ON voice_logs(provider);
CREATE INDEX IF NOT EXISTS idx_voice_logs_created_at ON voice_logs(created_at);

-- Create indexes for communication_preferences
CREATE INDEX IF NOT EXISTS idx_communication_preferences_user_id ON communication_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_communication_preferences_sms_enabled ON communication_preferences(sms_enabled);
CREATE INDEX IF NOT EXISTS idx_communication_preferences_otp_enabled ON communication_preferences(otp_enabled);

-- Create indexes for communication_templates
CREATE INDEX IF NOT EXISTS idx_communication_templates_template_name ON communication_templates(template_name);
CREATE INDEX IF NOT EXISTS idx_communication_templates_template_type ON communication_templates(template_type);
CREATE INDEX IF NOT EXISTS idx_communication_templates_language ON communication_templates(language);
CREATE INDEX IF NOT EXISTS idx_communication_templates_is_active ON communication_templates(is_active);

-- Create updated_at triggers
CREATE OR REPLACE FUNCTION update_sms_logs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_sms_logs_updated_at
    BEFORE UPDATE ON sms_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_sms_logs_updated_at();

CREATE OR REPLACE FUNCTION update_ussd_logs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_ussd_logs_updated_at
    BEFORE UPDATE ON ussd_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_ussd_logs_updated_at();

CREATE OR REPLACE FUNCTION update_communication_preferences_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_communication_preferences_updated_at
    BEFORE UPDATE ON communication_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_communication_preferences_updated_at();

CREATE OR REPLACE FUNCTION update_communication_templates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_communication_templates_updated_at
    BEFORE UPDATE ON communication_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_communication_templates_updated_at();

-- Add comments
COMMENT ON TABLE sms_logs IS 'Logs all SMS messages sent through Africa\'s Talking';
COMMENT ON TABLE otp_logs IS 'Logs all OTP requests and verifications';
COMMENT ON TABLE ussd_logs IS 'Logs USSD session interactions';
COMMENT ON TABLE voice_logs IS 'Logs voice call details';
COMMENT ON TABLE communication_preferences IS 'User communication preferences and settings';
COMMENT ON TABLE communication_templates IS 'Message templates for different communication types';

-- Add column comments for sms_logs
COMMENT ON COLUMN sms_logs.id IS 'Primary key';
COMMENT ON COLUMN sms_logs.message_id IS 'External message ID from provider';
COMMENT ON COLUMN sms_logs.user_id IS 'User who sent/received the message';
COMMENT ON COLUMN sms_logs.phone_number IS 'Recipient phone number';
COMMENT ON COLUMN sms_logs.message IS 'SMS message content';
COMMENT ON COLUMN sms_logs.sender_name IS 'Sender name displayed to recipient';
COMMENT ON COLUMN sms_logs.message_type IS 'Type of message (welcome, verification, etc.)';
COMMENT ON COLUMN sms_logs.status IS 'Message delivery status';
COMMENT ON COLUMN sms_logs.status_code IS 'Provider status code';
COMMENT ON COLUMN sms_logs.error_message IS 'Error message if failed';
COMMENT ON COLUMN sms_logs.cost IS 'Cost of sending the message';
COMMENT ON COLUMN sms_logs.provider IS 'Communication provider';
COMMENT ON COLUMN sms_logs.provider_data IS 'Raw provider response data';

-- Add column comments for otp_logs
COMMENT ON COLUMN otp_logs.id IS 'Primary key';
COMMENT ON COLUMN otp_logs.transaction_id IS 'External transaction ID';
COMMENT ON COLUMN otp_logs.user_id IS 'User who requested OTP';
COMMENT ON COLUMN otp_logs.phone_number IS 'Phone number for OTP';
COMMENT ON COLUMN otp_logs.brand_name IS 'Brand name shown to user';
COMMENT ON COLUMN otp_logs.otp_code IS 'OTP code (encrypted in production)';
COMMENT ON COLUMN otp_logs.otp_length IS 'Length of OTP code';
COMMENT ON COLUMN otp_logs.time_to_live IS 'OTP validity period in seconds';
COMMENT ON COLUMN otp_logs.status IS 'OTP status';
COMMENT ON COLUMN otp_logs.verification_attempts IS 'Number of verification attempts';
COMMENT ON COLUMN otp_logs.max_attempts IS 'Maximum allowed attempts';
COMMENT ON COLUMN otp_logs.provider IS 'Communication provider';
COMMENT ON COLUMN otp_logs.expires_at IS 'When OTP expires';

-- Add column comments for communication_preferences
COMMENT ON COLUMN communication_preferences.id IS 'Primary key';
COMMENT ON COLUMN communication_preferences.user_id IS 'User ID';
COMMENT ON COLUMN communication_preferences.sms_enabled IS 'Whether SMS is enabled';
COMMENT ON COLUMN communication_preferences.sms_welcome IS 'Send welcome SMS';
COMMENT ON COLUMN communication_preferences.sms_verification IS 'Send verification SMS';
COMMENT ON COLUMN communication_preferences.sms_deposit IS 'Send deposit notifications';
COMMENT ON COLUMN communication_preferences.sms_withdrawal IS 'Send withdrawal notifications';
COMMENT ON COLUMN communication_preferences.sms_bet_confirmation IS 'Send bet confirmations';
COMMENT ON COLUMN communication_preferences.sms_win_notifications IS 'Send win notifications';
COMMENT ON COLUMN communication_preferences.sms_promotional IS 'Send promotional messages';
COMMENT ON COLUMN communication_preferences.otp_enabled IS 'Whether OTP is enabled';
COMMENT ON COLUMN communication_preferences.otp_via_sms IS 'Send OTP via SMS';
COMMENT ON COLUMN communication_preferences.otp_via_voice IS 'Send OTP via voice call';
COMMENT ON COLUMN communication_preferences.voice_enabled IS 'Whether voice calls are enabled';
COMMENT ON COLUMN communication_preferences.voice_notifications IS 'Send voice notifications';
COMMENT ON COLUMN communication_preferences.ussd_enabled IS 'Whether USSD is enabled';
COMMENT ON COLUMN communication_preferences.max_sms_per_hour IS 'Maximum SMS per hour';
COMMENT ON COLUMN communication_preferences.max_otp_per_hour IS 'Maximum OTP attempts per hour';
