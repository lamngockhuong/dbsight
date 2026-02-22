package adapter

import (
	"testing"
)

func TestIsSelectQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{
			name:  "SELECT uppercase",
			query: "SELECT * FROM users",
			want:  true,
		},
		{
			name:  "SELECT lowercase",
			query: "select * from users",
			want:  true,
		},
		{
			name:  "SELECT mixed case",
			query: "SeLeCt * FROM users",
			want:  true,
		},
		{
			name:  "SELECT with leading whitespace",
			query: "  \n\t  SELECT * FROM users",
			want:  true,
		},
		{
			name:  "INSERT",
			query: "INSERT INTO users VALUES (1)",
			want:  false,
		},
		{
			name:  "UPDATE",
			query: "UPDATE users SET name='test'",
			want:  false,
		},
		{
			name:  "DELETE",
			query: "DELETE FROM users WHERE id=1",
			want:  false,
		},
		{
			name:  "INSERT with leading space",
			query: "  INSERT INTO users VALUES (1)",
			want:  false,
		},
		{
			name:  "Empty string",
			query: "",
			want:  false,
		},
		{
			name:  "Whitespace only",
			query: "   \n\t  ",
			want:  false,
		},
		{
			name:  "SELECT in middle",
			query: "EXPLAIN SELECT * FROM users",
			want:  false,
		},
		{
			name:  "CREATE TABLE",
			query: "CREATE TABLE users (id INT)",
			want:  false,
		},
		{
			name:  "DROP TABLE",
			query: "DROP TABLE users",
			want:  false,
		},
		{
			name:  "ALTER TABLE",
			query: "ALTER TABLE users ADD COLUMN name VARCHAR(255)",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSelectQuery(tt.query)
			if got != tt.want {
				t.Errorf("isSelectQuery(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}
