package utils

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

type AppEnv struct {
	AppName      string `mapstructure:"APP_NAME"`
	AppHost      string `mapstructure:"APP_HOST"`
	AppPort      string `mapstructure:"APP_PORT"`
	DbUser       string `mapstructure:"DB_USER"`
	DbPassword   string `mapstructure:"DB_PASSWORD"`
	DbHost       string `mapstructure:"DB_HOST"`
	DbPort       string `mapstructure:"DB_PORT"`
	DbDatabase   string `mapstructure:"DB_DATABASE"`
	DbAuthSource string `mapstructure:"DB_AUTH_SOURCE"`
	LogPath      string `mapstructure:"LOG_PATH"`
}

var appEnv AppEnv

// LoadConfig reads configuration from file or environment variables.
func BootstrapEnv() error {
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("BootstrapEnv ReadInConfig Error: %w", err)
	}

	if err := viper.Unmarshal(&appEnv); err != nil {
		return fmt.Errorf("BootstrapEnv Unmarshal Error: %w", err)
	}

	return nil
}

func GetAppEnv() AppEnv {
	return appEnv
}

func GetEnvOrDefault(field string, defaultValue reflect.Value) reflect.Value {
	r := reflect.ValueOf(appEnv)
	f := reflect.Indirect(r).FieldByName(field)

	if f.IsZero() {
		return defaultValue
	}

	return f
}
