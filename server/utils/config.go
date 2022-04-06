package utils

import "github.com/spf13/viper"

type AppEnv struct {
	AppEnv          string `mapstructure:"APP_ENV"`
	AppPort         string `mapstructure:"APP_PORT"`
	MongodbUrl      string `mapstructure:"MONGODB_URL"`
	MongodbDatabase string `mapstructure:"MONGODB_DATABASE"`
	LogPath         string `mapstructure:"LOG_PATH"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadAppEnv() (appEnv AppEnv, err error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&appEnv)
	return
}
