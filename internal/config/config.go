package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
)

type AppConfig struct {
	URL     string `json:"url,omitempty" env:"APP_URL"`
	Key     string `json:"key,omitempty" env:"APP_KEY"`
	Env     string `json:"env,omitempty" env:"ENV"`
	IsDebug bool   `json:"is_debug,omitempty" env:"DEBUG"`
}

func (c AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("url", c.URL),
		slog.String("env", c.Env),
		slog.Bool("debug", c.IsDebug),
	)
}

type DBConfig struct {
	Driver          string `json:"driver,omitempty"`
	User            string `json:"user" env:"POSTGRES_USER"`
	Pass            string `json:"pass" env:"POSTGRES_PASSWORD"`
	Host            string `json:"host" env:"POSTGRES_HOST"`
	Port            int    `json:"port" env:"POSTGRES_PORT"`
	SSLMode         string `json:"ssl_mode" env:"POSTGRES_SSLMODE"`
	PingTimeout     int    `json:"ping_timeout,omitempty"`
	DB              string `json:"db" env:"POSTGRES_DB"`
	MaxOpenConns    int    `json:"max_open_conns,omitempty"`
	MaxIdleConns    int    `json:"max_idle_conns,omitempty"`
	ConnMaxIdle     int    `json:"conn_max_idle,omitempty"`
	ConnMaxLifetime int    `json:"conn_max_lifetime,omitempty"`
}

func (c DBConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("driver", c.Driver),
		slog.String("host", c.Host),
		slog.Int("port", c.Port),
		slog.String("sslmode", c.SSLMode),
		slog.Int("ping_timeout", c.PingTimeout),
		slog.String("db", c.DB),
		slog.Int("max_open_conns", c.MaxOpenConns),
		slog.Int("max_idle_conns", c.MaxIdleConns),
		slog.Int("conn_max_idle", c.ConnMaxIdle),
		slog.Int("conn_max_lifetime", c.ConnMaxLifetime),
	)
}

type ServerConfig struct {
	Port            int `json:"port" env:"PORT"`
	ReadTimeout     int `json:"read_timeout,omitempty"`
	WriteTimeout    int `json:"write_timeout,omitempty"`
	IdleTimeout     int `json:"idle_timeout,omitempty"`
	ShutdownTimeout int `json:"shutdown_timeout,omitempty"`
}

type TemplateConfig struct {
	Path       string `json:"path,omitempty"`
	LayoutFile string `json:"layout_file,omitempty"`
}

type EmailConfig struct {
	From     string `json:"from,omitempty" env:"EMAIL_FROM"`
	Password string `json:"password,omitempty" env:"EMAIL_PASSWORD"`
	Host     string `json:"host,omitempty" env:"EMAIL_HOST"`
	Port     int    `json:"port,omitempty" env:"EMAIL_PORT"`
}

func (c EmailConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.Int("port", c.Port),
	)
}

type JWTConfig struct {
	SigningKey string `json:"signing_key,omitempty" env:"APP_KEY"`
	KeyLen     uint32 `json:"key_len,omitempty"`
	Issuer     string `json:"issuer,omitempty" env:"APP_URL"`
}

func (c JWTConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Uint64("keylen", uint64(c.KeyLen)),
		slog.String("issuer", c.Issuer),
	)
}

// TODO: reduce mem usage
type Config struct {
	App      AppConfig      `json:"app,omitempty"`
	Db       DBConfig       `json:"db,omitempty"`
	Server   ServerConfig   `json:"server,omitempty"`
	Email    EmailConfig    `json:"email,omitempty"`
	Template TemplateConfig `json:"template,omitempty"`
	JWT      JWTConfig      `json:"jwt,omitempty"`
}

func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app", c.App),
		slog.Any("db", c.Db),
		slog.Any("jwt", c.JWT),
		slog.Any("email", c.Email),
		slog.Any("template", c.Template),
		slog.Any("server", c.Server),
	)
}

func New(cfgFile string) (*Config, error) {
	slog.Info("Loading config...", "path", cfgFile)
	cfgFile = filepath.Clean(cfgFile)
	configFile, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("open config file %s: %w", cfgFile, err)
	}

	var cfg Config
	if err := json.Unmarshal(configFile, &cfg); err != nil {
		return nil, fmt.Errorf("decode config %s %w", configFile, err)
	}

	overrideWithEnv(reflect.ValueOf(&cfg).Elem())

	slog.Debug("new config", slog.Any("config", cfg))

	return &cfg, nil
}

func overrideWithEnv(v reflect.Value) {
	v = derefPointer(v)
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return
	}

	typeOfV := v.Type()
	for i := range v.NumField() {
		field := v.Field(i)
		structField := typeOfV.Field(i)

		if handleNestedStruct(field) {
			continue
		}

		processEnvOverride(field, structField.Tag.Get("env"))
	}
}

func derefPointer(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}

func handleNestedStruct(field reflect.Value) bool {
	if field.Kind() == reflect.Struct {
		overrideWithEnv(field)
		return true
	}
	return false
}

func processEnvOverride(field reflect.Value, envTag string) {
	if envTag == "" {
		return
	}

	if envVal, exists := os.LookupEnv(envTag); exists {
		setFieldFromEnv(field, envVal)
	}
}

func setFieldFromEnv(field reflect.Value, envVal string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(envVal)
	case reflect.Int:
		setIntField(field, envVal)
	case reflect.Bool:
		setBoolField(field, envVal)
	}
}

func setIntField(field reflect.Value, envVal string) {
	if intVal, err := strconv.Atoi(envVal); err == nil {
		field.SetInt(int64(intVal))
	}
}

func setBoolField(field reflect.Value, envVal string) {
	if boolVal, err := strconv.ParseBool(envVal); err == nil {
		field.SetBool(boolVal)
	}
}
