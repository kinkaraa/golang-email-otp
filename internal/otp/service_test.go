package otp

import (
	"context"
	"errors"
	"testing"
)

type fakeSender struct{ code string }

func (f *fakeSender) Send(_ context.Context, _, code string) error { f.code = code; return nil }
func TestOTP(t *testing.T) {
	f := &fakeSender{}
	s := New(f, "test-pepper")
	if err := s.Send(context.Background(), "test@example.com"); err != nil {
		t.Fatal(err)
	}
	if len(f.code) != 6 {
		t.Fatal("OTP must have six digits")
	}
	if !errors.Is(s.Send(context.Background(), "test@example.com"), ErrCooldown) {
		t.Fatal("expected cooldown")
	}
	if !errors.Is(s.Verify("test@example.com", "invalid"), ErrInvalid) {
		t.Fatal("expected invalid OTP")
	}
	if err := s.Verify("test@example.com", f.code); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(s.Verify("test@example.com", f.code), ErrInvalid) {
		t.Fatal("OTP should be single-use")
	}
}
