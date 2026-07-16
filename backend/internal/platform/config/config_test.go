package config

import "testing"

func TestLoadRejectsProviderTimeoutAboveP0Budget(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("PROVIDER_TIMEOUT", "3s")
	if _, err := Load(); err == nil {
		t.Fatal("expected provider timeout above P0 budget to be rejected")
	}
}

func TestLoadAcceptsProviderTimeoutWithinP0Budget(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("PROVIDER_TIMEOUT", "2200ms")
	if _, err := Load(); err != nil {
		t.Fatalf("expected valid configuration: %v", err)
	}
}

func setRequiredTestEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("REDIS_URL", "redis://test")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("CURSOR_SECRET", "abcdefghijklmnopqrstuvwxyzABCDEF")
	t.Setenv("AMAP_WEB_SERVICE_KEY", "test-key")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://client.example")
}
