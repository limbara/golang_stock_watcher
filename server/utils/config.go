package utils

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type AppEnv struct {
	AppName         string `mapstructure:"APP_NAME"`
	AppHost         string `mapstructure:"APP_HOST" validate:"required"`
	AppPort         string `mapstructure:"APP_PORT" validate:"required"`
	AppTimezone     string `mapstructure:"APP_TZ" validate:"required"`
	MongoDbUri      string `mapstructure:"MONGODB_URI" validate:"required"`
	MongoDbDatabase string `mapstructure:"MONGODB_DATABASE" validate:"required"`
	LogPath         string `mapstructure:"LOG_PATH"`
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

	validate := validator.New()

	if err := validate.Struct(appEnv); err != nil {
		return fmt.Errorf("Error validate NewDbConfig : %w", err)
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
