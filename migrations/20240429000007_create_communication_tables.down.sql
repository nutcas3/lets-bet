-- Drop communication tables in reverse order of creation

-- Drop triggers first
DROP TRIGGER IF EXISTS update_communication_templates_updated_at ON communication_templates;
DROP TRIGGER IF EXISTS update_communication_preferences_updated_at ON communication_preferences;
DROP TRIGGER IF EXISTS update_ussd_logs_updated_at ON ussd_logs;
DROP TRIGGER IF EXISTS update_sms_logs_updated_at ON sms_logs;

-- Drop functions
DROP FUNCTION IF EXISTS update_communication_templates_updated_at();
DROP FUNCTION IF EXISTS update_communication_preferences_updated_at();
DROP FUNCTION IF EXISTS update_ussd_logs_updated_at();
DROP FUNCTION IF EXISTS update_sms_logs_updated_at();

-- Drop tables
DROP TABLE IF EXISTS communication_templates;
DROP TABLE IF EXISTS communication_preferences;
DROP TABLE IF EXISTS voice_logs;
DROP TABLE IF EXISTS ussd_logs;
DROP TABLE IF EXISTS otp_logs;
DROP TABLE IF EXISTS sms_logs;
