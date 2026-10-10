-- +goose Up
CREATE TABLE registration_forms (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    period_id BIGINT UNSIGNED NOT NULL,
    division_id BIGINT UNSIGNED NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    title VARCHAR(200) NOT NULL,
    slug VARCHAR(191) NOT NULL,
    description TEXT NULL,
    open_at DATETIME(6) NOT NULL,
    close_at DATETIME(6) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'DRAFT',
    deleted_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_registration_forms_slug (slug),
    KEY idx_registration_forms_period_status_window (period_id, status, open_at, close_at),
    KEY idx_registration_forms_division_status_window (division_id, status, open_at, close_at),
    KEY idx_registration_forms_period_division (period_id, division_id),
    CONSTRAINT fk_registration_forms_period FOREIGN KEY (period_id) REFERENCES periods (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT fk_registration_forms_division_period FOREIGN KEY (division_id, period_id) REFERENCES divisions (id, period_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT fk_registration_forms_creator FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT chk_registration_forms_status CHECK (status IN ('DRAFT', 'OPEN', 'CLOSED')),
    CONSTRAINT chk_registration_forms_date_range CHECK (open_at < close_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE registration_fields (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    registration_form_id BIGINT UNSIGNED NOT NULL,
    field_key VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    label VARCHAR(160) NOT NULL,
    field_type VARCHAR(16) NOT NULL,
    placeholder VARCHAR(255) NULL,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INT NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_registration_fields_form_key (registration_form_id, field_key),
    UNIQUE KEY uq_registration_fields_id_form (id, registration_form_id),
    KEY idx_registration_fields_form_order (registration_form_id, display_order, id),
    CONSTRAINT fk_registration_fields_form FOREIGN KEY (registration_form_id) REFERENCES registration_forms (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT chk_registration_fields_type CHECK (field_type IN ('TEXT', 'TEXTAREA', 'EMAIL', 'NUMBER', 'SELECT', 'RADIO', 'CHECKBOX', 'DATE')),
    CONSTRAINT chk_registration_fields_order CHECK (display_order > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE registration_field_options (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    field_id BIGINT UNSIGNED NOT NULL,
    value VARCHAR(160) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    label VARCHAR(160) NOT NULL,
    display_order INT NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_registration_field_options_value (field_id, value),
    KEY idx_registration_field_options_order (field_id, display_order, id),
    CONSTRAINT fk_registration_field_options_field FOREIGN KEY (field_id) REFERENCES registration_fields (id) ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT chk_registration_field_options_order CHECK (display_order > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE registration_submissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    registration_form_id BIGINT UNSIGNED NOT NULL,
    full_name VARCHAR(160) NOT NULL,
    nim VARCHAR(32) NOT NULL,
    normalized_nim VARCHAR(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    email VARCHAR(254) NOT NULL,
    phone VARCHAR(32) NOT NULL,
    submitted_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    ip_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_registration_submissions_form_nim (registration_form_id, normalized_nim),
    KEY idx_registration_submissions_form_submitted (registration_form_id, submitted_at, id),
    CONSTRAINT fk_registration_submissions_form FOREIGN KEY (registration_form_id) REFERENCES registration_forms (id) ON UPDATE RESTRICT ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE registration_submission_answers (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    submission_id BIGINT UNSIGNED NOT NULL,
    field_id BIGINT UNSIGNED NULL,
    field_key_snapshot VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    label_snapshot VARCHAR(160) NOT NULL,
    type_snapshot VARCHAR(16) NOT NULL,
    value_text LONGTEXT NULL,
    value_json JSON NULL,
    display_order_snapshot INT NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_registration_submission_answers_key (submission_id, field_key_snapshot),
    KEY idx_registration_submission_answers_field (field_id),
    KEY idx_registration_submission_answers_order (submission_id, display_order_snapshot, id),
    CONSTRAINT fk_registration_submission_answers_submission FOREIGN KEY (submission_id) REFERENCES registration_submissions (id) ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT fk_registration_submission_answers_field FOREIGN KEY (field_id) REFERENCES registration_fields (id) ON UPDATE RESTRICT ON DELETE SET NULL,
    CONSTRAINT chk_registration_submission_answers_value CHECK (value_text IS NULL OR value_json IS NULL)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE registration_submission_answers;
DROP TABLE registration_submissions;
DROP TABLE registration_field_options;
DROP TABLE registration_fields;
DROP TABLE registration_forms;
