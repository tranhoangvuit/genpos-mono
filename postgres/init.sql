-- Enable logical replication for PowerSync
ALTER SYSTEM SET wal_level = 'logical';

-- Create a publication for PowerSync to track all tables (genpos_dev)
CREATE PUBLICATION powersync FOR ALL TABLES;

-- Create the test database
CREATE DATABASE genpos_test;

-- Connect to genpos_test and set up the same publication
\c genpos_test
CREATE PUBLICATION powersync FOR ALL TABLES;
