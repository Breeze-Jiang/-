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

type DisabledSender struct{}

func (DisabledSender) SendCode(context.Context, string, string) error {
	return errors.New("SMS provider is not configured")
}
