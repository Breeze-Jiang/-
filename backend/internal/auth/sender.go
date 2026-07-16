package auth

import (
	"context"
	"errors"
	"log/slog"
)

type DevelopmentSender struct{ log *slog.Logger }

func NewDevelopmentSender(log *slog.Logger) *DevelopmentSender { return &DevelopmentSender{log: log} }
func (s *DevelopmentSender) SendCode(ctx context.Context, phone, code string) error {
	s.log.Warn("development SMS delivery simulated", "phone_suffix", phone[len(phone)-4:])
	return nil
}

// FixedTestSender is intentionally available only when APP_ENV=test. It makes
// a local or CI verification code deterministic without exposing it through an
// HTTP endpoint or application logs.
type FixedTestSender struct{ code string }

func NewFixedTestSender(code string) (FixedTestSender, error) {
	if err := ValidateCode(code); err != nil {
		return FixedTestSender{}, err
	}
	return FixedTestSender{code: code}, nil
}

func (s FixedTestSender) SendCode(_ context.Context, _ string, code string) error {
	if code != s.code {
		return errors.New("test SMS code did not match configured value")
	}
	return nil
}

func (s FixedTestSender) fixedCode() string { return s.code }

type DisabledSender struct{}

func (DisabledSender) SendCode(context.Context, string, string) error {
	return errors.New("SMS provider is not configured")
}
