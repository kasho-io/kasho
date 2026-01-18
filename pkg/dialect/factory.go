package dialect

import (
	"fmt"
	"strings"
)

// FromConnectionString returns the appropriate dialect based on the connection string
func FromConnectionString(connStr string) (Dialect, error) {
	connStr = strings.ToLower(connStr)

	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		return NewPostgreSQL(), nil
	}

	if strings.HasPrefix(connStr, "mysql://") {
		return NewMySQL(), nil
	}

	// Default to PostgreSQL for backwards compatibility with connection strings
	// that don't have a scheme prefix (common for lib/pq)
	return NewPostgreSQL(), nil
}

// FromName returns the dialect by name
func FromName(name string) (Dialect, error) {
	switch strings.ToLower(name) {
	case "postgresql", "postgres", "pg":
		return NewPostgreSQL(), nil
	case "mysql", "mariadb":
		return NewMySQL(), nil
	default:
		return nil, fmt.Errorf("unknown dialect: %s", name)
	}
}
