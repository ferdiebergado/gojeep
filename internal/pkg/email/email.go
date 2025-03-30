package email

import (
	"bytes"
	"html/template"
	"net/smtp"
	"path/filepath"

	"github.com/ferdiebergado/gopherkit/env"
)

func SendEmail(to []string, subject string, body string, contentType string) error {
	from := env.MustGet("SMTP_FROM")
	pass := env.MustGet("SMTP_PASSWORD")
	host := env.MustGet("SMTP_HOST")
	port := env.MustGet("SMTP_PORT")

	auth := smtp.PlainAuth(
		"",
		from,
		pass,
		host,
	)

	headers := "To: " + to[0] + "\r\n" +
		"MIME-version: 1.0\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: " + contentType + "; charset=\"UTF-8\"" + "\r\n\r\n"

	message := headers + body

	return smtp.SendMail(
		host+":"+port,
		auth,
		from,
		to,
		[]byte(message),
	)
}

func SendHTMLEmail(to []string, subject string, tmpl string, data map[string]string) error {
	const templateDir = "web/templates"
	t, err := template.ParseFiles(filepath.Join(templateDir, "base.html"), filepath.Join(templateDir, tmpl+".html"))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}

	return SendEmail(to, subject, buf.String(), "text/html")
}

func SendPlainEmail(to []string, subject string, body string) error {
	return SendEmail(to, subject, body, "text/plain")
}
