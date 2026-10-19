# Golang Email OTP API

Starter project REST API OTP via Gmail SMTP (STARTTLS) with Go standard library, no external dependencies.

## Setup

Requires Go 1.22+ and Gmail 2-Step Verification + App Password.

```bash
cp .env.example .env
# Edit .env with your SMTP credentials and a random OTP_PEPPER (24+ characters)
go run ./cmd/server
```

Test endpoints:

```bash
curl -X POST http://localhost:8080/auth/send-otp \
  -H 'Content-Type: application/json' \
  -d '{"email":"recipient@example.com"}'

curl -X POST http://localhost:8080/auth/verify-otp \
  -H 'Content-Type: application/json' \
  -d '{"email":"recipient@example.com","otp":"123456"}'

go test ./...
```

## Behavior

- OTP: 6 digits generated with crypto/rand.
- TTL: 5 minutes; resend cooldown: 60 seconds.
- Maximum 5 failed verification attempts per OTP; successful verification is single-use.
- Stored as HMAC-SHA256 digest with a secret pepper, not plaintext.
- No OTP is logged or returned in HTTP responses.

## Important limitations

This is a **development starter**, not a complete production auth system. OTP records live **in memory**, so restart clears them and multiple replicas do not share state. Old expired entries may remain until the email is accessed again. Sending SMTP while holding a lock serializes sends. For production use Redis with atomic operations and TTL, per-IP/email rate limits, request/SMTP metrics, an email queue, account enumeration protection, TLS termination, proper email delivery monitoring, and a real authenticated session/token flow after verification. This demo only verifies possession of an email code; it does not create a login session.

Never commit `.env` or expose App Passwords. For Gmail, use App Password, not your normal account password. Port 587 must be reachable from your deployment host.
