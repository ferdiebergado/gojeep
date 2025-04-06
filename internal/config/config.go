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
	User    string `json:"user" env:"POSTGRES_USER"`
	Pass    string `json:"pass" env:"POSTGRES_PASSWORD"`
	Host    string `json:"host" env:"POSTGRES_HOST"`
	Port    int    `json:"port" env:"POSTGRES_PORT"`
	SSLMode string `json:"ssl_mode" env:"POSTGRES_SSLMODE"`
	DB      string `json:"db" env:"POSTGRES_DB"`
}

func (c DBConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.Int("port", c.Port),
		slog.String("sslmode", c.SSLMode),
		slog.String("db", c.DB),
	)
}

type EmailConfig struct {
	From     string
	Password string
	Host     string
	Port     int
}

func (c EmailConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.Int("port", c.Port),
	)
}

type JWTConfig struct {
	Issuer string
}

func (c JWTConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("issuer", c.Issuer),
	)
}

// TODO: reduce mem usage
type Config struct {
	App     AppConfig
	DB      DBConfig
	Email   EmailConfig
	JWT     JWTConfig
	Options Options
}

func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app", c.App),
		slog.Any("db", c.DB),
		slog.Any("jwt", c.JWT),
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
		Email: EmailConfig{
			From:     env.MustGet("EMAIL_FROM"),
			Password: env.MustGet("EMAIL_PASSWORD"),
			Host:     env.MustGet("EMAIL_HOST"),
			Port:     env.GetInt("EMAIL_PORT", 587),
		},
		JWT: JWTConfig{
			Issuer: env.MustGet("JWT_ISSUER"),
		},
		Options: opts,
	}

	slog.Debug("config loaded", slog.Any("config", cfg))

	return cfg, nil
}
