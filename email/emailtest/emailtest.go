package emailtest

import (
	"gosalusa.com/email"
)

type TestMailer struct {
	messages []*email.Message
}

var _ email.Mailer = (*TestMailer)(nil)

func NewTestMailer() *TestMailer {
	return &TestMailer{
		messages: []*email.Message{},
	}
}
func (m *TestMailer) Mail(msg *email.Message) error {
	m.messages = append(m.messages, msg)
	return nil
}

func (m *TestMailer) EmailsSent() []*email.Message {
	return m.messages
}

type TestMailerConfig struct {
}

func NewTestMailerConfig() *TestMailerConfig {
	return &TestMailerConfig{}
}

func (c *TestMailerConfig) Mailer() email.Mailer {
	return NewTestMailer()
}
