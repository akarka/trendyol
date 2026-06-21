CREATE TABLE IF NOT EXISTS trendyol_orders (
  uuid           CHAR(36)     PRIMARY KEY DEFAULT (UUID()),
  order_id       VARCHAR(255) NOT NULL,
  order_number   VARCHAR(255) NOT NULL,
  package_status VARCHAR(50)  NOT NULL,
  payload        JSON         NOT NULL,
  created_at     DATETIME     DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_order_status (order_id, package_status)
);

CREATE TABLE IF NOT EXISTS users (
  id            INT          AUTO_INCREMENT PRIMARY KEY,
  username      VARCHAR(100) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role          VARCHAR(50)  NOT NULL DEFAULT 'admin',
  created_at    DATETIME     DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS print_jobs (
  id           INT          AUTO_INCREMENT PRIMARY KEY,
  order_id     VARCHAR(255) NOT NULL,
  status       VARCHAR(50)  NOT NULL,
  error_msg    TEXT,
  attempted_at DATETIME     DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_order_id (order_id)
);

CREATE TABLE IF NOT EXISTS settings (
  `key`      VARCHAR(100) PRIMARY KEY,
  value      TEXT         NOT NULL,
  updated_at DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Ürün kataloğu (normalize). Kaynak: Zeytuni_Ops CSV → cmd/import-products.
-- Komisyon/KDV-dahil türevleri saklanmaz; price + vat_rate'ten hesaplanabilir.
CREATE TABLE IF NOT EXISTS categories (
  id   INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS brands (
  id   INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS products (
  sku              VARCHAR(40)   PRIMARY KEY,
  barcode          VARCHAR(40)   NOT NULL,
  name             VARCHAR(255)  NOT NULL,   -- iç ad (CSV 'Ürün')
  marketplace_name VARCHAR(255),             -- 'Ty Ürün Adı'
  category_id      INT           NOT NULL,
  brand_id         INT,
  net_weight       DECIMAL(10,3),            -- NULL = gramaj eksik (needs_fix)
  unit             VARCHAR(8),               -- g|kg|ml|l|adet|kase
  price            DECIMAL(10,2) NOT NULL,   -- 'Fiziki Mağaza Fiyatı'
  vat_rate         DECIMAL(5,2),             -- KDV % (ürüne göre değişir, türetilemez)
  is_active        BOOLEAN       NOT NULL DEFAULT 0,
  needs_fix        BOOLEAN       NOT NULL DEFAULT 0,  -- import'ta tespit edilen veri sorunu
  description      TEXT,
  created_at       DATETIME      DEFAULT CURRENT_TIMESTAMP,
  updated_at       DATETIME      DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_barcode (barcode),
  CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories(id),
  CONSTRAINT fk_products_brand    FOREIGN KEY (brand_id)    REFERENCES brands(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- İlk admin; gerçek şifre için: go run ./cmd/seed --username admin --password <pass>
INSERT IGNORE INTO users (username, password_hash, role)
VALUES ('admin', '$2a$10$PLACEHOLDER', 'admin');
