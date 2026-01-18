-- Create trial_requests table for landing page form submissions
CREATE TABLE IF NOT EXISTS trial_requests (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    telegram VARCHAR(50) NOT NULL,
    email VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for efficient ordering by creation date
CREATE INDEX idx_trial_requests_created_at ON trial_requests(created_at DESC);

-- Add comment to table
COMMENT ON TABLE trial_requests IS 'Trial lesson requests from landing page';
COMMENT ON COLUMN trial_requests.name IS 'Full name of the person requesting a trial lesson';
COMMENT ON COLUMN trial_requests.phone IS 'Contact phone number';
COMMENT ON COLUMN trial_requests.telegram IS 'Telegram username or handle';
COMMENT ON COLUMN trial_requests.email IS 'Optional email address';
