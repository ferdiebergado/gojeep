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

type EnvConfig struct {
	Env     string `json:"env,omitempty" env:"ENV"`
	IsDebug bool   `json:"is_debug,omitempty" env:"DEBUG"`
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

type ServerConfig struct {
	Port            int `json:"port" env:"PORT"`
	ReadTimeout     int `json:"read_timeout,omitempty"`
	WriteTimeout    int `json:"write_timeout,omitempty"`
	IdleTimeout     int `json:"idle_timeout,omitempty"`
	ShutdownTimeout int `json:"shutdown_timeout,omitempty"`
}

type TemplateConfig struct {
	Path         string `json:"path,omitempty"`
	LayoutFile   string `json:"layout_file,omitempty"`
	PartialsPath string `json:"partials_path,omitempty"`
	PagesPath    string `json:"pages_path,omitempty"`
}

type Config struct {
	App      EnvConfig      `json:"app,omitempty"`
	Db       DBConfig       `json:"db,omitempty"`
	Server   ServerConfig   `json:"server,omitempty"`
	Template TemplateConfig `json:"template,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	slog.Info("Loading config...", "path", path)
	path = filepath.Clean(path)
	configFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("open config file %s: %w", path, err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		return nil, fmt.Errorf("decode config %s %w", configFile, err)
	}

	overrideWithEnv(reflect.ValueOf(&config).Elem())

	cfgCopy := config
	cfgCopy.Db.Pass = "*"

	slog.Debug("loadconfig", slog.Any("config", cfgCopy))

	return &config, nil
}

func overrideWithEnv(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	typeOfV := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := typeOfV.Field(i)
		if field.Kind() == reflect.Struct {
			overrideWithEnv(field)
			continue
		}
		envTag := structField.Tag.Get("env")
		if envTag == "" {
			continue
		}
		if envVal, exists := os.LookupEnv(envTag); exists {
			switch field.Kind() {
			case reflect.String:
				field.SetString(envVal)
			case reflect.Int:
				if intVal, err := strconv.Atoi(envVal); err == nil {
					field.SetInt(int64(intVal))
				}
			case reflect.Bool:
				if boolVal, err := strconv.ParseBool(envVal); err == nil {
					field.SetBool(boolVal)
				}
			}
		}
	}
}
