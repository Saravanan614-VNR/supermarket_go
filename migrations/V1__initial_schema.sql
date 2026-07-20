-- ============================================================================
-- SupermarketService - Initial schema + mock data
-- Matches the GORM entity definitions in entities/*.go exactly (column names,
-- types, uniqueness, and foreign keys). Per config/database.go, AutoMigrate()
-- is intentionally NOT used in production - tables must exist before boot.
--
-- Usage:
--   mysql -u root -p < migrations/V1__initial_schema.sql
--
-- Make sure MYSQL_DSN in your .env points at the super-market database
-- created below (or rename the DB below to match your .env). The name
-- contains a hyphen, so it must stay backtick-quoted everywhere it is used;
-- left unquoted, MySQL parses super-market as super MINUS market.
-- ============================================================================

-- Drops any partial/previous state so this script is safe to re-run from
-- scratch (this is a dev seed script, not an incremental production migration).
DROP DATABASE IF EXISTS `super-market`;

CREATE DATABASE `super-market`
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE `super-market`;

SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------------------------------------------------------
-- Table: _user  (entities/user.go)
-- ----------------------------------------------------------------------------
DROP TABLE IF EXISTS `_user`;
CREATE TABLE `_user` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `full_name`  VARCHAR(255)    NOT NULL,
    `username`   VARCHAR(255)    NOT NULL,
    `password`   VARCHAR(255)    NOT NULL,
    `role`       VARCHAR(50)     NOT NULL, -- ADMIN, CASHIER, INVENTORY
    `deleted_at` TIMESTAMP       NULL DEFAULT NULL,
    `created_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_user_username_deleted_at` (`username`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- Table: category  (entities/category.go)
-- ----------------------------------------------------------------------------
DROP TABLE IF EXISTS `category`;
CREATE TABLE `category` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`        VARCHAR(255)    NOT NULL,
    `description` VARCHAR(255)    NOT NULL,
    `deleted_at`  TIMESTAMP       NULL DEFAULT NULL,
    `created_at`  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_category_name_deleted_at` (`name`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- Table: client  (entities/client.go)
-- ----------------------------------------------------------------------------
DROP TABLE IF EXISTS `client`;
CREATE TABLE `client` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`       VARCHAR(255)    NOT NULL,
    `dni`        VARCHAR(255)    NOT NULL,
    `email`      VARCHAR(255)    NOT NULL,
    `deleted_at` TIMESTAMP       NULL DEFAULT NULL,
    `created_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_client_dni_deleted_at` (`dni`, `deleted_at`),
    UNIQUE KEY `idx_client_email_deleted_at` (`email`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- Table: product  (entities/product.go)
-- ----------------------------------------------------------------------------
DROP TABLE IF EXISTS `product`;
CREATE TABLE `product` (
    `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`           VARCHAR(255)    NOT NULL,
    `brand`          VARCHAR(255)    NOT NULL,
    `base_price`     DOUBLE          NOT NULL,
    `tax_percentage` DOUBLE          NOT NULL,
    `category_id`    BIGINT UNSIGNED NOT NULL,
    `stock`          INT             NOT NULL,
    `deleted_at`     TIMESTAMP       NULL DEFAULT NULL,
    `created_at`     TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_product_name_deleted_at` (`name`, `deleted_at`),
    KEY `idx_product_category_id` (`category_id`),
    CONSTRAINT `fk_product_category` FOREIGN KEY (`category_id`) REFERENCES `category` (`id`)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- Table: promotion  (entities/promotion.go)
-- ----------------------------------------------------------------------------
DROP TABLE IF EXISTS `promotion`;
CREATE TABLE `promotion` (
    `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`                VARCHAR(255)    NOT NULL,
    `discount_percentage` DOUBLE          NOT NULL,
    `start_date`          DATETIME        NOT NULL,
    `end_date`            DATETIME        NOT NULL,
    `deleted_at`          TIMESTAMP       NULL DEFAULT NULL,
    `created_at`          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- Table: promotion_products  (entities/promotion.go - PromotionProduct)
-- ----------------------------------------------------------------------------
DROP TABLE IF EXISTS `promotion_products`;
CREATE TABLE `promotion_products` (
    `promotion_id` BIGINT UNSIGNED NOT NULL,
    `product_id`   BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`promotion_id`, `product_id`),
    KEY `idx_promotion_products_product` (`product_id`),
    CONSTRAINT `fk_promo_products_promotion` FOREIGN KEY (`promotion_id`) REFERENCES `promotion` (`id`)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT `fk_promo_products_product` FOREIGN KEY (`product_id`) REFERENCES `product` (`id`)
        ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- Table: sale  (entities/sale.go)
-- ----------------------------------------------------------------------------
DROP TABLE IF EXISTS `sale`;
CREATE TABLE `sale` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `total_price` DOUBLE          NOT NULL DEFAULT 0.0,
    `status`      VARCHAR(50)     NOT NULL, -- OPEN, CLOSED, CANCELED
    `client_id`   BIGINT UNSIGNED NULL,
    `cashier_id`  BIGINT UNSIGNED NOT NULL,
    `finished_at` DATETIME        NULL,
    `created_at`  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_sale_client_id` (`client_id`),
    KEY `idx_sale_cashier_id` (`cashier_id`),
    CONSTRAINT `fk_sale_client` FOREIGN KEY (`client_id`) REFERENCES `client` (`id`)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT `fk_sale_cashier` FOREIGN KEY (`cashier_id`) REFERENCES `_user` (`id`)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- Table: sale_detail  (entities/sale.go - SaleDetail)
-- ----------------------------------------------------------------------------
DROP TABLE IF EXISTS `sale_detail`;
CREATE TABLE `sale_detail` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `unit_price` DOUBLE          NOT NULL,
    `quantity`   INT             NOT NULL,
    `sub_total`  DOUBLE          NOT NULL,
    `product_id` BIGINT UNSIGNED NOT NULL,
    `sale_id`    BIGINT UNSIGNED NOT NULL,
    `created_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_sale_detail_product_id` (`product_id`),
    KEY `idx_sale_detail_sale_id` (`sale_id`),
    CONSTRAINT `fk_sale_detail_product` FOREIGN KEY (`product_id`) REFERENCES `product` (`id`)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT `fk_sale_detail_sale` FOREIGN KEY (`sale_id`) REFERENCES `sale` (`id`)
        ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================================
-- MOCK DATA
-- ============================================================================

-- ----------------------------------------------------------------------------
-- _user
-- Passwords are real bcrypt hashes (cost 10) generated with golang.org/x/crypto/bcrypt,
-- matching what UserService.RegisterUser/LoginUser expects. Plaintext noted per row.
--   admin      -> Admin@123
--   cashier1   -> Cashier@123
--   johndoe123 -> SecretP@ss123
-- ----------------------------------------------------------------------------
INSERT INTO `_user` (`id`, `full_name`, `username`, `password`, `role`, `created_at`, `updated_at`) VALUES
(1, 'System Administrator', 'admin',      '$2a$10$kiHDToXfjbYd/.bXfbh84.3Oc0abZckzGkaVefR.y69zi.zHS4hDi', 'ADMIN',     NOW(), NOW()),
(2, 'Carla Cashier',        'cashier1',   '$2a$10$w2FVgbtAWHa584gO0LxxaeZ.X2j.FlhFjekMQA2S6fgzdm2twDc1q', 'CASHIER',   NOW(), NOW()),
(3, 'Ivan Inventory',       'inventory1', '$2a$10$w2FVgbtAWHa584gO0LxxaeZ.X2j.FlhFjekMQA2S6fgzdm2twDc1q', 'INVENTORY', NOW(), NOW()),
(4, 'John Doe',             'johndoe123', '$2a$10$e8lZzIi1HJN/Drv/pn0zTOgSo8s5RB0Nimy6LsAHFfXMispT6dNCG', 'CASHIER',   NOW(), NOW());

-- ----------------------------------------------------------------------------
-- category
-- ----------------------------------------------------------------------------
INSERT INTO `category` (`id`, `name`, `description`, `created_at`, `updated_at`) VALUES
(1, 'Beverages',      'Soft drinks, juices, water, and energy drinks', NOW(), NOW()),
(2, 'Snacks',         'Chips, cookies, and packaged snack foods',      NOW(), NOW()),
(3, 'Dairy',          'Milk, cheese, yogurt, and dairy products',      NOW(), NOW()),
(4, 'Bakery',         'Bread, pastries, and baked goods',              NOW(), NOW()),
(5, 'Cleaning',       'Household cleaning and hygiene supplies',       NOW(), NOW());

-- ----------------------------------------------------------------------------
-- client
-- ----------------------------------------------------------------------------
INSERT INTO `client` (`id`, `name`, `dni`, `email`, `created_at`, `updated_at`) VALUES
(1, 'Jane Doe',      '1712345678', 'jane.doe@example.com',      NOW(), NOW()),
(2, 'Mark Smith',    '1798765432', 'mark.smith@example.com',    NOW(), NOW()),
(3, 'Lucia Fernandez','1755566677', 'lucia.fernandez@example.com', NOW(), NOW());

-- ----------------------------------------------------------------------------
-- product
-- ----------------------------------------------------------------------------
INSERT INTO `product` (`id`, `name`, `brand`, `base_price`, `tax_percentage`, `category_id`, `stock`, `created_at`, `updated_at`) VALUES
(1, 'Coca Cola 1.5L',       'Coca Cola',  1.85, 12.0, 1, 100, NOW(), NOW()),
(2, 'Pepsi 1.5L',           'Pepsi',      1.75, 12.0, 1, 80,  NOW(), NOW()),
(3, 'Lays Classic 150g',    'Lays',       2.10, 12.0, 2, 60,  NOW(), NOW()),
(4, 'Oreo Cookies 154g',    'Oreo',       1.50, 12.0, 2, 120, NOW(), NOW()),
(5, 'Whole Milk 1L',        'Rey Leche',  1.20, 0.0,  3, 200, NOW(), NOW()),
(6, 'Cheddar Cheese 500g',  'Kraft',      4.50, 12.0, 3, 40,  NOW(), NOW()),
(7, 'White Bread Loaf',     'Bimbo',      1.10, 0.0,  4, 90,  NOW(), NOW()),
(8, 'Dish Soap 750ml',      'Axion',      2.30, 12.0, 5, 70,  NOW(), NOW());

-- ----------------------------------------------------------------------------
-- promotion
-- ----------------------------------------------------------------------------
INSERT INTO `promotion` (`id`, `name`, `discount_percentage`, `start_date`, `end_date`, `created_at`, `updated_at`) VALUES
(1, 'Weekend Beverage Discount', 10.0, '2026-07-10 00:00:00', '2026-07-31 23:59:59', NOW(), NOW()),
(2, 'Snack Attack Sale',         15.0, '2026-07-01 00:00:00', '2026-07-20 23:59:59', NOW(), NOW());

-- ----------------------------------------------------------------------------
-- promotion_products (many-to-many)
-- ----------------------------------------------------------------------------
INSERT INTO `promotion_products` (`promotion_id`, `product_id`) VALUES
(1, 1), -- Weekend Beverage Discount -> Coca Cola 1.5L
(1, 2), -- Weekend Beverage Discount -> Pepsi 1.5L
(2, 3), -- Snack Attack Sale -> Lays Classic 150g
(2, 4); -- Snack Attack Sale -> Oreo Cookies 154g

-- ----------------------------------------------------------------------------
-- sale
-- ----------------------------------------------------------------------------
INSERT INTO `sale` (`id`, `total_price`, `status`, `client_id`, `cashier_id`, `finished_at`, `created_at`) VALUES
(1, 5.58,  'CLOSED', 1,    2, '2026-07-14 17:30:00', '2026-07-14 17:00:00'),
(2, 3.20,  'OPEN',   NULL, 2, NULL,                   NOW()),
(3, 12.45, 'CLOSED', 2,    4, '2026-07-15 12:10:00', '2026-07-15 12:00:00');

-- ----------------------------------------------------------------------------
-- sale_detail
-- ----------------------------------------------------------------------------
INSERT INTO `sale_detail` (`id`, `unit_price`, `quantity`, `sub_total`, `product_id`, `sale_id`, `created_at`, `updated_at`) VALUES
(1, 1.86, 3, 5.58, 1, 1, NOW(), NOW()),
(2, 1.20, 2, 2.40, 5, 2, NOW(), NOW()),
(3, 2.35, 3, 7.05, 3, 3, NOW(), NOW()),
(4, 1.68, 2, 3.36, 4, 3, NOW(), NOW()),
(5, 2.04, 1, 2.04, 8, 3, NOW(), NOW());
