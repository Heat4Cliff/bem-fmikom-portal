-- +goose Up
ALTER TABLE divisions
    ADD COLUMN display_order INT NOT NULL DEFAULT 0 AFTER description,
    ADD UNIQUE KEY uq_divisions_id_period (id, period_id),
    ADD KEY idx_divisions_period_order (period_id, status, display_order, name);

CREATE TABLE organization_members (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    period_id BIGINT UNSIGNED NOT NULL,
    division_id BIGINT UNSIGNED NULL,
    name VARCHAR(160) NOT NULL,
    nim VARCHAR(32) NULL,
    position VARCHAR(32) NOT NULL,
    photo_url VARCHAR(2048) NULL,
    display_order INT NOT NULL DEFAULT 0,
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    active_head_division_id BIGINT UNSIGNED GENERATED ALWAYS AS (
        CASE WHEN position = 'KEPALA_BIDANG' AND status = 'ACTIVE' THEN division_id ELSE NULL END
    ) STORED,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_org_members_one_active_head (period_id, active_head_division_id),
    KEY idx_org_members_period_status_order (period_id, status, display_order, name),
    KEY idx_org_members_division_status (division_id, status, display_order),
    KEY idx_org_members_period_position_status (period_id, position, status),
    CONSTRAINT fk_org_members_period FOREIGN KEY (period_id) REFERENCES periods (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT fk_org_members_division_period FOREIGN KEY (division_id, period_id) REFERENCES divisions (id, period_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT chk_org_members_status CHECK (status IN ('ACTIVE', 'INACTIVE')),
    CONSTRAINT chk_org_members_position CHECK (position IN ('KETUA', 'WAKIL_KETUA', 'SEKRETARIS', 'BENDAHARA', 'KEPALA_BIDANG', 'ANGGOTA')),
    CONSTRAINT chk_org_members_head_division CHECK (position <> 'KEPALA_BIDANG' OR division_id IS NOT NULL)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE organization_members;
ALTER TABLE divisions
    DROP KEY idx_divisions_period_order,
    DROP KEY uq_divisions_id_period,
    DROP COLUMN display_order;
