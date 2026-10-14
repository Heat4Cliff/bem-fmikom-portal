-- +goose Up
ALTER TABLE divisions
    ADD COLUMN show_in_public_structure BOOLEAN NOT NULL DEFAULT TRUE AFTER display_order,
    ADD KEY idx_divisions_public_structure (period_id, status, show_in_public_structure, display_order);

-- Existing canonical target/leadership groups remain in divisions for other
-- workflows, but are explicitly removed from the public work-division chart.
-- Runtime presentation uses this metadata and never tests division names.
UPDATE divisions
SET show_in_public_structure = FALSE
WHERE LOWER(TRIM(name)) IN ('fakultas', 'bph');

-- +goose Down
ALTER TABLE divisions
    DROP KEY idx_divisions_public_structure,
    DROP COLUMN show_in_public_structure;
