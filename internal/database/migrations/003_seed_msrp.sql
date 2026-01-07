-- Migration: 003_seed_msrp
-- Purpose: Seed MSRP reference data for Pokemon TCG products (2024-2025 sets)
-- Date: 2026-01-07
-- Feature: User Story 2 - Calculate Product Profitability

-- Insert or update products with MSRP data
-- Using ON CONFLICT to update existing products or insert new ones

-- ============================================
-- DISPLAYS (Booster Boxes - 36 boosters)
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
-- Scarlet & Violet Base Sets
('Display Ecarlate et Violet', 'display', 'Ecarlate et Violet', 'sv-base', 159.99),
('Display Evolutions a Paldea', 'display', 'Evolutions a Paldea', 'sv-paldea-evolved', 159.99),
('Display Obsidian Flames', 'display', 'Obsidian Flames', 'sv-obsidian-flames', 159.99),

-- Scarlet & Violet 151
('Display Ecarlate et Violet 151', 'display', 'Ecarlate et Violet 151', 'sv-151', 159.99),
('Display SV 151', 'display', 'Ecarlate et Violet 151', 'sv-151', 159.99),
('Display Pokemon 151', 'display', 'Ecarlate et Violet 151', 'sv-151', 159.99),

-- Paldean Fates
('Display Destinees de Paldea', 'display', 'Destinees de Paldea', 'sv-paldean-fates', 159.99),
('Display Paldean Fates', 'display', 'Destinees de Paldea', 'sv-paldean-fates', 159.99),

-- Temporal Forces
('Display Forces Temporelles', 'display', 'Forces Temporelles', 'sv-temporal-forces', 159.99),
('Display Temporal Forces', 'display', 'Forces Temporelles', 'sv-temporal-forces', 159.99),

-- Twilight Masquerade
('Display Masques du Crepuscule', 'display', 'Masques du Crepuscule', 'sv-twilight-masque', 159.99),
('Display Twilight Masquerade', 'display', 'Masques du Crepuscule', 'sv-twilight-masque', 159.99),

-- Shrouded Fable
('Display Fable Nimbee', 'display', 'Fable Nimbee', 'sv-shrouded-fable', 159.99),
('Display Shrouded Fable', 'display', 'Fable Nimbee', 'sv-shrouded-fable', 159.99),

-- Stellar Crown
('Display Couronne Stellaire', 'display', 'Couronne Stellaire', 'sv-stellar-crown', 159.99),
('Display Stellar Crown', 'display', 'Couronne Stellaire', 'sv-stellar-crown', 159.99),

-- Surging Sparks
('Display Etincelles Deferlantes', 'display', 'Etincelles Deferlantes', 'sv-surging-sparks', 159.99),
('Display Surging Sparks', 'display', 'Etincelles Deferlantes', 'sv-surging-sparks', 159.99),

