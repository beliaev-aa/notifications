package main

import (
	"github.com/beliaev-aa/notifications/internal/app"
	"github.com/beliaev-aa/notifications/internal/utils"
	"time"
)

func main() {
	logger := utils.InitLogger()

	cfg := &app.Config{
		HTTPAddr:        ":3000",
		ShutdownTimeout: 5 * time.Second,
	}

	application := app.NewApp(cfg, logger)

	if err := application.Run(); err != nil {
		logger.WithError(err).Fatal("Error running notification server")
	}
}
