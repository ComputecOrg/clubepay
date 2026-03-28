package email

import (
	"fmt"
	"net/smtp"
)

// Sender defines the interface for sending emails.
type Sender interface {
	Send(to, subject, body string) error
}

// SMTPSender sends emails via SMTP.
type SMTPSender struct {
	Host     string
	Port     string
	Username string
	Password string
}

// NewSMTP creates a new SMTP sender. Returns nil if host is empty.
func NewSMTP(host, port, username, password string) *SMTPSender {
	if host == "" {
		return nil
	}
	return &SMTPSender{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
}

func (s *SMTPSender) Send(to, subject, body string) error {
	from := s.Username
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
	addr := s.Host + ":" + s.Port

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

// MockSender records sent messages for testing.
type MockSender struct {
	Sent []Message
}

// Message represents a sent email message.
type Message struct {
	To      string
	Subject string
	Body    string
}

func (m *MockSender) Send(to, subject, body string) error {
	m.Sent = append(m.Sent, Message{To: to, Subject: subject, Body: body})
	return nil
}
