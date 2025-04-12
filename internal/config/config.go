package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ferdiebergado/gopherkit/env"
)

type Options struct {
	Server   ServerOptions   `json:"server,omitempty"`
	DB       DBOptions       `json:"db,omitempty"`
	Template TemplateOptions `json:"template,omitempty"`
	JWT      JWTOptions      `json:"jwt,omitempty"`
	Email    EmailOptions    `json:"email,omitempty"`
}

type EmailOptions struct {
	Sender string `json:"sender,omitempty"`
}

type ServerOptions struct {
	ReadTimeout     int `json:"read_timeout,omitempty"`
	WriteTimeout    int `json:"write_timeout,omitempty"`
	IdleTimeout     int `json:"idle_timeout,omitempty"`
	ShutdownTimeout int `json:"shutdown_timeout,omitempty"`
}

type DBOptions struct {
	Driver          string `json:"driver,omitempty"`
	PingTimeout     int    `json:"ping_timeout,omitempty"`
	MaxOpenConns    int    `json:"max_open_conns,omitempty"`
	MaxIdleConns    int    `json:"max_idle_conns,omitempty"`
	ConnMaxIdle     int    `json:"conn_max_idle,omitempty"`
	ConnMaxLifetime int    `json:"conn_max_lifetime,omitempty"`
}

type TemplateOptions struct {
	Path       string `json:"path,omitempty"`
	LayoutFile string `json:"layout_file,omitempty"`
}

type JWTOptions struct {
	JTILen uint32 `json:"jti_len,omitempty"`
	Issuer string `json:"issuer,omitempty"`
}

type AppConfig struct {
	URL      string
	Port     int
	Key      string
	Env      string
	LogLevel string
}

func (c AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("url", c.URL),
		slog.String("env", c.Env),
		slog.String("log_level", c.LogLevel),
	)
}

type DBConfig struct {
	User    string
	Pass    string
	Host    string
	Port    int
	SSLMode string
	DB      string
}

func (c DBConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.Int("port", c.Port),
		slog.String("sslmode", c.SSLMode),
		slog.String("db", c.DB),
	)
}

type SMTPConfig struct {
	User     string
	Password string
	Host     string
	Port     int
}

func (c SMTPConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.Int("port", c.Port),
	)
}

// TODO: reduce mem usage
type Config struct {
	App     AppConfig
	DB      DBConfig
	Email   SMTPConfig
	Options Options
}

func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app", c.App),
		slog.Any("db", c.DB),
		slog.Any("email", c.Email),
		slog.Any("options", c.Options),
	)
}

func New(cfgFile string) (*Config, error) {
	slog.Info("Loading config...", "path", cfgFile)
	cfgFile = filepath.Clean(cfgFile)
	configFile, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("open config file %s: %w", cfgFile, err)
	}

	var opts Options
	if err := json.Unmarshal(configFile, &opts); err != nil {
		return nil, fmt.Errorf("decode options %s %w", configFile, err)
	}

	cfg := &Config{
		App: AppConfig{
			URL:      env.MustGet("APP_URL"),
			Port:     env.GetInt("PORT", 8888),
			Key:      env.MustGet("APP_KEY"),
			Env:      env.Get("ENV", "development"),
			LogLevel: env.Get("LOG_LEVEL", "INFO"),
		},
		DB: DBConfig{
			User:    env.MustGet("POSTGRES_USER"),
			Pass:    env.MustGet("POSTGRES_PASSWORD"),
			Host:    env.MustGet("POSTGRES_HOST"),
			Port:    env.GetInt("POSTGRES_PORT", 5432),
			SSLMode: env.MustGet("POSTGRES_SSLMODE"),
			DB:      env.MustGet("POSTGRES_DB"),
		},
		Email: SMTPConfig{
			User:     env.MustGet("SMTP_USER"),
			Password: env.MustGet("SMTP_PASS"),
			Host:     env.MustGet("SMTP_HOST"),
			Port:     env.GetInt("SMTP_PORT", 587),
		},
		Options: opts,
	}

	slog.Debug("config loaded", slog.Any("config", cfg))

	return cfg, nil
}

func LoadFile(cfgFile string) (*Config, error) {
	cfg, err := New(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}
