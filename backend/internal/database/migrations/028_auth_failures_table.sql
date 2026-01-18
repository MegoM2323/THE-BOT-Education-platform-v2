-- 028_auth_failures_table.sql
-- Auth Failures tracking table for rate limiting and security monitoring
-- Stores failed authentication attempts with IP address, email, and reason

-- Create auth_failures table
CREATE TABLE auth_failures (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    reason VARCHAR(255) NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
-- Index for querying failures by email (user-based rate limiting)
CREATE INDEX idx_auth_failures_email ON auth_failures(email);

-- Index for querying failures by IP address (IP-based rate limiting)
CREATE INDEX idx_auth_failures_ip_address ON auth_failures(ip_address);

-- Index for querying recent failures (used in time-window queries)
CREATE INDEX idx_auth_failures_created_at ON auth_failures(created_at);

-- Composite index for efficient email + time-window queries
CREATE INDEX idx_auth_failures_email_time ON auth_failures(email, created_at DESC);

-- Composite index for efficient IP + time-window queries
CREATE INDEX idx_auth_failures_ip_time ON auth_failures(ip_address, created_at DESC);

-- Composite index with created_at DESC for efficiently querying recent failures
-- (We don't use a partial index with CURRENT_TIMESTAMP as it requires IMMUTABLE functions)
-- Application will filter failures by time window in code instead

-- Comments for table and columns
COMMENT ON TABLE auth_failures IS 'Records of failed authentication attempts for rate limiting and security monitoring';
COMMENT ON COLUMN auth_failures.id IS 'Unique identifier for this failure record';
COMMENT ON COLUMN auth_failures.email IS 'Email address associated with the failed login attempt';
COMMENT ON COLUMN auth_failures.ip_address IS 'IP address from which the failed attempt was made (supports IPv4 and IPv6)';
COMMENT ON COLUMN auth_failures.reason IS 'Reason for authentication failure (e.g., invalid_password, user_not_found, account_locked, account_deleted)';
COMMENT ON COLUMN auth_failures.user_agent IS 'User agent string from the request (for forensics and analysis)';
COMMENT ON COLUMN auth_failures.created_at IS 'Timestamp when the authentication failure occurred';
