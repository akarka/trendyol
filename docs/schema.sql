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
