//go:generate mockgen -destination=mock/mailer_mock.go -package=mock . Mailer
package email

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	"github.com/ferdiebergado/gojeep/internal/config"
)

type templateMap map[string]*template.Template

type Mailer interface {
	SendPlain(to []string, subject string, body string) error
	SendHTML(to []string, subject string, tmplName string, data map[string]string) error
}

type mailer struct {
	from      string
	pass      string
	host      string
	port      int
	sender    string
	templates templateMap
}

var _ Mailer = (*mailer)(nil)

func New(cfg *config.SMTPConfig) (Mailer, error) {
	opts := cfg.Options
	path := opts.TemplatePath
	layoutFile := filepath.Join(path, opts.LayoutFile)
	layoutTmpl := template.Must(template.New("layout").ParseFiles(layoutFile))
	tmplMap, err := parsePages(path, layoutTmpl)
	if err != nil {
		return nil, err
	}

	return &mailer{
		from:      cfg.User,
		pass:      cfg.Password,
		host:      cfg.Host,
		port:      cfg.Port,
		sender:    opts.Sender,
		templates: tmplMap,
	}, nil
}

func (e *mailer) send(to []string, subject string, body string, contentType string) error {
	from := e.from
	host := e.host
	auth := smtp.PlainAuth(
		"",
		from,
		e.pass,
		host,
	)

	recipients := strings.Join(to, ", ")
	headers := "From: " + e.sender + "\r\n" +
		"To: " + recipients + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0\r\n" +
		"Content-Type: " + contentType + "; charset=\"UTF-8\"\r\n\r\n"

	message := headers + body
	addr := fmt.Sprintf("%s:%d", host, e.port)

	err := smtp.SendMail(
		addr,
		auth,
		from,
		to,
		[]byte(message),
	)

	if err != nil {
		return err
	}

	slog.Info("Email sent.")
	return nil
}

func (e *mailer) SendHTML(to []string, subject string, tmplName string, data map[string]string) error {
	tmpl, ok := e.templates[tmplName]
	if !ok {
		return fmt.Errorf("template does not exist: %s", tmplName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	return e.send(to, subject, buf.String(), "text/html")
}

func (e *mailer) SendPlain(to []string, subject string, body string) error {
	return e.send(to, subject, body, "text/plain")
}

func parsePages(templateDir string, layoutTmpl *template.Template) (templateMap, error) {
	tmplMap := make(templateMap)
	err := fs.WalkDir(os.DirFS(templateDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		const suffix = ".html"
		if !d.IsDir() && strings.HasSuffix(path, suffix) {
			name := strings.TrimPrefix(path, "/")
			name = strings.TrimSuffix(name, suffix)
			tmplMap[name] = template.Must(template.Must(layoutTmpl.Clone()).ParseFiles(filepath.Join(templateDir, path)))
			slog.Debug("parsed page", "path", path, "name", name)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("load pages templates: %w", err)
	}

	return tmplMap, nil
}
