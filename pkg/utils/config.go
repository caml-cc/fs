package utils

import (
	"errors"
	"fs/internal/models"
	"os"

	"github.com/joho/godotenv"
)

var Conf models.Config

func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	Conf = models.Config{
		PORT:    os.Getenv("PORT"),
		API_KEY: os.Getenv("API_KEY"),
	}

	if Conf.API_KEY == "" || Conf.PORT == "" {
		return errors.New("env file is not filled in correctly")
	}

	return nil
}

func fallback(value, def string) string {
	if value == "" {
		return def
	}
	return value
}
