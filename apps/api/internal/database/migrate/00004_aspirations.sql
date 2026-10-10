-- +goose Up
CREATE TABLE aspirations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    division_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(160) NULL,
    nim VARCHAR(32) NULL,
    email VARCHAR(254) NULL,
    subject VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    submitted_at DATETIME(6) NOT NULL,
    read_at DATETIME(6) NULL,
    resolved_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_aspirations_division_status_created (division_id, status, created_at),
    KEY idx_aspirations_status_submitted (status, submitted_at),
    CONSTRAINT fk_aspirations_division FOREIGN KEY (division_id) REFERENCES divisions (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT chk_aspirations_status CHECK (status IN ('PENDING', 'READ', 'RESOLVED')),
    CONSTRAINT chk_aspirations_anonymous_identity CHECK (is_anonymous = FALSE OR (name IS NULL AND nim IS NULL AND email IS NULL))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE aspirations;
