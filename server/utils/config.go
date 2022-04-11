package utils

import (
	"github.com/spf13/viper"
)

// LoadConfig reads configuration from file or environment variables.
func BootstrapEnv() {
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	// ignoring error on read file to let validate handle it instead
	// because in production we won't have an env file
	viper.ReadInConfig()

	viper.SetDefault("LOG_PATH", "./storage/error")
	viper.SetDefault("APP_TZ", "UTC")
}
