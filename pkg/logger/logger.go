package logger

import (
	"go.uber.org/zap"
)

// New creates a new logger instance
func New(environment string) *zap.Logger {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
		config.DisableStacktrace = true
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Customize the configuration
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}

	return logger
}
