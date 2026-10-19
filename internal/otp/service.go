package otp

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

var ErrInvalid = errors.New("invalid or expired OTP")
var ErrCooldown = errors.New("please wait before requesting another OTP")
var ErrAttempts = errors.New("too many verification attempts")

type Sender interface {
	Send(ctx context.Context, to, code string) error
}
type record struct {
	hash              string
	expires, resendAt time.Time
	attempts          int
}
type Service struct {
	mu      sync.Mutex
	entries map[string]record
	pepper  string
	sender  Sender
}

func New(sender Sender, pepper string) *Service {
	return &Service{entries: make(map[string]record), pepper: pepper, sender: sender}
}
func (s *Service) digest(email, code string) string {
	mac := hmac.New(sha256.New, []byte(s.pepper))
	mac.Write([]byte(email + ":" + code))
	return hex.EncodeToString(mac.Sum(nil))
}
func normalize(email string) string { return strings.ToLower(strings.TrimSpace(email)) }
func (s *Service) Send(ctx context.Context, email string) error {
	email = normalize(email)
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if old, ok := s.entries[email]; ok && now.Before(old.resendAt) {
		return ErrCooldown
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	// Lock is held until SMTP accepts the email, avoiding concurrent sends for the same address.
	if err := s.sender.Send(ctx, email, code); err != nil {
		return err
	}
	s.entries[email] = record{hash: s.digest(email, code), expires: time.Now().Add(5 * time.Minute), resendAt: time.Now().Add(60 * time.Second)}
	return nil
}
func (s *Service) Verify(email, code string) error {
	email = normalize(email)
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.entries[email]
	if !ok || !time.Now().Before(r.expires) {
		delete(s.entries, email)
		return ErrInvalid
	}
	if r.attempts >= 5 {
		return ErrAttempts
	}
	r.attempts++
	if !hmac.Equal([]byte(r.hash), []byte(s.digest(email, code))) {
		s.entries[email] = r
		return ErrInvalid
	}
	delete(s.entries, email)
	return nil
}
