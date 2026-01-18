package dialect

import (
	"testing"
)

func TestFromConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		connStr  string
		wantType string
	}{
		// PostgreSQL variants
		{"postgres scheme", "postgres://user:pass@localhost/db", "postgresql"},
		{"postgresql scheme", "postgresql://user:pass@localhost/db", "postgresql"},
		{"postgres uppercase", "POSTGRES://user:pass@localhost/db", "postgresql"},
		{"postgresql mixed case", "PostgreSQL://user:pass@localhost/db", "postgresql"},

		// MySQL variants
		{"mysql scheme", "mysql://user:pass@localhost/db", "mysql"},
		{"mysql uppercase", "MYSQL://user:pass@localhost/db", "mysql"},

		// Default (backwards compatibility)
		{"no scheme defaults to postgres", "host=localhost user=postgres", "postgresql"},
		{"empty string defaults to postgres", "", "postgresql"},
		{"unknown scheme defaults to postgres", "unknown://localhost", "postgresql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromConnectionString(tt.connStr)
			if err != nil {
				t.Errorf("FromConnectionString() error = %v", err)
				return
			}
			if got.Name() != tt.wantType {
				t.Errorf("FromConnectionString() dialect = %v, want %v", got.Name(), tt.wantType)
			}
		})
	}
}

func TestFromName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
		wantErr  bool
	}{
		// PostgreSQL variants
		{"postgresql", "postgresql", "postgresql", false},
		{"postgres", "postgres", "postgresql", false},
		{"pg", "pg", "postgresql", false},
		{"PostgreSQL uppercase", "POSTGRESQL", "postgresql", false},
		{"Postgres mixed", "Postgres", "postgresql", false},

		// MySQL variants
		{"mysql", "mysql", "mysql", false},
		{"mariadb", "mariadb", "mysql", false},
		{"MySQL uppercase", "MYSQL", "mysql", false},
		{"MariaDB mixed", "MariaDB", "mysql", false},

		// Error cases
		{"unknown dialect", "oracle", "", true},
		{"empty string", "", "", true},
		{"sqlite unsupported", "sqlite", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Name() != tt.wantType {
				t.Errorf("FromName() dialect = %v, want %v", got.Name(), tt.wantType)
			}
		})
	}
}

func TestFromName_ErrorMessage(t *testing.T) {
	_, err := FromName("oracle")
	if err == nil {
		t.Error("FromName() expected error for unknown dialect")
		return
	}
	if err.Error() != "unknown dialect: oracle" {
		t.Errorf("FromName() error message = %v, want 'unknown dialect: oracle'", err.Error())
	}
}
