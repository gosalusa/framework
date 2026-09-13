package emailtest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/email"
	"gosalusa.com/email/emailtest"
)

func TestNewTestMailer(t *testing.T) {
	var _ email.Mailer = (*emailtest.TestMailer)(nil)

	m := emailtest.NewTestMailer()
	assert.Empty(t, m.EmailsSent())

	msg := &email.Message{
		To:       []string{"a@example.com"},
		Subject:  "hi",
		HTMLBody: "<p>hello</p>",
	}
	err := m.Mail(msg)
	assert.NoError(t, err)
	err = m.Mail(&email.Message{To: []string{"b@example.com"}})
	assert.NoError(t, err)

	sent := m.EmailsSent()
	assert.Len(t, sent, 2)
	assert.Same(t, msg, sent[0])
	assert.Equal(t, "b@example.com", sent[1].To[0])
}
