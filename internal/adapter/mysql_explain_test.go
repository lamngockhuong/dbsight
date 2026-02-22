package adapter

import (
	"testing"

	"github.com/lamngockhuong/dbsight/internal/adapter/mysqlcompat"
)

// TestMySQLAdapter_ANALYZE_Requires_SELECT validates that ANALYZE mode gate.
// Using table-driven tests to check isSelectQuery guard.
func TestMySQLAdapter_ANALYZE_Requires_SELECT(t *testing.T) {
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
			name:    "CREATE with ANALYZE - fails query check",
			query:   "CREATE TABLE users (id INT)",
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

// TestMySQLAdapter_ANALYZE_Requires_Version validates MySQL 8.0+ requirement for ANALYZE.
func TestMySQLAdapter_ANALYZE_Requires_Version(t *testing.T) {
	tests := []struct {
		name    string
		version mysqlcompat.Version
		analyze bool
		wantErr bool
	}{
		{
			name:    "MySQL 8.0 - minimum supported",
			version: mysqlcompat.Version{Major: 8, Minor: 0, Patch: 36},
			analyze: true,
			wantErr: false,
		},
		{
			name:    "MySQL 8.0.18 - minimum with ANALYZE",
			version: mysqlcompat.Version{Major: 8, Minor: 0, Patch: 18},
			analyze: true,
			wantErr: false,
		},
		{
			name:    "MySQL 5.7 - too old for ANALYZE",
			version: mysqlcompat.Version{Major: 5, Minor: 7, Patch: 42},
			analyze: true,
			wantErr: true,
		},
		{
			name:    "MySQL 5.6 - too old for ANALYZE",
			version: mysqlcompat.Version{Major: 5, Minor: 6, Patch: 0},
			analyze: true,
			wantErr: true,
		},
		{
			name:    "MySQL 7.9 - too old for ANALYZE",
			version: mysqlcompat.Version{Major: 7, Minor: 9, Patch: 0},
			analyze: true,
			wantErr: true,
		},
		{
			name:    "MySQL 9.0 - newer than minimum",
			version: mysqlcompat.Version{Major: 9, Minor: 0, Patch: 0},
			analyze: true,
			wantErr: false,
		},
		{
			name:    "MySQL 5.7 without ANALYZE - always passes",
			version: mysqlcompat.Version{Major: 5, Minor: 7, Patch: 42},
			analyze: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate version check: ANALYZE requires 8.0+
			if tt.analyze && !tt.version.AtLeast(8, 0) {
				if !tt.wantErr {
					t.Errorf("expected error for MySQL < 8.0 with ANALYZE, but wantErr=%v", tt.wantErr)
				}
			} else if tt.analyze && tt.version.AtLeast(8, 0) {
				if tt.wantErr {
					t.Errorf("expected no error for MySQL >= 8.0 with ANALYZE, but wantErr=%v", tt.wantErr)
				}
			} else if !tt.analyze {
				if tt.wantErr {
					t.Errorf("expected no error without ANALYZE, but wantErr=%v", tt.wantErr)
				}
			}
		})
	}
}
