package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	Port           string
	Env            AppEnv
	PostgresConfig *PostgresConfig
}

// PostgresConfig holds the PostgreSQL database configuration.
type PostgresConfig struct {
	ConnStr string
}

// AppEnv represents the application environment.
type AppEnv string

const (
	AppEnvLocal       AppEnv = "local"
	AppEnvDevelopment AppEnv = "development"
	AppEnvProduction  AppEnv = "production"
)

var (
	allowedEnv = map[AppEnv]bool{
		AppEnvLocal:       true,
		AppEnvDevelopment: true,
	}
	envFilePath = "config/config.%s.yml"
)

// GetEnv retrieves the application environment from the APP_ENV environment variable.
// Defaults to "local" if not set or invalid.
func GetEnv() AppEnv {
	appEnv := AppEnv(os.Getenv("APP_ENV"))
	if _, ok := allowedEnv[appEnv]; ok {
		return appEnv
	}
	return AppEnvLocal
}

// NewConfig creates a new Config instance based on the provided environment.
func NewConfig(env AppEnv) *Config {
	return &Config{
		Port: viper.GetString("server.port"),
		Env:  env,
		PostgresConfig: &PostgresConfig{
			ConnStr: viper.GetString("postgres.connstr"),
		},
	}
}

// InitViper initializes Viper to read the configuration file based on the environment.
func InitViper(env AppEnv) error {
	path := fmt.Sprintf(envFilePath, env)

	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return errors.New("failed to read config")
	}

	return nil
}
