-- Migration 000001: Initialization
-- This migration ensures UUID extension is available.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
