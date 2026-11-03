package database

import (
	"testing"
)

func TestNewMySQLConnectionRequiresDSN(t *testing.T) {
	if _, err := NewMySQLConnection(""); err == nil {
		t.Fatal("NewMySQLConnection succeeded without a DSN")
	}
}
