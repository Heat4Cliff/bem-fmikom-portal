-- +goose Up
ALTER TABLE landing_settings
    ADD COLUMN navbar_logo_url TEXT NULL AFTER hero_image_url,
    ADD COLUMN footer_logo_url TEXT NULL AFTER navbar_logo_url,
    ADD COLUMN cabinet_logo_url TEXT NULL AFTER footer_logo_url,
    ADD COLUMN cabinet_name VARCHAR(180) NULL AFTER cabinet_logo_url,
    ADD COLUMN cabinet_period VARCHAR(80) NULL AFTER cabinet_name,
    ADD COLUMN cabinet_tagline VARCHAR(180) NULL AFTER cabinet_period;

ALTER TABLE media_assets
    DROP CONSTRAINT chk_media_assets_purpose,
    ADD CONSTRAINT chk_media_assets_purpose CHECK (purpose IN ('ARTICLE_COVER','ANNOUNCEMENT_IMAGE','GALLERY_PREVIEW','MEMBER_PHOTO','LANDING_HERO','LANDING_BRANDING'));

-- +goose Down
ALTER TABLE media_assets
    DROP CONSTRAINT chk_media_assets_purpose,
    ADD CONSTRAINT chk_media_assets_purpose CHECK (purpose IN ('ARTICLE_COVER','ANNOUNCEMENT_IMAGE','GALLERY_PREVIEW','MEMBER_PHOTO','LANDING_HERO'));

ALTER TABLE landing_settings
    DROP COLUMN cabinet_tagline,
    DROP COLUMN cabinet_period,
    DROP COLUMN cabinet_name,
    DROP COLUMN cabinet_logo_url,
    DROP COLUMN footer_logo_url,
    DROP COLUMN navbar_logo_url;
