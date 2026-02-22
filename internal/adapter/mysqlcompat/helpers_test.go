package mysqlcompat

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantMajor int
		wantMinor int
		wantPatch int
		wantMaria bool
		wantErr   bool
	}{
		{
			name:      "MySQL 8.0",
			raw:       "8.0.36",
			wantMajor: 8,
			wantMinor: 0,
			wantPatch: 36,
			wantMaria: false,
			wantErr:   false,
		},
		{
			name:      "MySQL 8.0 with Ubuntu suffix",
			raw:       "8.0.36-0ubuntu0.22.04.1",
			wantMajor: 8,
			wantMinor: 0,
			wantPatch: 36,
			wantMaria: false,
			wantErr:   false,
		},
		{
			name:      "MySQL 5.7",
			raw:       "5.7.42",
			wantMajor: 5,
			wantMinor: 7,
			wantPatch: 42,
			wantMaria: false,
			wantErr:   false,
		},
		{
			name:      "MariaDB 10.11",
			raw:       "10.11.6-MariaDB",
			wantMajor: 10,
			wantMinor: 11,
			wantPatch: 6,
			wantMaria: true,
			wantErr:   false,
		},
		{
			name:      "MariaDB 11.2",
			raw:       "11.2.2-MariaDB-1:11.2.2+maria~ubu2204",
			wantMajor: 11,
			wantMinor: 2,
			wantPatch: 2,
			wantMaria: true,
			wantErr:   false,
		},
		{
			name:      "MariaDB with lowercase mariadb",
			raw:       "10.5.15-mariadb-focal",
			wantMajor: 10,
			wantMinor: 5,
			wantPatch: 15,
			wantMaria: true,
			wantErr:   false,
		},
		{
			name:      "Empty string",
			raw:       "",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
			wantMaria: false,
			wantErr:   true,
		},
		{
			name:      "Garbage string",
			raw:       "not a version",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
			wantMaria: false,
			wantErr:   true,
		},
		{
			name:      "Single number",
			raw:       "8",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
			wantMaria: false,
			wantErr:   true,
		},
		{
			name:      "Major.Minor only",
			raw:       "8.0",
			wantMajor: 8,
			wantMinor: 0,
			wantPatch: 0,
			wantMaria: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseVersion(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if v.Major != tt.wantMajor {
				t.Errorf("Major = %d, want %d", v.Major, tt.wantMajor)
			}
			if v.Minor != tt.wantMinor {
				t.Errorf("Minor = %d, want %d", v.Minor, tt.wantMinor)
			}
			if v.Patch != tt.wantPatch {
				t.Errorf("Patch = %d, want %d", v.Patch, tt.wantPatch)
			}
			if v.IsMariaDB != tt.wantMaria {
				t.Errorf("IsMariaDB = %v, want %v", v.IsMariaDB, tt.wantMaria)
			}
			if v.Raw != tt.raw {
				t.Errorf("Raw = %q, want %q", v.Raw, tt.raw)
			}
		})
	}
}

func TestVersion_AtLeast(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		major   int
		minor   int
		want    bool
	}{
		{
			name:    "8.0 >= 8.0",
			version: Version{Major: 8, Minor: 0, Patch: 36},
			major:   8,
			minor:   0,
			want:    true,
		},
		{
			name:    "8.0 >= 5.7",
			version: Version{Major: 8, Minor: 0, Patch: 36},
			major:   5,
			minor:   7,
			want:    true,
		},
		{
			name:    "5.7 < 8.0",
			version: Version{Major: 5, Minor: 7, Patch: 42},
			major:   8,
			minor:   0,
			want:    false,
		},
		{
			name:    "10.11 >= 10.5",
			version: Version{Major: 10, Minor: 11, Patch: 6},
			major:   10,
			minor:   5,
			want:    true,
		},
		{
			name:    "10.5 >= 10.11",
			version: Version{Major: 10, Minor: 5, Patch: 15},
			major:   10,
			minor:   11,
			want:    false,
		},
		{
			name:    "11.2 >= 10.1",
			version: Version{Major: 11, Minor: 2, Patch: 2},
			major:   10,
			minor:   1,
			want:    true,
		},
		{
			name:    "Major difference boundary",
			version: Version{Major: 9, Minor: 9, Patch: 9},
			major:   10,
			minor:   0,
			want:    false,
		},
		{
			name:    "Major difference high",
			version: Version{Major: 12, Minor: 0, Patch: 0},
			major:   10,
			minor:   0,
			want:    true,
		},
		{
			name:    "Same major, equal minor",
			version: Version{Major: 10, Minor: 11, Patch: 0},
			major:   10,
			minor:   11,
			want:    true,
		},
		{
			name:    "Same major, greater minor",
			version: Version{Major: 10, Minor: 12, Patch: 0},
			major:   10,
			minor:   11,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.AtLeast(tt.major, tt.minor)
			if got != tt.want {
				t.Errorf("AtLeast(%d, %d) = %v, want %v", tt.major, tt.minor, got, tt.want)
			}
		})
	}
}
