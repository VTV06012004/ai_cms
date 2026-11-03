package config

import (
	"errors"
	"reflect"
	"testing"
)

func envFrom(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func validEnv() map[string]string {
	return map[string]string{
		"JWT_SECRET": "a-long-enough-secret",
		"DB_DSN":     "user:pass@tcp(db:3306)/cms",
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	cfg, err := load(envFrom(validEnv()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" {
		t.Fatalf("port = %q, want 8080", cfg.Port)
	}
	if !reflect.DeepEqual(cfg.CORSAllowedOrigins, defaultCORSOrigins) {
		t.Fatalf("cors = %v, want defaults", cfg.CORSAllowedOrigins)
	}
}

func TestLoadRequiresJWTSecretAndDSN(t *testing.T) {
	env := validEnv()
	env["JWT_SECRET"] = "   "
	if _, err := load(envFrom(env)); !errors.Is(err, ErrMissingJWTSecret) {
		t.Fatalf("expected ErrMissingJWTSecret, got %v", err)
	}

	env = validEnv()
	env["DB_DSN"] = ""
	if _, err := load(envFrom(env)); !errors.Is(err, ErrMissingDSN) {
		t.Fatalf("expected ErrMissingDSN, got %v", err)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	for _, port := range []string{"abc", "0", "70000", "-1"} {
		env := validEnv()
		env["PORT"] = port
		if _, err := load(envFrom(env)); err == nil {
			t.Errorf("PORT=%q was accepted", port)
		}
	}
}

func TestLoadParsesCustomValues(t *testing.T) {
	env := validEnv()
	env["PORT"] = "9090"
	env["CORS_ALLOWED_ORIGINS"] = " http://a.test , ,http://b.test "
	env["INITIAL_ADMIN_EMAIL"] = "admin@example.com"

	cfg, err := load(envFrom(env))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "9090" {
		t.Fatalf("port = %q", cfg.Port)
	}
	want := []string{"http://a.test", "http://b.test"}
	if !reflect.DeepEqual(cfg.CORSAllowedOrigins, want) {
		t.Fatalf("cors = %v, want %v", cfg.CORSAllowedOrigins, want)
	}
	if cfg.InitialAdmin.Email != "admin@example.com" {
		t.Fatalf("initial admin email = %q", cfg.InitialAdmin.Email)
	}
}
