package httpapi

import (
	"email-otp-go/internal/otp"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"
)

type Handler struct{ Service *otp.Service }

func write(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
func decode(w http.ResponseWriter, r *http.Request, dest any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(dest); err != nil {
		write(w, 400, "invalid JSON request")
		return false
	}
	return true
}
func validEmail(email string) bool {
	a, err := mail.ParseAddress(email)
	return err == nil && a.Address == email && len(email) <= 254 && !strings.ContainsAny(email, "\r\n")
}
func (h Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { write(w, 200, "ok") })
	mux.HandleFunc("POST /auth/send-otp", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email string `json:"email"`
		}
		if !decode(w, r, &req) {
			return
		}
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		if !validEmail(req.Email) {
			write(w, 400, "invalid email")
			return
		}
		ctx := r.Context()
		if err := h.Service.Send(ctx, req.Email); err != nil {
			log.Printf("send OTP to %q failed: %v", req.Email, err)
			if errors.Is(err, otp.ErrCooldown) {
				write(w, 429, "please wait 60 seconds before retrying")
				return
			}
			write(w, 502, "unable to send OTP")
			return
		}
		write(w, 200, "OTP sent; valid for 5 minutes")
	})
	mux.HandleFunc("POST /auth/verify-otp", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email string `json:"email"`
			OTP   string `json:"otp"`
		}
		if !decode(w, r, &req) {
			return
		}
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		if !validEmail(req.Email) || len(req.OTP) != 6 {
			write(w, 400, "invalid email or OTP format")
			return
		}
		for _, c := range req.OTP {
			if c < '0' || c > '9' {
				write(w, 400, "OTP must contain 6 digits")
				return
			}
		}
		if err := h.Service.Verify(req.Email, req.OTP); err != nil {
			write(w, 400, "invalid, expired, or locked OTP")
			return
		}
		write(w, 200, "email verified")
	})
	return http.TimeoutHandler(mux, 25*time.Second, `{"message":"request timeout"}`)
}
