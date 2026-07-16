package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment            string
	HTTPAddr               string
	DatabaseURL            string
	RedisURL               string
	JWTSecret              string
	AccessTTL              time.Duration
	RefreshTTL             time.Duration
	AmapBaseURL            string
	AmapKey                string
	AllowedOrigins         []string
	ProviderTimeout        time.Duration
	TrustedProxyCIDRs      []string
	CursorSecret           string
	AdminAddr              string
	RideHailingURLTemplate string
	AliyunAccessKeyID      string
	AliyunAccessKeySecret  string
	AliyunSMSSignName      string
	AliyunSMSTemplateCode  string
}

func Load() (Config, error) {
	c := Config{
		Environment:            env("APP_ENV", "development"),
		HTTPAddr:               env("HTTP_ADDR", ":8080"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		RedisURL:               os.Getenv("REDIS_URL"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		AccessTTL:              duration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTTL:             duration("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		AmapBaseURL:            env("AMAP_BASE_URL", "https://restapi.amap.com"),
		AmapKey:                os.Getenv("AMAP_WEB_SERVICE_KEY"),
		AllowedOrigins:         csv(os.Getenv("CORS_ALLOWED_ORIGINS")),
		ProviderTimeout:        duration("PROVIDER_TIMEOUT", 2200*time.Millisecond),
		TrustedProxyCIDRs:      csv(os.Getenv("TRUSTED_PROXY_CIDRS")),
		CursorSecret:           os.Getenv("CURSOR_SECRET"),
		AdminAddr:              env("ADMIN_ADDR", "127.0.0.1:9090"),
		RideHailingURLTemplate: os.Getenv("RIDE_HAILING_URL_TEMPLATE"),
		AliyunAccessKeyID:      os.Getenv("ALIYUN_ACCESS_KEY_ID"), AliyunAccessKeySecret: os.Getenv("ALIYUN_ACCESS_KEY_SECRET"), AliyunSMSSignName: os.Getenv("ALIYUN_SMS_SIGN_NAME"), AliyunSMSTemplateCode: os.Getenv("ALIYUN_SMS_TEMPLATE_CODE"),
	}
	var missing []string
	for k, v := range map[string]string{"DATABASE_URL": c.DatabaseURL, "REDIS_URL": c.RedisURL, "JWT_SECRET": c.JWTSecret, "CURSOR_SECRET": c.CursorSecret, "AMAP_WEB_SERVICE_KEY": c.AmapKey} {
		if strings.TrimSpace(v) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	if len(c.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	if len(c.CursorSecret) < 32 || c.CursorSecret == c.JWTSecret {
		return Config{}, errors.New("CURSOR_SECRET must contain at least 32 characters and differ from JWT_SECRET")
	}
	if c.AccessTTL <= 0 || c.RefreshTTL <= c.AccessTTL {
		return Config{}, errors.New("token TTL configuration is invalid")
	}
	if c.ProviderTimeout <= 0 || c.ProviderTimeout > 2200*time.Millisecond {
		return Config{}, errors.New("PROVIDER_TIMEOUT must be greater than zero and at most 2200ms")
	}
	if len(c.AllowedOrigins) == 0 {
		return Config{}, errors.New("CORS_ALLOWED_ORIGINS must not be empty")
	}
	if c.Environment == "production" && (c.AliyunAccessKeyID == "" || c.AliyunAccessKeySecret == "" || c.AliyunSMSSignName == "" || c.AliyunSMSTemplateCode == "") {
		return Config{}, errors.New("production requires complete Aliyun SMS configuration")
	}
	return c, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func duration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return -1
	}
	return d
}
func csv(v string) []string {
	var out []string
	for _, s := range strings.Split(v, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}
func Int(key string, fallback int) int {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return v
}
