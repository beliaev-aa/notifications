package main

import (
	"github.com/beliaev-aa/notifications/internal/app"
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/internal/utils"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := utils.InitLogger()

	// Загружаем конфигурацию из YAML и ENV
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Проверяем и обновляем уровень логирования из конфигурации, если он отличается
	if cfg.Logger.Level != "" {
		configLogLevel, errParseLevel := logrus.ParseLevel(cfg.Logger.Level)
		if errParseLevel != nil {
			logger.WithError(errParseLevel).WithFields(logrus.Fields{
				"level":        cfg.Logger.Level,
				"usedLogLevel": logger.Level,
			}).Info("Failed to parse log level from configuration")
		} else {
			currentLevel := logger.GetLevel()
			if currentLevel != configLogLevel {
				logger.SetLevel(configLogLevel)
				logger.WithFields(logrus.Fields{
					"previousLevel": currentLevel.String(),
					"newLevel":      configLogLevel.String(),
				}).Info("Logger level updated from configuration")
			}
		}
	}

	application := app.NewApp(cfg, logger)

	if err = application.Run(); err != nil {
		logger.WithError(err).Fatal("Error running notification server")
	}
}
