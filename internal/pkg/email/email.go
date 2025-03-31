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

	"github.com/ferdiebergado/gopherkit/env"
)

const suffix = ".html"

type templateMap map[string]*template.Template

type TemplateConfig struct {
	Path       string
	PagesPath  string
	LayoutFile string
}

type Config struct {
	From     string
	To       string
	Pass     string
	Host     string
	Port     int
	Template TemplateConfig
}

type Email struct {
	from      string
	to        string
	pass      string
	host      string
	port      int
	templates templateMap
}

func New(cfg Config) (*Email, error) {
	path := cfg.Template.Path
	pagesPath := cfg.Template.PagesPath
	layoutFile := filepath.Join(path, cfg.Template.LayoutFile)
	layoutTmpl := template.Must(template.New("layout").ParseFiles(layoutFile))
	tmplMap, err := parsePages(path, pagesPath, layoutTmpl)
	if err != nil {
		return nil, err
	}

	return &Email{
		templates: tmplMap,
	}, nil
}

func (e *Email) Send(to []string, subject string, body string, contentType string) error {
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

func (e *Email) SendHTML(to []string, subject string, tmplName string, data map[string]string) error {
	tmpl, ok := e.templates[tmplName]
	if !ok {
		return fmt.Errorf("template does not exist: %s", tmplName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	return e.Send(to, subject, buf.String(), "text/html")
}

func (e *Email) SendPlain(to []string, subject string, body string) error {
	return e.Send(to, subject, body, "text/plain")
}

func parsePages(templateDir, pagesDir string, layoutTmpl *template.Template) (templateMap, error) {
	tmplMap := make(templateMap)
	err := fs.WalkDir(os.DirFS(templateDir), pagesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, suffix) {
			name := strings.TrimPrefix(path, pagesDir+"/")
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
