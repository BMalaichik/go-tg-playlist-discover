package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

// Get -
func Get(key string) string {
	return viper.GetString(key)
}

// Set -
func Set(key string, value interface{}) {
	viper.Set(key, value)
}

var env string
var envConfigPath string
var defaultEnv = "development"

func init() {
	botAPIKey := os.Getenv("BOT_API_KEY")
	env := os.Getenv("ENV")

	if env == "" {
		env = defaultEnv
	}

	envConfigPath = "./configs/" + env

	viper.SetConfigType("json")
	viper.Set(ENV, env)
	viper.Set(BotAPIKey, botAPIKey)

	viper.AddConfigPath(envConfigPath)
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatal("Configuration initialization failed", err)
	}
}
