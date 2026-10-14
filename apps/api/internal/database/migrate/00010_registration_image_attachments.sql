-- +goose Up
ALTER TABLE registration_fields
    DROP CONSTRAINT chk_registration_fields_type,
    ADD CONSTRAINT chk_registration_fields_type CHECK (field_type IN ('TEXT', 'TEXTAREA', 'EMAIL', 'NUMBER', 'SELECT', 'RADIO', 'CHECKBOX', 'DATE', 'IMAGE'));

CREATE TABLE registration_submission_files (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    submission_id BIGINT UNSIGNED NOT NULL,
    field_id BIGINT UNSIGNED NULL,
    field_key_snapshot VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    storage_key VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    original_mime_type VARCHAR(64) NOT NULL,
    stored_mime_type VARCHAR(64) NOT NULL,
    size_bytes BIGINT UNSIGNED NOT NULL,
    width INT UNSIGNED NOT NULL,
    height INT UNSIGNED NOT NULL,
    sha256 CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_registration_submission_files_key (submission_id, field_key_snapshot),
    UNIQUE KEY uq_registration_submission_files_storage (storage_key),
    KEY idx_registration_submission_files_field (field_id),
    CONSTRAINT fk_registration_submission_files_submission FOREIGN KEY (submission_id) REFERENCES registration_submissions (id) ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT fk_registration_submission_files_field FOREIGN KEY (field_id) REFERENCES registration_fields (id) ON UPDATE RESTRICT ON DELETE SET NULL,
    CONSTRAINT chk_registration_submission_files_size CHECK (size_bytes > 0),
    CONSTRAINT chk_registration_submission_files_dimensions CHECK (width > 0 AND height > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS registration_submission_files;
-- IMAGE form configuration is no longer valid after this migration is reverted.
DELETE FROM registration_fields WHERE field_type = 'IMAGE';
ALTER TABLE registration_fields
    DROP CONSTRAINT chk_registration_fields_type,
    ADD CONSTRAINT chk_registration_fields_type CHECK (field_type IN ('TEXT', 'TEXTAREA', 'EMAIL', 'NUMBER', 'SELECT', 'RADIO', 'CHECKBOX', 'DATE'));
