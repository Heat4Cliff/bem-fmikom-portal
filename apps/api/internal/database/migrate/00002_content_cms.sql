-- +goose Up
CREATE TABLE article_categories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(160) NOT NULL,
    description TEXT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_at DATETIME(6) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_article_categories_slug (slug),
    KEY idx_article_categories_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE articles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    title VARCHAR(240) NOT NULL,
    slug VARCHAR(240) NOT NULL,
    excerpt VARCHAR(500) NULL,
    content_html MEDIUMTEXT NOT NULL,
    thumbnail_url VARCHAR(2048) NULL,
    division_id BIGINT UNSIGNED NULL,
    author_id BIGINT UNSIGNED NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'DRAFT',
    published_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_at DATETIME(6) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_articles_slug (slug),
    KEY idx_articles_status_published (status, published_at),
    KEY idx_articles_division_status (division_id, status, deleted_at),
    KEY idx_articles_author (author_id),
    CONSTRAINT fk_articles_division FOREIGN KEY (division_id) REFERENCES divisions (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT fk_articles_author FOREIGN KEY (author_id) REFERENCES users (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT chk_articles_status CHECK (status IN ('DRAFT', 'PUBLISHED')),
    CONSTRAINT chk_articles_publish_state CHECK ((status = 'DRAFT' AND published_at IS NULL) OR (status = 'PUBLISHED' AND published_at IS NOT NULL))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE article_category_map (
    article_id BIGINT UNSIGNED NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (article_id, category_id),
    KEY idx_article_category_map_category (category_id, article_id),
    CONSTRAINT fk_article_category_map_article FOREIGN KEY (article_id) REFERENCES articles (id) ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT fk_article_category_map_category FOREIGN KEY (category_id) REFERENCES article_categories (id) ON UPDATE RESTRICT ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE gallery_categories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    title VARCHAR(200) NOT NULL,
    slug VARCHAR(220) NOT NULL,
    description TEXT NULL,
    period_id BIGINT UNSIGNED NULL,
    division_id BIGINT UNSIGNED NULL,
    event_date DATE NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_at DATETIME(6) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_gallery_categories_slug (slug),
    KEY idx_gallery_categories_scope (division_id, period_id, deleted_at),
    CONSTRAINT fk_gallery_categories_period FOREIGN KEY (period_id) REFERENCES periods (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT fk_gallery_categories_division FOREIGN KEY (division_id) REFERENCES divisions (id) ON UPDATE RESTRICT ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE gallery_items (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    category_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT NULL,
    google_photos_url VARCHAR(2048) NOT NULL,
    preview_image_url VARCHAR(2048) NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_at DATETIME(6) NULL,
    PRIMARY KEY (id),
    KEY idx_gallery_items_category_order (category_id, sort_order, id),
    KEY idx_gallery_items_deleted (deleted_at),
    CONSTRAINT fk_gallery_items_category FOREIGN KEY (category_id) REFERENCES gallery_categories (id) ON UPDATE RESTRICT ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE announcements (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    title VARCHAR(240) NOT NULL,
    slug VARCHAR(240) NOT NULL,
    content_html MEDIUMTEXT NOT NULL,
    image_url VARCHAR(2048) NULL,
    start_at DATETIME(6) NOT NULL,
    end_at DATETIME(6) NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_at DATETIME(6) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_announcements_slug (slug),
    KEY idx_announcements_public_window (is_active, start_at, end_at, deleted_at),
    KEY idx_announcements_created_by (created_by),
    CONSTRAINT fk_announcements_creator FOREIGN KEY (created_by) REFERENCES users (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT chk_announcements_window CHECK (end_at IS NULL OR start_at <= end_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE announcements;
DROP TABLE gallery_items;
DROP TABLE gallery_categories;
DROP TABLE article_category_map;
DROP TABLE articles;
DROP TABLE article_categories;
