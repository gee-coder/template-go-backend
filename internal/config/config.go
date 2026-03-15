package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the root application config.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Auth     AuthConfig     `mapstructure:"auth"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Swagger  SwaggerConfig  `mapstructure:"swagger"`
}

// AppConfig describes global application settings.
type AppConfig struct {
	Name      string `mapstructure:"name"`
	Env       string `mapstructure:"env"`
	Debug     bool   `mapstructure:"debug"`
	UploadDir string `mapstructure:"uploadDir"`
}

// AuthConfig describes public auth settings.
type AuthConfig struct {
	EnableEmailLogin        bool `mapstructure:"enableEmailLogin"`
	EnablePhoneLogin        bool `mapstructure:"enablePhoneLogin"`
	EnableEmailRegistration bool `mapstructure:"enableEmailRegistration"`
	EnablePhoneRegistration bool `mapstructure:"enablePhoneRegistration"`
}

// HTTPConfig describes HTTP server settings.
type HTTPConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"readTimeout"`
	WriteTimeout    time.Duration `mapstructure:"writeTimeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdownTimeout"`
	AllowedOrigins  []string      `mapstructure:"allowedOrigins"`
}

// DatabaseConfig describes MySQL settings.
type DatabaseConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
}

// RedisConfig describes Redis settings.
type RedisConfig struct {
	Addr      string `mapstructure:"addr"`
	Password  string `mapstructure:"password"`
	DB        int    `mapstructure:"db"`
	KeyPrefix string `mapstructure:"keyPrefix"`
}

// JWTConfig describes JWT settings.
type JWTConfig struct {
	Issuer     string        `mapstructure:"issuer"`
	Secret     string        `mapstructure:"secret"`
	AccessTTL  time.Duration `mapstructure:"accessTTL"`
	RefreshTTL time.Duration `mapstructure:"refreshTTL"`
}

// SwaggerConfig describes docs settings.
type SwaggerConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// Address returns the HTTP bind address.
func (c HTTPConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// UploadPath returns the asset upload directory.
func (c AppConfig) UploadPath() string {
	if strings.TrimSpace(c.UploadDir) == "" {
		return "./storage/uploads"
	}
	return c.UploadDir
}

// Load loads config from APP_CONFIG or the default local config.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(getConfigPath())
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
