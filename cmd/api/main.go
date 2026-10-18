package main

import (
	"github.com/adrian-kurek/fit_the_meet/common/logger"
	"github.com/adrian-kurek/fit_the_meet/config"
	"github.com/joho/godotenv"
)

func main() {
	logger := logger.Setup()
	err := godotenv.Load()
	if err != nil {
		logger.Error("failed to load environment variables", "err:", err.Error())
		panic(err)
	}

	db, err := config.ConnectToDB()
	if err != nil {
		logger.Error("failed to connect to databse service", "err:", err.Error())
		panic(err)
	}

	defer db.Close()

	cacheService, err := config.ConnectToCache()
	if err != nil {
		logger.Error("failed to connect to cache service", "err:", err.Error())
		panic(err)
	}
	defer cacheService.Close()

	logger.Info("started successfully")
}
