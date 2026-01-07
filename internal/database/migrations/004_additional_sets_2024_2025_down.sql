-- Migration: 004_additional_sets_2024_2025 (down)
-- Purpose: Remove seed data for additional Pokemon TCG sets (2024-2025)
-- Date: 2026-01-07

-- Delete the newly added products
DELETE FROM products WHERE set_code IN (
    'sv-journey-together',
    'sv-destined-rivals',
    'sv-terastal-fest',
    'sv-holiday-2024',
    'sv-advent-2024',
    'sv-terastal-upc',
    'sv-pc-exclusive',
    'sv-trainer-gallery',
    'sv-pikachu-2025',
    'sv-starters',
    'sv-2025',
    'sv-terastal',
    'swsh-crown-zenith',
    'swsh-pokemon-go',
    'swsh-celebrations'
);

-- Refresh the materialized view after deletion
SELECT refresh_product_stats();
