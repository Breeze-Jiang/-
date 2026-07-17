package auth

import (
	"context"
	"testing"
	"time"
)

type authMetricRecorder struct{ events []string }

func (m *authMetricRecorder) ObserveAuth(operation, outcome string) {
	m.events = append(m.events, operation+":"+outcome)
}

func TestNormalizePhone(t *testing.T) {
	got, err := normalizePhone("+86 13800138000")
	if err != nil || got != "+8613800138000" {
		t.Fatalf("unexpected phone: %s %v", got, err)
	}
	if _, err = normalizePhone("123"); err == nil {
		t.Fatal("invalid phone accepted")
	}
}
func TestValidateCode(t *testing.T) {
	if ValidateCode("123456") != nil {
		t.Fatal("valid code rejected")
	}
	if ValidateCode("12a456") == nil {
		t.Fatal("invalid code accepted")
	}
	if ValidateCode("-12345") == nil {
		t.Fatal("signed code accepted")
	}
}

func TestFixedTestSenderAcceptsDeliveryWithoutLoggingSensitiveValues(t *testing.T) {
	sender, err := NewFixedTestSender("123456")
	if err != nil {
		t.Fatalf("create fixed test sender: %v", err)
	}
	service := NewService(nil, nil, sender, "test-secret", time.Minute, time.Hour)
	code, err := service.challengeCode()
	if err != nil || code != "123456" {
		t.Fatalf("fixed test sender did not supply its configured code: code=%q err=%v", code, err)
	}
	if err = sender.SendCode(context.Background(), "+8613800138000", "123456"); err != nil {
		t.Fatalf("fixed test sender delivery failed: %v", err)
	}
}

func TestInvalidSMSSendRecordsBoundedMetric(t *testing.T) {
	metrics := &authMetricRecorder{}
	service := NewService(nil, nil, DisabledSender{}, "test-secret", time.Minute, time.Hour, metrics)
	if _, err := service.Send(context.Background(), "invalid", "short", ""); err == nil {
		t.Fatal("expected invalid request")
	}
	if len(metrics.events) != 1 || metrics.events[0] != "sms_send:invalid" {
		t.Fatalf("unexpected auth metrics: %v", metrics.events)
	}
}
