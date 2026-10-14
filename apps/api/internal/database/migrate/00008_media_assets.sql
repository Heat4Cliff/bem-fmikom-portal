-- +goose Up
CREATE TABLE media_assets (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    uploaded_by BIGINT UNSIGNED NULL,
    purpose VARCHAR(32) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    storage_key VARCHAR(96) NOT NULL,
    mime_type VARCHAR(64) NOT NULL,
    file_extension VARCHAR(8) NOT NULL,
    size_bytes BIGINT UNSIGNED NOT NULL,
    width INT UNSIGNED NOT NULL,
    height INT UNSIGNED NOT NULL,
    sha256 CHAR(64) NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_media_assets_storage_key (storage_key),
    KEY idx_media_assets_uploader_purpose (uploaded_by, purpose, created_at),
    KEY idx_media_assets_sha256 (sha256),
    CONSTRAINT fk_media_assets_uploader FOREIGN KEY (uploaded_by) REFERENCES users (id) ON UPDATE RESTRICT ON DELETE SET NULL,
    CONSTRAINT chk_media_assets_purpose CHECK (purpose IN ('ARTICLE_COVER','ANNOUNCEMENT_IMAGE','GALLERY_PREVIEW','MEMBER_PHOTO','LANDING_HERO')),
    CONSTRAINT chk_media_assets_type CHECK (mime_type IN ('image/jpeg','image/png')),
    CONSTRAINT chk_media_assets_extension CHECK (file_extension IN ('.jpg','.jpeg','.png')),
    CONSTRAINT chk_media_assets_size CHECK (size_bytes > 0 AND width > 0 AND height > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE media_assets;
