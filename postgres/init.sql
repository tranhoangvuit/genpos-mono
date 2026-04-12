-- Enable logical replication for PowerSync
ALTER SYSTEM SET wal_level = 'logical';

-- Create a publication for PowerSync to track all tables
CREATE PUBLICATION powersync FOR ALL TABLES;
