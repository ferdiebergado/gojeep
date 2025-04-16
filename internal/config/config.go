package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ferdiebergado/gopherkit/env"
)

const (
	envDefaultAppPort  = 8888
	envDefaultDBPort   = 5432
	envDefaultSMTPPort = 587
)

type ServerOptions struct {
	ReadTimeout     int `json:"read_timeout,omitempty"`
	WriteTimeout    int `json:"write_timeout,omitempty"`
	IdleTimeout     int `json:"idle_timeout,omitempty"`
	ShutdownTimeout int `json:"shutdown_timeout,omitempty"`
}

type ServerConfig struct {
	URL      string
	Port     int
	Key      string
	Env      string
	LogLevel string
	Options  ServerOptions
}

func (c ServerConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("url", c.URL),
		slog.String("env", c.Env),
		slog.String("log_level", c.LogLevel),
		slog.Any("options", c.Options),
	)
}

type DBOptions struct {
	Driver          string `json:"driver,omitempty"`
	PingTimeout     int    `json:"ping_timeout,omitempty"`
	MaxOpenConns    int    `json:"max_open_conns,omitempty"`
	MaxIdleConns    int    `json:"max_idle_conns,omitempty"`
	ConnMaxIdle     int    `json:"conn_max_idle,omitempty"`
	ConnMaxLifetime int    `json:"conn_max_lifetime,omitempty"`
}

type DBConfig struct {
	User    string
	Pass    string
	Host    string
	Port    int
	SSLMode string
	DB      string
	Options DBOptions
}

func (c DBConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.Int("port", c.Port),
		slog.String("sslmode", c.SSLMode),
		slog.String("db", c.DB),
		slog.Any("options", c.Options),
	)
}

type EmailOptions struct {
	Sender       string `json:"sender,omitempty"`
	VerifyTTL    int    `json:"verify_ttl,omitempty"`
	TemplatePath string `json:"template_path,omitempty"`
	LayoutFile   string `json:"layout_file,omitempty"`
}

type SMTPConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Options  EmailOptions
}

func (c SMTPConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.Int("port", c.Port),
		slog.Any("options", c.Options),
	)
}

type JWTOptions struct {
	JTILen uint32 `json:"jti_len,omitempty"`
	Issuer string `json:"issuer,omitempty"`
}

type Options struct {
	Server ServerOptions `json:"server,omitempty"`
	DB     DBOptions     `json:"db,omitempty"`
	JWT    JWTOptions    `json:"jwt,omitempty"`
	Email  EmailOptions  `json:"email,omitempty"`
}

// TODO: reduce mem usage
type Config struct {
	Server ServerConfig
	DB     DBConfig
	Email  SMTPConfig
	JWT    JWTOptions
}

func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app", c.Server.Options),
		slog.Any("db", c.DB.Options),
		slog.Any("email", c.Email.Options),
		slog.Any("jwt", c.JWT),
	)
}

func New(cfgFile string) (*Config, error) {
	slog.Info("Loading config...", "path", cfgFile)
	opts, err := loadCfgFile(cfgFile)

	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			URL:      env.MustGet("APP_URL"),
			Port:     env.GetInt("PORT", envDefaultAppPort),
			Key:      env.MustGet("APP_KEY"),
			Env:      env.Get("ENV", "development"),
			LogLevel: env.Get("LOG_LEVEL", "INFO"),
			Options:  opts.Server,
		},
		DB: DBConfig{
			User:    env.MustGet("POSTGRES_USER"),
			Pass:    env.MustGet("POSTGRES_PASSWORD"),
			Host:    env.MustGet("POSTGRES_HOST"),
			Port:    env.GetInt("POSTGRES_PORT", envDefaultDBPort),
			SSLMode: env.MustGet("POSTGRES_SSLMODE"),
			DB:      env.MustGet("POSTGRES_DB"),
			Options: opts.DB,
		},
		Email: SMTPConfig{
			User:     env.MustGet("SMTP_USER"),
			Password: env.MustGet("SMTP_PASS"),
			Host:     env.MustGet("SMTP_HOST"),
			Port:     env.GetInt("SMTP_PORT", envDefaultSMTPPort),
			Options:  opts.Email,
		},
		JWT: opts.JWT,
	}

	slog.Debug("config loaded", slog.Any("config", cfg))

	return cfg, nil
}

func loadCfgFile(cfgFile string) (*Options, error) {
	cfgFile = filepath.Clean(cfgFile)
	configFile, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", cfgFile, err)
	}

	var opts Options
	if err := json.Unmarshal(configFile, &opts); err != nil {
		return nil, fmt.Errorf("decode json config %s: %w", configFile, err)
	}

	return &opts, nil
}
