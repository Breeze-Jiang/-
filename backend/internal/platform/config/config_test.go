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

func TestLoadRequiresFixedTestSMSCodeOnlyInTestEnvironment(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("APP_ENV", "test")
	t.Setenv("TEST_SMS_CODE", "")
	if _, err := Load(); err == nil {
		t.Fatal("test environment without TEST_SMS_CODE was accepted")
	}

	t.Setenv("TEST_SMS_CODE", "123456")
	if _, err := Load(); err != nil {
		t.Fatalf("valid test SMS configuration was rejected: %v", err)
	}

	t.Setenv("TEST_SMS_CODE", "invalid")
	if _, err := Load(); err == nil {
		t.Fatal("invalid TEST_SMS_CODE was accepted")
	}
}

func TestLoadRejectsTestSMSCodeInProduction(t *testing.T) {
	setRequiredTestEnvironment(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("ALIYUN_ACCESS_KEY_ID", "access-key")
	t.Setenv("ALIYUN_ACCESS_KEY_SECRET", "access-secret")
	t.Setenv("ALIYUN_SMS_SIGN_NAME", "sign")
	t.Setenv("ALIYUN_SMS_TEMPLATE_CODE", "template")
	t.Setenv("TEST_SMS_CODE", "123456")
	if _, err := Load(); err == nil {
		t.Fatal("production configuration accepted TEST_SMS_CODE")
	}
}

func setRequiredTestEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("REDIS_URL", "redis://test")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("CURSOR_SECRET", "abcdefghijklmnopqrstuvwxyzABCDEF")
	t.Setenv("AMAP_WEB_SERVICE_KEY", "test-key")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://client.example")
}
