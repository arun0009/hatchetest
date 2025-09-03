-- Initialize Hatchet database
-- This script runs when the PostgreSQL container starts for the first time

-- Create the hatchet database if it doesn't exist
-- (PostgreSQL creates it automatically based on POSTGRES_DB environment variable)

-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE hatchet TO hatchet;

-- Create extensions that Hatchet might need
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set timezone
SET timezone = 'UTC';

-- Log initialization
DO $$
BEGIN
    RAISE NOTICE 'Hatchet database initialized successfully';
END $$;
