#!/usr/bin/env bash
# generate-data.sh — Generate medium/heavy test data for DBSight test containers
# Usage: ./generate-data.sh [light|medium|heavy] [postgres|mysql57|mysql80|mariadb1011|mariadb11|all]
set -euo pipefail

SCALE="${1:-medium}"
TARGET="${2:-all}"
COMPOSE_FILE="$(cd "$(dirname "$0")/.." && pwd)/docker-compose.yml"
SERVICES=(postgres mysql57 mysql80 mariadb1011 mariadb11)

# Scale configuration
case "$SCALE" in
  light)
    echo "Light data already seeded by init scripts (~1,320 rows)."
    echo "Run 'docker compose -f $COMPOSE_FILE down -v && docker compose -f $COMPOSE_FILE up -d' to reset."
    exit 0
    ;;
  medium)
    USERS=5000; PRODUCTS=1000; ORDERS=10000; ORDER_ITEMS=30000; REVIEWS=5000
    ;;
  heavy)
    USERS=10000; PRODUCTS=2000; ORDERS=25000; ORDER_ITEMS=60000; REVIEWS=10000
    ;;
  *)
    echo "Usage: $0 [light|medium|heavy] [postgres|mysql57|mysql80|mariadb1011|mariadb11|all]"
    exit 1
    ;;
esac

# Filter targets
if [ "$TARGET" != "all" ]; then
  found=false
  for s in "${SERVICES[@]}"; do
    [ "$s" = "$TARGET" ] && found=true
  done
  if ! $found; then
    echo "Error: Unknown target '$TARGET'. Use: postgres|mysql57|mysql80|mariadb1011|mariadb11|all"
    exit 1
  fi
  SERVICES=("$TARGET")
fi

STATUSES=("pending" "processing" "shipped" "delivered" "cancelled")
BATCH=1000

# --- Helper: run SQL against a service (test-only credentials) ---
run_sql() {
  local svc="$1"
  if [ "$svc" = "postgres" ]; then
    docker compose -f "$COMPOSE_FILE" exec -T "$svc" \
      psql -U dbsight -d ecommerce -q 2>/dev/null
  else
    local client="mysql"
    # MariaDB 10.4+ ships 'mariadb' binary; 'mysql' symlink deprecated
    if docker compose -f "$COMPOSE_FILE" exec -T "$svc" which mariadb >/dev/null 2>&1; then
      client="mariadb"
    fi
    docker compose -f "$COMPOSE_FILE" exec -T "$svc" \
      "$client" -u dbsight -psecret ecommerce 2>/dev/null
  fi
}

# --- Helper: check service is running ---
check_service() {
  local svc="$1"
  if ! docker compose -f "$COMPOSE_FILE" ps --status running "$svc" 2>/dev/null | grep -q "$svc"; then
    echo "  Warning: $svc is not running, skipping."
    return 1
  fi
  return 0
}

# --- Generate and insert users ---
generate_users() {
  local svc="$1"
  echo "  Generating $USERS users..."
  local i=101  # start after seed data (100 users)
  local end=$((100 + USERS))
  while [ $i -le $end ]; do
    local batch_end=$((i + BATCH - 1))
    [ $batch_end -gt $end ] && batch_end=$end
    local sql="INSERT INTO users (name, email, phone, address) VALUES "
    local sep=""
    local j=$i
    while [ $j -le $batch_end ]; do
      sql+="${sep}('User ${j}','user${j}@example.com','555-${j}','${j} Test Street')"
      sep=","
      j=$((j + 1))
    done
    echo "${sql};" | run_sql "$svc"
    i=$((batch_end + 1))
  done
}

# --- Generate and insert products ---
generate_products() {
  local svc="$1"
  echo "  Generating $PRODUCTS products..."
  local i=201  # start after seed data (200 products)
  local end=$((200 + PRODUCTS))
  while [ $i -le $end ]; do
    local batch_end=$((i + BATCH - 1))
    [ $batch_end -gt $end ] && batch_end=$end
    local sql="INSERT INTO products (name, description, price, category_id, stock) VALUES "
    local sep=""
    local j=$i
    while [ $j -le $batch_end ]; do
      local cat_id=$(( (j % 20) + 1 ))
      local price=$(( (j % 500) + 10 )).99
      local stock=$(( (j % 200) + 5 ))
      sql+="${sep}('Product ${j}','Description for product ${j}',${price},${cat_id},${stock})"
      sep=","
      j=$((j + 1))
    done
    echo "${sql};" | run_sql "$svc"
    i=$((batch_end + 1))
  done
}

