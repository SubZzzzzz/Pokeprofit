-- Migration: 004_additional_sets_2024_2025
-- Purpose: Add seed data for additional Pokemon TCG sets (2024-2025)
-- Date: 2026-01-07
-- Phase 6: Polish & Cross-Cutting Concerns

-- ============================================
-- 2025 SETS - DISPLAYS
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
-- Journey Together (January 2025)
('Display Voyage Ensemble', 'display', 'Voyage Ensemble', 'sv-journey-together', 159.99),
('Display Journey Together', 'display', 'Voyage Ensemble', 'sv-journey-together', 159.99),

-- Prismatic Evolutions - Additional variants
('Display Evolutions Prismatiques Mini Booster', 'display', 'Evolutions Prismatiques', 'sv-prismatic-evo', 129.99),

-- Surging Sparks - Additional variants (Late 2024)
('Display Etincelles Deferlantes Expansion', 'display', 'Etincelles Deferlantes', 'sv-surging-sparks', 159.99),

-- Destined Rivals (Expected Q2 2025)
('Display Rivaux Destines', 'display', 'Rivaux Destines', 'sv-destined-rivals', 159.99),
('Display Destined Rivals', 'display', 'Rivaux Destines', 'sv-destined-rivals', 159.99),

-- Terastal Festival (Expected Q3 2025)
('Display Festival Terastal', 'display', 'Festival Terastal', 'sv-terastal-fest', 159.99),
('Display Terastal Festival', 'display', 'Festival Terastal', 'sv-terastal-fest', 159.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- 2025 SETS - ETBs
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
-- Journey Together
('ETB Voyage Ensemble', 'etb', 'Voyage Ensemble', 'sv-journey-together', 54.99),
('ETB Journey Together', 'etb', 'Voyage Ensemble', 'sv-journey-together', 54.99),

-- Destined Rivals
('ETB Rivaux Destines', 'etb', 'Rivaux Destines', 'sv-destined-rivals', 54.99),
('ETB Destined Rivals', 'etb', 'Rivaux Destines', 'sv-destined-rivals', 54.99),

-- Terastal Festival
('ETB Festival Terastal', 'etb', 'Festival Terastal', 'sv-terastal-fest', 54.99),
('ETB Terastal Festival', 'etb', 'Festival Terastal', 'sv-terastal-fest', 54.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- 2024-2025 SPECIAL COLLECTIONS
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
-- 2024 Holiday Collections
('Coffret Noel 2024 Pikachu', 'collection', 'Collection Noel 2024', 'sv-holiday-2024', 49.99),
('Coffret Calendrier Avent 2024', 'collection', 'Collection Noel 2024', 'sv-advent-2024', 79.99),
('Advent Calendar Pokemon 2024', 'collection', 'Collection Noel 2024', 'sv-advent-2024', 79.99),

-- Ultra Premium Collections 2024-2025
('Coffret Terastal Ultra Premium', 'collection', 'Collection Ultra Premium', 'sv-terastal-upc', 119.99),
('Terastal Ultra Premium Collection', 'collection', 'Collection Ultra Premium', 'sv-terastal-upc', 119.99),

-- Pokemon Center Exclusive Collections
('Coffret Pokemon Center Exclusive 2025', 'collection', 'Pokemon Center Exclusive', 'sv-pc-exclusive', 89.99),

-- Trainer Gallery Collections
('Coffret Galerie Dresseurs SV', 'collection', 'Galerie Dresseurs', 'sv-trainer-gallery', 39.99),

-- Special Illustration Collections
('Coffret Illustrations Speciales Prismatic', 'collection', 'Evolutions Prismatiques', 'sv-prismatic-evo', 69.99),

-- Pikachu Collections
('Coffret Pikachu Ex 2025', 'collection', 'Collection Pikachu', 'sv-pikachu-2025', 34.99),
('Pikachu Ex Premium Collection', 'collection', 'Collection Pikachu', 'sv-pikachu-2025', 59.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- 2025 BUNDLES
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
('Bundle Voyage Ensemble', 'bundle', 'Voyage Ensemble', 'sv-journey-together', 29.99),
('Bundle Journey Together', 'bundle', 'Voyage Ensemble', 'sv-journey-together', 29.99),
('Bundle Rivaux Destines', 'bundle', 'Rivaux Destines', 'sv-destined-rivals', 29.99),
('Bundle Festival Terastal', 'bundle', 'Festival Terastal', 'sv-terastal-fest', 29.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- 2025 TINS & POKEBOX
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
-- Starter Tins 2025
('Pokebox Starters Paldea 2025', 'tin', 'Ecarlate et Violet', 'sv-starters', 24.99),
('Pokebox Pohmarmotte', 'tin', 'Evolutions Prismatiques', 'sv-prismatic-evo', 24.99),
('Pokebox Pikachu 2025', 'tin', 'Collection 2025', 'sv-2025', 24.99),
('Pokebox Raichu', 'tin', 'Collection 2025', 'sv-2025', 24.99),

-- EX Tins
('Tin Dracaufeu Ex Terastal', 'tin', 'Collection Terastal', 'sv-terastal', 24.99),
('Tin Mewtwo Ex Terastal', 'tin', 'Collection Terastal', 'sv-terastal', 24.99),
('Tin Rayquaza Ex Terastal', 'tin', 'Collection Terastal', 'sv-terastal', 24.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- 2025 BOOSTERS
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
('Booster Voyage Ensemble', 'booster', 'Voyage Ensemble', 'sv-journey-together', 4.99),
('Booster Journey Together', 'booster', 'Voyage Ensemble', 'sv-journey-together', 4.99),
('Booster Rivaux Destines', 'booster', 'Rivaux Destines', 'sv-destined-rivals', 4.99),
('Booster Festival Terastal', 'booster', 'Festival Terastal', 'sv-terastal-fest', 4.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- OLDER SETS - HIGH VALUE (2023 reference)
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
-- Crown Zenith (2023 - High collector value)
('Display Zenith Supreme', 'display', 'Zenith Supreme', 'swsh-crown-zenith', 179.99),
('Display Crown Zenith', 'display', 'Zenith Supreme', 'swsh-crown-zenith', 179.99),
('ETB Zenith Supreme', 'etb', 'Zenith Supreme', 'swsh-crown-zenith', 59.99),
('ETB Crown Zenith', 'etb', 'Zenith Supreme', 'swsh-crown-zenith', 59.99),

-- Pokemon GO (2022 - Still trading)
('Display Pokemon GO', 'display', 'Pokemon GO', 'swsh-pokemon-go', 159.99),
('ETB Pokemon GO', 'etb', 'Pokemon GO', 'swsh-pokemon-go', 49.99),

-- Celebrations (25th Anniversary - High collector value)
('Coffret Ultra Premium 25 Anniversaire', 'collection', 'Celebrations', 'swsh-celebrations', 149.99),
('Ultra Premium Collection 25th Anniversary', 'collection', 'Celebrations', 'swsh-celebrations', 149.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- ============================================
-- SPECIAL PRODUCTS - BLISTERS & MISC
-- ============================================
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
-- 3-Pack Blisters (categorized as bundles)
('Tripack Evolutions Prismatiques', 'bundle', 'Evolutions Prismatiques', 'sv-prismatic-evo', 14.99),
('Tripack Prismatic Evolutions', 'bundle', 'Evolutions Prismatiques', 'sv-prismatic-evo', 14.99),
('Tripack SV 151', 'bundle', 'Ecarlate et Violet 151', 'sv-151', 14.99),
('Tripack Destinees Paldea', 'bundle', 'Destinees de Paldea', 'sv-paldean-fates', 14.99),

-- Build & Battle Boxes
('Build Battle Evolutions Prismatiques', 'collection', 'Evolutions Prismatiques', 'sv-prismatic-evo', 29.99),
('Build Battle Etincelles Deferlantes', 'collection', 'Etincelles Deferlantes', 'sv-surging-sparks', 29.99),
('Build Battle Couronne Stellaire', 'collection', 'Couronne Stellaire', 'sv-stellar-crown', 29.99)

ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    set_name = EXCLUDED.set_name,
    set_code = EXCLUDED.set_code,
    updated_at = NOW();

-- Refresh the materialized view after seeding
SELECT refresh_product_stats();
