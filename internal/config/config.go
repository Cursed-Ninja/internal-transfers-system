package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Port           string
	Env            AppEnv
	PostgresConfig *PostgresConfig
}

type PostgresConfig struct {
	Database string
	Port     string
	Username string
	Password string
}

type AppEnv string

const (
	AppEnvLocal       AppEnv = "local"
	AppEnvDevelopment AppEnv = "development"
)

var (
	allowedEnv = map[AppEnv]bool{
		AppEnvLocal:       true,
		AppEnvDevelopment: true,
	}
	envFilePath = "config/config.%s.yml"
)

func GetEnv() AppEnv {
	appEnv := AppEnv(os.Getenv("APP_ENV"))
	if _, ok := allowedEnv[appEnv]; ok {
		return appEnv
	}
	return AppEnvLocal
}

func NewConfig(env AppEnv) *Config {
	return &Config{
		Port: viper.GetString("server.port"),
		Env:  env,
		PostgresConfig: &PostgresConfig{
			Database: viper.GetString("postgres.db"),
			Username: viper.GetString("postgres.username"),
			Password: viper.GetString("postgres.password"),
			Port:     viper.GetString("postgres.port"),
		},
	}
}

func InitViper(env AppEnv) error {
	path := fmt.Sprintf(envFilePath, env)

	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return errors.New("failed to read config")
	}

	return nil
}
