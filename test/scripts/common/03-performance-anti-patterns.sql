-- Performance anti-patterns: redundant indexes + slow query patterns
-- Populates performance_schema.events_statements_summary_by_digest
-- Compatible with MySQL 5.7, 8.0 and MariaDB 10.11, 11.x

-- =============================================================================
-- Redundant indexes (detectable by DBSight index analysis)
-- =============================================================================

-- products: idx_products_name is prefix-redundant with idx_products_name_cat
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_name_cat ON products(name, category_id);

-- orders: idx_orders_status is prefix-redundant with idx_orders_status_created
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_status_created ON orders(status, created_at);

-- order_items: idx_oi_order is prefix-redundant with idx_oi_order_product
CREATE INDEX idx_oi_order ON order_items(order_id);
CREATE INDEX idx_oi_order_product ON order_items(order_id, product_id);

-- =============================================================================
-- Slow query patterns (populate performance_schema digest stats)
-- =============================================================================

-- Full table scan: no index on orders.user_id
SELECT * FROM orders WHERE user_id = 1;
SELECT * FROM orders WHERE user_id = 5;
SELECT * FROM orders WHERE user_id = 10;

-- LIKE wildcard: full table scan on reviews (zero indexes)
SELECT * FROM reviews WHERE content LIKE '%great%';
SELECT * FROM reviews WHERE content LIKE '%amazing%';

-- Large JOIN without optimal indexes
SELECT o.*, oi.*, p.name
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
JOIN products p ON oi.product_id = p.id
WHERE o.status = 'pending';

SELECT o.*, oi.*, p.name
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
JOIN products p ON oi.product_id = p.id
WHERE o.status = 'delivered';

-- Subquery: correlated lookup
SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total_amount > 100);
SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total_amount > 500);

-- Filesort: ORDER BY on unindexed column
SELECT * FROM orders ORDER BY created_at DESC LIMIT 50;
SELECT * FROM orders ORDER BY total_amount DESC LIMIT 20;

-- GROUP BY with HAVING
SELECT product_id, COUNT(*) as cnt FROM order_items GROUP BY product_id HAVING cnt > 5;
SELECT user_id, COUNT(*) as cnt FROM orders GROUP BY user_id HAVING cnt > 3;

-- DISTINCT on non-indexed columns
SELECT DISTINCT user_id, status FROM orders;
SELECT DISTINCT product_id, rating FROM reviews;

-- Full table scan on reviews (zero indexes)
SELECT * FROM reviews WHERE rating >= 4;
SELECT * FROM reviews WHERE user_id = 1;
SELECT * FROM reviews ORDER BY created_at DESC LIMIT 100;
