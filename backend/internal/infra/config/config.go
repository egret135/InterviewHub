package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Seed     SeedConfig     `mapstructure:"seed"`
	DeepSeek DeepSeekConfig `mapstructure:"deepseek"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type SeedConfig struct {
	ContentFile string `mapstructure:"content_file"`
}

type DeepSeekConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	Delay   string `mapstructure:"delay"`
	Retries int    `mapstructure:"retries"`
}

func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return envConfig(), nil
	}
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func envConfig() *Config {
	return &Config{
		Server: ServerConfig{Port: getEnvInt("SERVER_PORT", "17001")},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "interview_hub"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Seed: SeedConfig{
			ContentFile: getEnv("SEED_FILE", "../resource-01/content.md"),
		},
		DeepSeek: DeepSeekConfig{
			APIKey:  getEnv("DEEPSEEK_API_KEY", ""),
			Model:   getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
			Delay:   getEnv("DEEPSEEK_DELAY", "2s"),
			Retries: getEnvInt("DEEPSEEK_RETRIES", "3"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key, fallback string) int {
	s := os.Getenv(key)
	if s == "" {
		s = fallback
	}
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
