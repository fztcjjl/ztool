package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Database string `json:"database,omitempty"`
}

var config Config

func GetConfig() *Config {
	return &config
}

func Init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Fatal(fmt.Sprintf("Read config file fail: %s", err.Error()))
	}
	for k, v := range viper.AllSettings() {
		viper.SetDefault(k, v)
	}
	config.Host = viper.GetString("host")
	config.Port = viper.GetInt("port")
	config.User = viper.GetString("user")
	config.Password = viper.GetString("password")
	config.Database = viper.GetString("database")
}
