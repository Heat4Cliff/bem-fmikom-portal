-- +goose Up
CREATE TABLE landing_settings (
    id TINYINT UNSIGNED NOT NULL,
    hero_headline VARCHAR(180) NOT NULL,
    hero_subheadline VARCHAR(500) NOT NULL,
    hero_primary_cta_label VARCHAR(80) NOT NULL,
    hero_primary_cta_url VARCHAR(500) NOT NULL,
    hero_secondary_cta_label VARCHAR(80) NULL,
    hero_secondary_cta_url VARCHAR(500) NULL,
    hero_image_url VARCHAR(2048) NULL,
    about_title VARCHAR(180) NOT NULL,
    about_description TEXT NOT NULL,
    vision TEXT NOT NULL,
    missions_json JSON NOT NULL,
    contact_email VARCHAR(254) NULL,
    contact_phone VARCHAR(64) NULL,
    contact_address VARCHAR(500) NULL,
    location_url VARCHAR(2048) NULL,
    instagram_url VARCHAR(2048) NULL,
    youtube_url VARCHAR(2048) NULL,
    tiktok_url VARCHAR(2048) NULL,
    other_social_url VARCHAR(2048) NULL,
    updated_by BIGINT UNSIGNED NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT chk_landing_settings_singleton CHECK (id = 1),
    CONSTRAINT fk_landing_settings_updated_by FOREIGN KEY (updated_by) REFERENCES users (id) ON UPDATE RESTRICT ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO landing_settings (
    id, hero_headline, hero_subheadline, hero_primary_cta_label, hero_primary_cta_url,
    about_title, about_description, vision, missions_json
) VALUES (
    1,
    'Sistem Informasi & Wadah Sinergi Mahasiswa FMIKOM UNUGHA',
    'Temukan informasi organisasi, kegiatan, pengumuman, dan layanan BEM FMIKOM UNUGHA melalui satu portal.',
    'Jelajahi informasi', '#berita',
    'Ruang informasi dan kolaborasi mahasiswa',
    'Portal ini menjadi kanal publik untuk mengenal BEM FMIKOM UNUGHA dan menemukan informasi organisasi. Konten resmi dapat diperbarui ketika pengelolaan konten tersedia.',
    '', JSON_ARRAY()
);

-- +goose Down
DROP TABLE landing_settings;