# --- Generate and insert orders ---
generate_orders() {
  local svc="$1"
  echo "  Generating $ORDERS orders..."
  local total_users=$((100 + USERS))
  local i=301  # start after seed data (300 orders)
  local end=$((300 + ORDERS))
  while [ $i -le $end ]; do
    local batch_end=$((i + BATCH - 1))
    [ $batch_end -gt $end ] && batch_end=$end
    local sql="INSERT INTO orders (user_id, status, total_amount, created_at) VALUES "
    local sep=""
    local j=$i
    while [ $j -le $batch_end ]; do
      local uid=$(( (j % total_users) + 1 ))
      local status="${STATUSES[$(( j % 5 ))]}"
      local amount=$(( (j % 2000) + 10 )).99
      # Spread dates across 2025
      local month=$(( (j % 12) + 1 ))
      local day=$(( (j % 28) + 1 ))
      local hour=$(( (j % 12) + 8 ))
      local ts
      ts=$(printf "2025-%02d-%02d %02d:00:00" "$month" "$day" "$hour")
      sql+="${sep}(${uid},'${status}',${amount},'${ts}')"
      sep=","
      j=$((j + 1))
    done
    echo "${sql};" | run_sql "$svc"
    i=$((batch_end + 1))
  done
}

# --- Generate and insert order_items ---
generate_order_items() {
  local svc="$1"
  echo "  Generating $ORDER_ITEMS order_items..."
  local total_orders=$((300 + ORDERS))
  local total_products=$((200 + PRODUCTS))
  local i=1
  while [ $i -le $ORDER_ITEMS ]; do
    local batch_end=$((i + BATCH - 1))
    [ $batch_end -gt $ORDER_ITEMS ] && batch_end=$ORDER_ITEMS
    local sql="INSERT INTO order_items (order_id, product_id, quantity, price) VALUES "
    local sep=""
    local j=$i
    while [ $j -le $batch_end ]; do
      local oid=$(( (j % total_orders) + 1 ))
      local pid=$(( (j % total_products) + 1 ))
      local qty=$(( (j % 5) + 1 ))
      local price=$(( (j % 500) + 5 )).99
      sql+="${sep}(${oid},${pid},${qty},${price})"
      sep=","
      j=$((j + 1))
    done
    echo "${sql};" | run_sql "$svc"
    i=$((batch_end + 1))
  done
}

# --- Generate and insert reviews ---
generate_reviews() {
  local svc="$1"
  echo "  Generating $REVIEWS reviews..."
  local total_users=$((100 + USERS))
  local total_products=$((200 + PRODUCTS))
  local i=201  # start after seed data (200 reviews)
  local end=$((200 + REVIEWS))
  while [ $i -le $end ]; do
    local batch_end=$((i + BATCH - 1))
    [ $batch_end -gt $end ] && batch_end=$end
    local sql="INSERT INTO reviews (user_id, product_id, rating, content, created_at) VALUES "
    local sep=""
    local j=$i
    while [ $j -le $batch_end ]; do
      local uid=$(( (j % total_users) + 1 ))
      local pid=$(( (j % total_products) + 1 ))
      local rating=$(( (j % 5) + 1 ))
      local month=$(( (j % 12) + 1 ))
      local day=$(( (j % 28) + 1 ))
      local ts
      ts=$(printf "2025-%02d-%02d 12:00:00" "$month" "$day")
      sql+="${sep}(${uid},${pid},${rating},'Review ${j} for product ${pid}','${ts}')"
      sep=","
      j=$((j + 1))
    done
    echo "${sql};" | run_sql "$svc"
    i=$((batch_end + 1))
  done
}

