package main

import (
	"fmt"
	"fs/api"
	"fs/internal/database"
	"fs/internal/store"
	"fs/pkg/utils"
	"os"
	"time"
)

func main() {
	err := utils.LoadConfig()
	if err != nil {
		println("ERROR: " + err.Error())
		os.Exit(1)
	}

	err = database.InitSQLiteDB()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	store.StartExpiryCleanupWorker(time.Minute)

	api.StartServer(utils.Conf)

	database.DbClose()
}
