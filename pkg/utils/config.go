package utils

import (
	"errors"
	"fs/internal/models"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var Conf models.Config

func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	maxStorageInt, err := strconv.Atoi(os.Getenv("MAX_STORAGE"))
	if err != nil {
		return err
	}

	keepInt, err := strconv.Atoi(os.Getenv("KEEP"))
	if err != nil {
		return err
	}

	Conf = models.Config{
		PORT:        os.Getenv("PORT"),
		API_KEY:     os.Getenv("API_KEY"),
		MAX_STORAGE: maxStorageInt,
		KEEP:        keepInt,
	}

	if Conf.API_KEY == "" || Conf.PORT == "" || Conf.MAX_STORAGE <= 0 || Conf.KEEP <= 0 {
		return errors.New("env file is not filled in correctly")
	}

	return nil
}
