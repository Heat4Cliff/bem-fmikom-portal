-- +goose Up
CREATE TABLE documents (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    period_id BIGINT UNSIGNED NOT NULL,
    division_id BIGINT UNSIGNED NOT NULL,
    uploaded_by BIGINT UNSIGNED NOT NULL,
    name VARCHAR(240) NOT NULL,
    category VARCHAR(24) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    storage_key VARCHAR(160) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    mime_type VARCHAR(127) NOT NULL,
    file_extension VARCHAR(10) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    size_bytes BIGINT UNSIGNED NOT NULL,
    sha256 CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    delete_pending BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_documents_storage_key (storage_key),
    KEY idx_documents_period_division_category_created (period_id, division_id, category, created_at),
    KEY idx_documents_division_created (division_id, created_at),
    KEY idx_documents_extension_created (file_extension, created_at),
    KEY idx_documents_uploader (uploaded_by),
    KEY idx_documents_delete_pending (delete_pending, updated_at),
    CONSTRAINT fk_documents_period FOREIGN KEY (period_id) REFERENCES periods (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT fk_documents_division_period FOREIGN KEY (division_id, period_id) REFERENCES divisions (id, period_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT fk_documents_uploaded_by FOREIGN KEY (uploaded_by) REFERENCES users (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT chk_documents_name CHECK (CHAR_LENGTH(TRIM(name)) BETWEEN 1 AND 240),
    CONSTRAINT chk_documents_category CHECK (category IN ('SK', 'PROPOSAL', 'LPJ', 'OTHER')),
    CONSTRAINT chk_documents_extension CHECK (file_extension IN ('.pdf', '.doc', '.docx', '.xls', '.xlsx')),
    CONSTRAINT chk_documents_size CHECK (size_bytes > 0),
    CONSTRAINT chk_documents_sha256 CHECK (sha256 REGEXP '^[0-9a-f]{64}$')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE documents;
