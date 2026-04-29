package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	APP_NAME                         string
	HOST                             string
	PORT                             string
	JWT_SECRET                       string
	DBHOST                           string
	DBPORT                           string
	DBUSER                           string
	DBPASS                           string
	DBNAME                           string
	KafkaBroker                      string
	KafkaTopic                       string
	REDIS_HOST                       string
	REDIS_PORT                       string
	DB_DEVELOPMENT_URL               string
	ENV                              string
	OPEN_FGA_API_URL                 string
	OPEN_FGA_MODEL_DIR               string
	ACCESS_TOKEN_EXPIRATION          int
	REFRESH_TOKEN_EXPIRATION         int
	FORGOT_PASSWORD_TOKEN_EXPIRATION int
}

var cfg *Config

func Load() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg = &Config{
		APP_NAME:                         viper.GetString("APP_NAME"),
		HOST:                             viper.GetString("HOST"),
		PORT:                             viper.GetString("PORT"),
		JWT_SECRET:                       viper.GetString("JWT_SECRET"),
		ACCESS_TOKEN_EXPIRATION:          viper.GetInt("ACCESS_TOKEN_EXPIRATION"),
		REFRESH_TOKEN_EXPIRATION:         viper.GetInt("REFRESH_TOKEN_EXPIRATION"),
		FORGOT_PASSWORD_TOKEN_EXPIRATION: viper.GetInt("FORGOT_PASSWORD_TOKEN_EXPIRATION"),
		DBHOST:                           viper.GetString("DB_HOST"),
		DBPORT:                           viper.GetString("DB_PORT"),
		DBUSER:                           viper.GetString("DB_USER"),
		DBPASS:                           viper.GetString("DB_PASS"),
		DBNAME:                           viper.GetString("DB_NAME"),
		KafkaBroker:                      viper.GetString("KAFKA_BROKERS"),
		KafkaTopic:                       viper.GetString("KAFKA_TOPIC"),
		REDIS_HOST:                       viper.GetString("REDIS_HOST"),
		REDIS_PORT:                       viper.GetString("REDIS_PORT"),
		DB_DEVELOPMENT_URL:               viper.GetString("DB_DEVELOPMENT_URL"),
		ENV:                              viper.GetString("ENV"),
		OPEN_FGA_API_URL:                 viper.GetString("OPEN_FGA_API_URL"),
		OPEN_FGA_MODEL_DIR:               viper.GetString("OPEN_FGA_MODEL_DIR"),
	}

	return cfg, nil
}

func Get() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
