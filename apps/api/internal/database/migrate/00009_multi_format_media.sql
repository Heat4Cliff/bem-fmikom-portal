-- +goose Up
ALTER TABLE media_assets
    ADD COLUMN original_mime_type VARCHAR(64) NULL AFTER original_filename;

UPDATE media_assets SET original_mime_type = mime_type WHERE original_mime_type IS NULL;

ALTER TABLE media_assets
    MODIFY COLUMN original_mime_type VARCHAR(64) NOT NULL,
    DROP CONSTRAINT chk_media_assets_type,
    DROP CONSTRAINT chk_media_assets_extension,
    ADD CONSTRAINT chk_media_assets_type CHECK (mime_type IN ('image/jpeg','image/png','image/gif','image/webp')),
    ADD CONSTRAINT chk_media_assets_extension CHECK (file_extension IN ('.jpg','.jpeg','.jfif','.png','.webp','.gif')),
    ADD CONSTRAINT chk_media_assets_original_type CHECK (original_mime_type IN ('image/jpeg','image/png','image/gif','image/webp','image/tiff','image/bmp'));

-- +goose Down
ALTER TABLE media_assets
    DROP CONSTRAINT chk_media_assets_original_type,
    DROP CONSTRAINT chk_media_assets_type,
    DROP CONSTRAINT chk_media_assets_extension,
    ADD CONSTRAINT chk_media_assets_type CHECK (mime_type IN ('image/jpeg','image/png')),
    ADD CONSTRAINT chk_media_assets_extension CHECK (file_extension IN ('.jpg','.jpeg','.png'));

ALTER TABLE media_assets DROP COLUMN original_mime_type;
