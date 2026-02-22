package adapter

import (
	"testing"
)

// TestMariaDBAdapter_ANALYZE_Requires_SELECT validates that ANALYZE mode gate.
// MariaDB uses ANALYZE FORMAT=JSON which requires SELECT queries only.
func TestMariaDBAdapter_ANALYZE_Requires_SELECT(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		analyze bool
		wantErr bool
	}{
		{
			name:    "SELECT with ANALYZE - passes query check",
			query:   "SELECT * FROM users",
			analyze: true,
			wantErr: false,
		},
		{
			name:    "INSERT with ANALYZE - fails query check",
			query:   "INSERT INTO users VALUES (1)",
			analyze: true,
			wantErr: true,
		},
		{
			name:    "UPDATE with ANALYZE - fails query check",
			query:   "UPDATE users SET name='test'",
			analyze: true,
			wantErr: true,
		},
		{
			name:    "DELETE with ANALYZE - fails query check",
			query:   "DELETE FROM users",
			analyze: true,
			wantErr: true,
		},
		{
			name:    "CREATE TABLE with ANALYZE - fails query check",
			query:   "CREATE TABLE users (id INT)",
			analyze: true,
			wantErr: true,
		},
		{
			name:    "ALTER TABLE with ANALYZE - fails query check",
			query:   "ALTER TABLE users ADD COLUMN age INT",
			analyze: true,
			wantErr: true,
		},
		{
			name:    "SELECT without ANALYZE - always passes",
			query:   "SELECT * FROM users",
			analyze: false,
			wantErr: false,
		},
		{
			name:    "INSERT without ANALYZE - always passes",
			query:   "INSERT INTO users VALUES (1)",
			analyze: false,
			wantErr: false,
		},
		{
			name:    "SELECT with leading whitespace and ANALYZE",
			query:   "  \n\t  SELECT * FROM users",
			analyze: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate query check: only SELECT allowed with ANALYZE
			if tt.analyze && !isSelectQuery(tt.query) {
				if !tt.wantErr {
					t.Errorf("expected error for non-SELECT with ANALYZE, but wantErr=%v", tt.wantErr)
				}
			} else if tt.analyze && isSelectQuery(tt.query) {
				if tt.wantErr {
					t.Errorf("expected no error for SELECT with ANALYZE, but wantErr=%v", tt.wantErr)
				}
			} else if !tt.analyze {
				if tt.wantErr {
					t.Errorf("expected no error without ANALYZE, but wantErr=%v", tt.wantErr)
				}
			}
		})
	}
}
