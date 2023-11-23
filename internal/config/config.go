package config

import "github.com/spf13/viper"

type Config struct {
	Environment     string `mapstructure:"ENVIRONMENT"`
	ServerAddress   string `mapstructure:"SERVER_ADDRESS"`
	LogLevel        string `mapstructure:"LOG_LEVEL"`
	LogFormat       string `mapstructure:"LOG_FORMAT"`
	PodNameSpace    string `mapstructure:"POD_NAMESPACE"`
	PodAPP          string `mapstructure:"APP_NAME"`
	TracingExporter string `mapstructure:"USE_TRACING_EXPORTER"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
