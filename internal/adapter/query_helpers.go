package adapter

import "strings"

// isSelectQuery returns true if the query is a read-only SELECT (case-insensitive).
// Handles both plain SELECT and CTE (WITH ... SELECT) queries.
// Used by MySQL and MariaDB adapters to gate EXPLAIN ANALYZE to SELECT-only.
func isSelectQuery(query string) bool {
	upper := strings.ToUpper(strings.TrimSpace(query))
	return strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "WITH")
}
