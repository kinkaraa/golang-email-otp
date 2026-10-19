package otp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type SMTP struct{ Host, Port, User, Password, From string }

func (s SMTP) Send(ctx context.Context, to, code string) error {
	if _, err := mail.ParseAddress(to); err != nil {
		return errors.New("invalid recipient")
	}
	if _, err := mail.ParseAddress(s.From); err != nil {
		return errors.New("invalid sender")
	}
	if strings.ContainsAny(to, "\r\n") || strings.ContainsAny(s.From, "\r\n") {
		return errors.New("invalid address")
	}
	if _, err := strconv.Atoi(s.Port); err != nil {
		return errors.New("invalid SMTP port")
	}
	addr := net.JoinHostPort(s.Host, s.Port)
	d := net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(20 * time.Second))
	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); !ok {
		return errors.New("SMTP server does not support STARTTLS")
	}
	if err := client.StartTLS(&tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12}); err != nil {
		return err
	}
	if err := client.Auth(smtp.PlainAuth("", s.User, s.Password, s.Host)); err != nil {
		return err
	}
	if err := client.Mail(s.From); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Kode Verifikasi OTP\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nKode OTP Anda: %s\r\nBerlaku selama 5 menit. Jangan bagikan kode ini.\r\n", s.From, to, code)
	if _, err := w.Write([]byte(msg)); err != nil {
		w.Close()
		return err
	}
	return w.Close()
}
