package main

import (
	"email-otp-go/internal/config"
	"email-otp-go/internal/httpapi"
	"email-otp-go/internal/otp"
	"log"
	"net/http"
)

func main() {
	config.LoadDotEnv(".env")
	c := config.Load()
	if c.User == "" || c.Password == "" || c.From == "" || len(c.Pepper) < 24 {
		log.Fatal("set SMTP_USER, SMTP_PASSWORD, SMTP_FROM, and OTP_PEPPER (24+ chars) in .env")
	}
	sender := otp.SMTP{Host: c.Host, Port: c.Port, User: c.User, Password: c.Password, From: c.From}
	service := otp.New(sender, c.Pepper)
	server := &http.Server{Addr: c.Addr, Handler: (httpapi.Handler{Service: service}).Routes(), ReadHeaderTimeout: 5_000_000_000}
	log.Printf("listening on %s", c.Addr)
	log.Fatal(server.ListenAndServe())
}
