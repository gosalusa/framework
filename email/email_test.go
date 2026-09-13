package email_test

import (
	"bufio"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gosalusa.com/di"
	"gosalusa.com/email"
)

type testConfig struct {
	port int
}

func (c *testConfig) GetHTTPPort() int { return c.port }
func (c *testConfig) GetBaseURL() string {
	return "https://example.com"
}

func TestSMTPMailer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	received := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		var buf strings.Builder
		write := func(line string) {
			_, _ = conn.Write([]byte(line))
		}

		write("220 localhost ESMTP\r\n")
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimRight(line, "\r\n")
			buf.WriteString(line + "\n")

			switch {
			case strings.HasPrefix(line, "EHLO"):
				write("250-localhost\r\n250 8BITMIME\r\n")
			case strings.HasPrefix(line, "MAIL FROM"):
				write("250 OK\r\n")
			case strings.HasPrefix(line, "RCPT TO"):
				write("250 OK\r\n")
			case strings.HasPrefix(line, "DATA"):
				write("354 End data with <CR><LF>.<CR><LF>\r\n")
				for {
					dl, err := reader.ReadString('\n')
					if err != nil {
						break
					}
					buf.WriteString(dl)
					if strings.TrimRight(dl, "\r\n") == "." {
						write("250 OK\r\n")
						break
					}
				}
			case strings.HasPrefix(line, "QUIT"):
				write("221 Bye\r\n")
				received <- buf.String()
				return
			}
		}
	}()

	port := ln.Addr().(*net.TCPAddr).Port
	m := email.NewSMTPMailer("127.0.0.1", port, "user", "pass", "from@example.com")
	err = m.Mail(&email.Message{
		To:       []string{"to@example.com"},
		Subject:  "Hello",
		HTMLBody: "<p>hi</p>",
	})
	require.NoError(t, err)

	data := <-received
	assert.Contains(t, data, "MAIL FROM:<from@example.com>")
	assert.Contains(t, data, "RCPT TO:<to@example.com>")
	assert.Contains(t, data, "Subject: Hello")
	assert.Contains(t, data, "<p>hi</p>")
}

func TestRegister(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	cfg := &email.SMTPConfig{}

	email.Register(ctx, cfg)

	m, err := di.Resolve[email.Mailer](ctx)
	assert.NoError(t, err)
	assert.NotNil(t, m)
}
