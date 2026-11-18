package store

import (
	"database/sql"
	"testing"
)

func TestConfigurePostgresPoolBoundsConnections(t *testing.T) {
	t.Setenv("UNISON_DB_MAX_CONNS", "")
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	configurePostgresPool(conn)
	if got := conn.Stats().MaxOpenConnections; got != defaultPostgresMaxOpenConns {
		t.Fatalf("default MaxOpenConnections = %d, want %d", got, defaultPostgresMaxOpenConns)
	}

	t.Setenv("UNISON_DB_MAX_CONNS", "14")
	configurePostgresPool(conn)
	if got := conn.Stats().MaxOpenConnections; got != 14 {
		t.Fatalf("override MaxOpenConnections = %d, want 14", got)
	}

	t.Setenv("UNISON_DB_MAX_CONNS", "invalid")
	configurePostgresPool(conn)
	if got := conn.Stats().MaxOpenConnections; got != defaultPostgresMaxOpenConns {
		t.Fatalf("invalid env MaxOpenConnections = %d, want %d", got, defaultPostgresMaxOpenConns)
	}
}