# --- Re-run slow queries to refresh query stats ---
refresh_query_stats() {
  local svc="$1"
  echo "  Refreshing query stats with slow queries..."
  if [ "$svc" = "postgres" ]; then
    # PostgreSQL: populates pg_stat_statements
    run_sql "$svc" <<'SQL'
SELECT * FROM orders WHERE user_id = 1;
SELECT * FROM orders WHERE user_id = 50;
SELECT * FROM orders WHERE user_id = 100;
SELECT * FROM reviews WHERE content LIKE '%great%';
SELECT * FROM reviews WHERE content LIKE '%amazing%';
SELECT o.*, oi.*, p.name FROM orders o JOIN order_items oi ON o.id = oi.order_id JOIN products p ON oi.product_id = p.id WHERE o.status = 'pending';
SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total_amount > 100);
SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total_amount > 500);
SELECT * FROM orders ORDER BY created_at DESC LIMIT 50;
SELECT * FROM orders ORDER BY total_amount DESC LIMIT 20;
SELECT product_id, COUNT(*) AS cnt FROM order_items GROUP BY product_id HAVING COUNT(*) > 5;
SELECT user_id, COUNT(*) AS cnt FROM orders GROUP BY user_id HAVING COUNT(*) > 3;
SELECT DISTINCT user_id, status FROM orders;
SELECT DISTINCT product_id, rating FROM reviews;
SELECT * FROM reviews WHERE rating >= 4;
SELECT * FROM reviews WHERE user_id = 1;
SELECT * FROM reviews ORDER BY created_at DESC LIMIT 100;
SQL
  else
    # MySQL/MariaDB: populates performance_schema
    run_sql "$svc" <<'SQL'
SELECT * FROM orders WHERE user_id = 1;
SELECT * FROM orders WHERE user_id = 50;
SELECT * FROM orders WHERE user_id = 100;
SELECT * FROM reviews WHERE content LIKE '%great%';
SELECT * FROM reviews WHERE content LIKE '%amazing%';
SELECT o.*, oi.*, p.name FROM orders o JOIN order_items oi ON o.id = oi.order_id JOIN products p ON oi.product_id = p.id WHERE o.status = 'pending';
SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total_amount > 100);
SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total_amount > 500);
SELECT * FROM orders ORDER BY created_at DESC LIMIT 50;
SELECT * FROM orders ORDER BY total_amount DESC LIMIT 20;
SELECT product_id, COUNT(*) as cnt FROM order_items GROUP BY product_id HAVING cnt > 5;
SELECT user_id, COUNT(*) as cnt FROM orders GROUP BY user_id HAVING cnt > 3;
SELECT DISTINCT user_id, status FROM orders;
SELECT DISTINCT product_id, rating FROM reviews;
SELECT * FROM reviews WHERE rating >= 4;
SELECT * FROM reviews WHERE user_id = 1;
SELECT * FROM reviews ORDER BY created_at DESC LIMIT 100;
SQL
  fi
}

# --- Print summary ---
print_summary() {
  local svc="$1"
  echo "  Row counts for $svc:"
  # Standard SQL works on both PostgreSQL and MySQL/MariaDB
  run_sql "$svc" <<'SQL'
SELECT 'users' AS tbl, COUNT(*) AS cnt FROM users
UNION ALL SELECT 'categories', COUNT(*) FROM categories
UNION ALL SELECT 'products', COUNT(*) FROM products
UNION ALL SELECT 'orders', COUNT(*) FROM orders
UNION ALL SELECT 'order_items', COUNT(*) FROM order_items
UNION ALL SELECT 'reviews', COUNT(*) FROM reviews;
SQL
}

# =============================================================================
# Main
# =============================================================================
echo "=== DBSight Data Generator ==="
echo "Scale: $SCALE | Target: $TARGET"
echo ""

for svc in "${SERVICES[@]}"; do
  echo "--- $svc ---"
  check_service "$svc" || continue

  generate_users "$svc"
  generate_products "$svc"
  generate_orders "$svc"
  generate_order_items "$svc"
  generate_reviews "$svc"
  refresh_query_stats "$svc"
  print_summary "$svc"
  echo ""
done

echo "=== Done ==="
