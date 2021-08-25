package logging

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

const logFolder = "logs"

func loggerGenerator(fileName string) (logger *zap.Logger, err error) {

	// See the documentation for Config and zapcore.EncoderConfig for all the
	// available options.
	configJSON := fmt.Sprintf(`{
		"level": "debug",
		"encoding": "json",
		"outputPaths": ["stdout", %v/main],
		"errorOutputPaths": ["stderr"],
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "levelEncoder": "lowercase"
		}
	  }`, logFolder)
	globalZapConfig := []byte(configJSON)

	var cfg zap.Config
	if err := json.Unmarshal(globalZapConfig, &cfg); err != nil {
		return nil, err
	}

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
