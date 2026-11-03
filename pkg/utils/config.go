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

	// MAX_UPLOAD is the largest single request body (in GB, decimals allowed) the
	// server accepts. Larger files must be uploaded in fragments. Defaults to 1.
	maxUploadGB := 1.0
	if raw := os.Getenv("MAX_UPLOAD"); raw != "" {
		maxUploadGB, err = strconv.ParseFloat(raw, 64)
		if err != nil {
			return errors.New("MAX_UPLOAD must be a number (GB)")
		}
	}
	if maxUploadGB <= 0 {
		return errors.New("MAX_UPLOAD must be greater than 0")
	}

	Conf = models.Config{
		PORT:        os.Getenv("PORT"),
		API_KEY:     os.Getenv("API_KEY"),
		MAX_STORAGE: maxStorageInt,
		MAX_UPLOAD:  int64(maxUploadGB * (1 << 30)),
		KEEP:        keepInt,
		UPLOAD_DIR:  os.Getenv("UPLOAD_DIR"),
	}

	if Conf.UPLOAD_DIR == "" {
		Conf.UPLOAD_DIR = "./internal/uploads"
	}

	if Conf.API_KEY == "" || Conf.PORT == "" || Conf.MAX_STORAGE <= 0 || Conf.KEEP <= 0 {
		return errors.New("env file is not filled in correctly")
	}

	return nil
}
