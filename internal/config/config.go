package config

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

type ServerType string

const (
	ServerTypeAsync ServerType = "ASYNC"
	ServerTypeSync  ServerType = "SYNC"
)

type Config struct {
	ServiceName string
	AppEnv      string     `mapstructure:"APP_ENV"`
	Host        string     `mapstructure:"HOST"`
	Port        int        `mapstructure:"PORT"`
	MaxClients  int        `mapstructure:"MAX_CLIENTS"`
	ServerType  ServerType `mapstructure:"SERVER_TYPE"`
}

func NewConfig(path string) (*Config, error) {
	config := &Config{
		ServiceName: "memodb",
	}
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("MAX_CLIENTS", 100)
	viper.SetDefault("SERVER_TYPE", ServerTypeAsync)

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	if config.MaxClients > 10000 {
		return nil, errors.New("max clients cannot be more than 10000")
	}

	if config.AppEnv == "development" {
		log.Println("The App is running in development env")
	}

	if config.ServerType == ServerTypeAsync {
		log.Println("The server is running in async mode : on single thread ")
	} else {
		log.Println("The server is running in sync mode : on multiple threads")
	}

	return config, nil
}
