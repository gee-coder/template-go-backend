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
	Storage  StorageConfig  `mapstructure:"storage"`
	SMS      SMSConfig      `mapstructure:"sms"`
	Swagger  SwaggerConfig  `mapstructure:"swagger"`
}

// AppConfig describes global application settings.
type AppConfig struct {
	Name      string `mapstructure:"name"`
	Env       string `mapstructure:"env"`
	Debug     bool   `mapstructure:"debug"`
	UploadDir string `mapstructure:"uploadDir"`
}

// StorageConfig describes object storage settings.
type StorageConfig struct {
	Provider  string             `mapstructure:"provider"`
	Local     LocalStorageConfig `mapstructure:"local"`
	MinIO     MinIOStorageConfig `mapstructure:"minio"`
	AliyunOSS AliyunOSSConfig    `mapstructure:"aliyunOSS"`
	HuaweiOBS HuaweiOBSConfig    `mapstructure:"huaweiOBS"`
}

// LocalStorageConfig describes local filesystem storage settings.
type LocalStorageConfig struct {
	PublicBaseURL string `mapstructure:"publicBaseURL"`
}

// MinIOStorageConfig describes MinIO object storage settings.
type MinIOStorageConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"accessKeyID"`
	AccessKeySecret string `mapstructure:"accessKeySecret"`
	Bucket          string `mapstructure:"bucket"`
	UseSSL          bool   `mapstructure:"useSSL"`
	PublicBaseURL   string `mapstructure:"publicBaseURL"`
	RootPath        string `mapstructure:"rootPath"`
}

// AliyunOSSConfig describes Aliyun OSS settings.
type AliyunOSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"accessKeyID"`
	AccessKeySecret string `mapstructure:"accessKeySecret"`
	Bucket          string `mapstructure:"bucket"`
	PublicBaseURL   string `mapstructure:"publicBaseURL"`
	RootPath        string `mapstructure:"rootPath"`
}

// HuaweiOBSConfig describes Huawei OBS settings.
type HuaweiOBSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"accessKeyID"`
	AccessKeySecret string `mapstructure:"accessKeySecret"`
	Bucket          string `mapstructure:"bucket"`
	PublicBaseURL   string `mapstructure:"publicBaseURL"`
	RootPath        string `mapstructure:"rootPath"`
}

// SMSConfig describes SMS verification settings.
type SMSConfig struct {
	Provider string          `mapstructure:"provider"`
	CodeTTL  time.Duration   `mapstructure:"codeTTL"`
	Cooldown time.Duration   `mapstructure:"cooldown"`
	Mock     MockSMSConfig   `mapstructure:"mock"`
	Aliyun   AliyunSMSConfig `mapstructure:"aliyun"`
	Huawei   HuaweiSMSConfig `mapstructure:"huawei"`
}

// MockSMSConfig describes the built-in mock SMS provider.
type MockSMSConfig struct {
	RevealCode bool   `mapstructure:"revealCode"`
	FixedCode  string `mapstructure:"fixedCode"`
}

// AliyunSMSConfig reserves Aliyun SMS settings for future provider switching.
type AliyunSMSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"accessKeyID"`
	AccessKeySecret string `mapstructure:"accessKeySecret"`
	SignName        string `mapstructure:"signName"`
	TemplateCode    string `mapstructure:"templateCode"`
}

// HuaweiSMSConfig reserves Huawei Cloud SMS settings for future provider switching.
type HuaweiSMSConfig struct {
	Endpoint   string `mapstructure:"endpoint"`
	AppKey     string `mapstructure:"appKey"`
	AppSecret  string `mapstructure:"appSecret"`
	Sender     string `mapstructure:"sender"`
	TemplateID string `mapstructure:"templateID"`
	Signature  string `mapstructure:"signature"`
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

// ResolvedProvider returns the configured storage provider.
func (c StorageConfig) ResolvedProvider() string {
	provider := strings.ToLower(strings.TrimSpace(c.Provider))
	if provider == "" {
		return "minio"
	}
	return provider
}

// ResolvedCodeTTL returns the configured SMS code TTL.
func (c SMSConfig) ResolvedCodeTTL() time.Duration {
	if c.CodeTTL <= 0 {
		return 5 * time.Minute
	}
	return c.CodeTTL
}

// ResolvedCooldown returns the configured SMS resend cooldown.
func (c SMSConfig) ResolvedCooldown() time.Duration {
	if c.Cooldown <= 0 {
		return time.Minute
	}
	return c.Cooldown
}

// ResolvedProvider returns the configured SMS provider.
func (c SMSConfig) ResolvedProvider() string {
	provider := strings.ToLower(strings.TrimSpace(c.Provider))
	if provider == "" {
		return "mock"
	}
	return provider
}

// Load loads config from APP_CONFIG or the default local config.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(getConfigPath())
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
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