-- Prismatic Evolutions (2025)
('Display Evolutions Prismatiques', 'display', 'Evolutions Prismatiques', 'sv-prismatic-evo', 159.99),
('Display Prismatic Evolutions', 'display', 'Evolutions Prismatiques', 'sv-prismatic-evo', 159.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- ETB (Elite Trainer Boxes)
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
-- SV 151
('ETB Ecarlate et Violet 151', 'etb', 'Ecarlate et Violet 151', 'sv-151', 54.99),
('ETB SV 151', 'etb', 'Ecarlate et Violet 151', 'sv-151', 54.99),
('ETB Pokemon 151', 'etb', 'Ecarlate et Violet 151', 'sv-151', 54.99),
('Elite Trainer Box 151', 'etb', 'Ecarlate et Violet 151', 'sv-151', 54.99),

-- Paldean Fates
('ETB Destinees de Paldea', 'etb', 'Destinees de Paldea', 'sv-paldean-fates', 54.99),
('ETB Paldean Fates', 'etb', 'Destinees de Paldea', 'sv-paldean-fates', 54.99),

-- Temporal Forces
('ETB Forces Temporelles', 'etb', 'Forces Temporelles', 'sv-temporal-forces', 54.99),
('ETB Temporal Forces', 'etb', 'Forces Temporelles', 'sv-temporal-forces', 54.99),

-- Twilight Masquerade
('ETB Masques du Crepuscule', 'etb', 'Masques du Crepuscule', 'sv-twilight-masque', 54.99),
('ETB Twilight Masquerade', 'etb', 'Masques du Crepuscule', 'sv-twilight-masque', 54.99),

-- Stellar Crown
('ETB Couronne Stellaire', 'etb', 'Couronne Stellaire', 'sv-stellar-crown', 54.99),
('ETB Stellar Crown', 'etb', 'Couronne Stellaire', 'sv-stellar-crown', 54.99),

-- Surging Sparks
('ETB Etincelles Deferlantes', 'etb', 'Etincelles Deferlantes', 'sv-surging-sparks', 54.99),
('ETB Surging Sparks', 'etb', 'Etincelles Deferlantes', 'sv-surging-sparks', 54.99),

-- Prismatic Evolutions
('ETB Evolutions Prismatiques', 'etb', 'Evolutions Prismatiques', 'sv-prismatic-evo', 54.99),
('ETB Prismatic Evolutions', 'etb', 'Evolutions Prismatiques', 'sv-prismatic-evo', 54.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- COLLECTIONS / COFFRETS (Premium Collections)
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
-- Ultra Premium Collections
('Coffret Dracaufeu Ultra Premium', 'collection', 'Ecarlate et Violet', 'sv-charizard-upc', 119.99),
('Charizard Ultra Premium Collection', 'collection', 'Ecarlate et Violet', 'sv-charizard-upc', 119.99),
('Coffret Mew Ultra Premium', 'collection', 'Ecarlate et Violet 151', 'sv-151-upc', 119.99),
('Mew Ultra Premium Collection', 'collection', 'Ecarlate et Violet 151', 'sv-151-upc', 119.99),

-- Premium Collections
('Coffret Premium Rayquaza', 'collection', 'Ecarlate et Violet', 'sv-rayquaza', 49.99),
('Coffret Premium Pikachu', 'collection', 'Ecarlate et Violet', 'sv-pikachu', 49.99),

-- Special Collections
('Coffret Collection Speciale SV 151', 'collection', 'Ecarlate et Violet 151', 'sv-151', 39.99),
('Coffret Collection Speciale Paldean Fates', 'collection', 'Destinees de Paldea', 'sv-paldean-fates', 39.99),

-- Poster Collections
('Coffret Poster Collection', 'collection', 'Ecarlate et Violet', 'sv-base', 29.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- BUNDLES (6 Boosters)
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
('Bundle SV 151', 'bundle', 'Ecarlate et Violet 151', 'sv-151', 29.99),
('Bundle Destinees de Paldea', 'bundle', 'Destinees de Paldea', 'sv-paldean-fates', 29.99),
('Bundle Forces Temporelles', 'bundle', 'Forces Temporelles', 'sv-temporal-forces', 29.99),
('Bundle Masques du Crepuscule', 'bundle', 'Masques du Crepuscule', 'sv-twilight-masque', 29.99),
('Bundle Couronne Stellaire', 'bundle', 'Couronne Stellaire', 'sv-stellar-crown', 29.99),
('Bundle Evolutions Prismatiques', 'bundle', 'Evolutions Prismatiques', 'sv-prismatic-evo', 29.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- TINS / POKEBOX
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
('Pokebox Dracaufeu', 'tin', 'Ecarlate et Violet', 'sv-base', 24.99),
('Pokebox Pikachu', 'tin', 'Ecarlate et Violet', 'sv-base', 24.99),
('Pokebox Mewtwo', 'tin', 'Ecarlate et Violet', 'sv-base', 24.99),
('Pokebox Rayquaza', 'tin', 'Ecarlate et Violet', 'sv-base', 24.99),
('Tin SV 151', 'tin', 'Ecarlate et Violet 151', 'sv-151', 24.99),
('Pokebox Destinees de Paldea', 'tin', 'Destinees de Paldea', 'sv-paldean-fates', 24.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- BOOSTERS (Individual Packs)
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
('Booster SV 151', 'booster', 'Ecarlate et Violet 151', 'sv-151', 4.99),
('Booster Destinees de Paldea', 'booster', 'Destinees de Paldea', 'sv-paldean-fates', 4.99),
('Booster Forces Temporelles', 'booster', 'Forces Temporelles', 'sv-temporal-forces', 4.99),
('Booster Evolutions Prismatiques', 'booster', 'Evolutions Prismatiques', 'sv-prismatic-evo', 4.99),
('Booster Masques du Crepuscule', 'booster', 'Masques du Crepuscule', 'sv-twilight-masque', 4.99),
('Booster Couronne Stellaire', 'booster', 'Couronne Stellaire', 'sv-stellar-crown', 4.99),
('Booster Etincelles Deferlantes', 'booster', 'Etincelles Deferlantes', 'sv-surging-sparks', 4.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- Refresh the materialized view after seeding
SELECT refresh_product_stats();
