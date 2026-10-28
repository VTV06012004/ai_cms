package database

import (
	"testing"
)

func TestNewMySQLConnectionRequiresDSN(t *testing.T) {
	t.Setenv("DB_DSN", "")

	if _, err := NewMySQLConnection(); err == nil {
		t.Fatal("NewMySQLConnection succeeded without DB_DSN")
	}
}
