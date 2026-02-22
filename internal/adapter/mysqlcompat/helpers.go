package mysqlcompat

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// Version holds parsed MySQL/MariaDB version info.
type Version struct {
	Major     int
	Minor     int
	Patch     int
	IsMariaDB bool
	Raw       string
}

// ParseVersion parses SELECT VERSION() output.
// MySQL: "8.0.36" or "8.0.36-0ubuntu0.22.04.1"
// MariaDB: "10.11.6-MariaDB" or "11.2.2-MariaDB-1:11.2.2+maria~ubu2204"
func ParseVersion(raw string) (Version, error) {
	v := Version{Raw: raw}
	v.IsMariaDB = strings.Contains(strings.ToLower(raw), "mariadb")

	// Extract numeric prefix before first non-digit-dot char
	numPart := raw
	for i, c := range raw {
		if c != '.' && (c < '0' || c > '9') {
			numPart = raw[:i]
			break
		}
	}

	parts := strings.SplitN(numPart, ".", 3)
	if len(parts) < 2 {
		return v, fmt.Errorf("cannot parse version: %s", raw)
	}

	var err error
	if v.Major, err = strconv.Atoi(parts[0]); err != nil {
		return v, fmt.Errorf("parse major: %w", err)
	}
	if v.Minor, err = strconv.Atoi(parts[1]); err != nil {
		return v, fmt.Errorf("parse minor: %w", err)
	}
	if len(parts) == 3 {
		v.Patch, _ = strconv.Atoi(parts[2]) // patch optional
	}
	return v, nil
}

// AtLeast returns true if version >= major.minor.
func (v Version) AtLeast(major, minor int) bool {
	if v.Major != major {
		return v.Major > major
	}
	return v.Minor >= minor
}

// AtLeastPatch returns true if version >= major.minor.patch.
func (v Version) AtLeastPatch(major, minor, patch int) bool {
	if v.Major != major {
		return v.Major > major
	}
	if v.Minor != minor {
		return v.Minor > minor
	}
	return v.Patch >= patch
}

// DetectVersion queries SELECT VERSION() and parses result.
func DetectVersion(ctx context.Context, db *sql.DB) (Version, error) {
	var raw string
	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&raw); err != nil {
		return Version{}, fmt.Errorf("detect version: %w", err)
	}
	return ParseVersion(raw)
}

// CheckPerfSchema returns true if performance_schema is enabled.
func CheckPerfSchema(ctx context.Context, db *sql.DB) (bool, error) {
	var val string
	err := db.QueryRowContext(ctx,
		"SHOW VARIABLES LIKE 'performance_schema'",
	).Scan(new(string), &val)
	if err != nil {
		return false, fmt.Errorf("check performance_schema: %w", err)
	}
	return strings.ToUpper(val) == "ON", nil
}
