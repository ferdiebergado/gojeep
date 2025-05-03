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
	Options  *ServerOptions
}

func (c *ServerConfig) LogValue() slog.Value {
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
	Options *DBOptions
}

func (c *DBConfig) LogValue() slog.Value {
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
	Options  *EmailOptions
}

func (c *SMTPConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.Int("port", c.Port),
		slog.Any("options", c.Options),
	)
}

type JWTOptions struct {
	JTILen   uint32 `json:"jti_len,omitempty"`
	Issuer   string `json:"issuer,omitempty"`
	Duration int    `json:"duration,omitempty"`
}

type Argon2Options struct {
	Memory     uint32 `json:"memory,omitempty"`
	Iterations uint32 `json:"iterations,omitempty"`
	Threads    uint8  `json:"threads,omitempty"`
	SaltLength uint32 `json:"salt_length,omitempty"`
	KeyLength  uint32 `json:"key_length,omitempty"`
}

type CookieOptions struct {
	Name   string `json:"name,omitempty"`
	MaxAge int    `json:"max_age,omitempty"`
}

type Options struct {
	Server *ServerOptions `json:"server,omitempty"`
	DB     *DBOptions     `json:"db,omitempty"`
	JWT    *JWTOptions    `json:"jwt,omitempty"`
	Email  *EmailOptions  `json:"email,omitempty"`
	Hash   *Argon2Options `json:"hash,omitempty"`
	Cookie *CookieOptions `json:"cookie,omitempty"`
}

type Config struct {
	Server *ServerConfig
	DB     *DBConfig
	Email  *SMTPConfig
	JWT    *JWTOptions
	Hash   *Argon2Options
	Cookie *CookieOptions
}

func (c *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("server", c.Server),
		slog.Any("db", c.DB),
		slog.Any("email", c.Email),
		slog.Any("jwt", c.JWT),
		slog.Any("hash", c.Hash),
		slog.Any("cookie", c.Cookie),
	)
}

func Load(cfgFile string) (*Config, error) {
	slog.Info("Loading config...", "path", cfgFile)
	opts, err := parseCfgFile(cfgFile)

	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: &ServerConfig{
			URL:      env.MustGet("SERVER_URL"),
			Port:     env.GetInt("SERVER_PORT", envDefaultAppPort),
			Key:      env.MustGet("SERVER_KEY"),
			Env:      env.Get("SERVER_ENV", "development"),
			LogLevel: env.Get("SERVER_LOG_LEVEL", "INFO"),
			Options:  opts.Server,
		},
		DB: &DBConfig{
			User:    env.MustGet("POSTGRES_USER"),
			Pass:    env.MustGet("POSTGRES_PASSWORD"),
			Host:    env.MustGet("POSTGRES_HOST"),
			Port:    env.GetInt("POSTGRES_PORT", envDefaultDBPort),
			SSLMode: env.MustGet("POSTGRES_SSLMODE"),
			DB:      env.MustGet("POSTGRES_DB"),
			Options: opts.DB,
		},
		Email: &SMTPConfig{
			User:     env.MustGet("SMTP_USER"),
			Password: env.MustGet("SMTP_PASS"),
			Host:     env.MustGet("SMTP_HOST"),
			Port:     env.GetInt("SMTP_PORT", envDefaultSMTPPort),
			Options:  opts.Email,
		},
		JWT:    opts.JWT,
		Hash:   opts.Hash,
		Cookie: opts.Cookie,
	}

	slog.Debug("config loaded", slog.Any("config", cfg))

	return cfg, nil
}

func parseCfgFile(cfgFile string) (*Options, error) {
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
