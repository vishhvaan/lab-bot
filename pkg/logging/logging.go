package logging

import (
	"errors"

	"go.uber.org/zap"
)

const logFolder = "logs"

// adapted from from zap.NewProductionConfig()
func newConfig(fileName string) zap.Config {
	return zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout", logFolder + fileName},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func loggerGenerator(fileName string) (logger *zap.Logger, err error) {
	cfg := newConfig(fileName)

	logger, err = cfg.Build()
	if err != nil {
		return nil, err
	}
	defer logger.Sync()

	return logger, nil
}

func StartGlobalLogger() (err error) {
	logger, err := loggerGenerator("main")
	if err != nil {
		return err
	}

	undo := zap.ReplaceGlobals(logger)
	defer undo()
	return nil
}

func CreateLogger(fileName string) (logger *zap.Logger, err error) {
	if fileName == "main" {
		return nil, errors.New("cannot create logger with name \"main\"")
	}

	return loggerGenerator(fileName)
}
