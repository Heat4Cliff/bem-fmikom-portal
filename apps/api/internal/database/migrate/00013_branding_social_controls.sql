-- +goose Up
ALTER TABLE landing_settings
    ADD COLUMN brand_logo_shape VARCHAR(16) NOT NULL DEFAULT 'original' AFTER navbar_logo_url,
    ADD COLUMN brand_logo_fit VARCHAR(16) NOT NULL DEFAULT 'contain' AFTER brand_logo_shape,
    ADD COLUMN footer_use_brand_logo TINYINT(1) NOT NULL DEFAULT 1 AFTER footer_logo_url,
    ADD COLUMN footer_logo_shape VARCHAR(16) NOT NULL DEFAULT 'original' AFTER footer_use_brand_logo,
    ADD COLUMN footer_logo_fit VARCHAR(16) NOT NULL DEFAULT 'contain' AFTER footer_logo_shape,
    ADD COLUMN cabinet_logo_shape VARCHAR(16) NOT NULL DEFAULT 'original' AFTER cabinet_logo_url,
    ADD COLUMN cabinet_logo_fit VARCHAR(16) NOT NULL DEFAULT 'contain' AFTER cabinet_logo_shape,
    ADD COLUMN social_links_json JSON NULL AFTER other_social_url;

-- Preserve a previously configured footer override when upgrading an existing installation.
UPDATE landing_settings SET footer_use_brand_logo = 0 WHERE footer_logo_url IS NOT NULL;

-- +goose Down
ALTER TABLE landing_settings
    DROP COLUMN social_links_json,
    DROP COLUMN cabinet_logo_fit,
    DROP COLUMN cabinet_logo_shape,
    DROP COLUMN footer_logo_fit,
    DROP COLUMN footer_logo_shape,
    DROP COLUMN footer_use_brand_logo,
    DROP COLUMN brand_logo_fit,
    DROP COLUMN brand_logo_shape,
    ALGORITHM=COPY;
