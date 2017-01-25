package app

import (
	"errors"
	"github.com/op/go-logging"
	"io"
	"os"
)

type LogConfig struct {
	LogFile string
	Stdout  bool
	Level   string
	Color   bool
}

type Log struct {
	*logging.Logger
	file *os.File
}

const MODULE = "asb"

// TODO: Consider no output?
func NewLog(config LogConfig) (*Log, error) {
	var err error

	if config.LogFile == "" && !config.Stdout {
		return nil, errors.New("Cannot have a blank logfile and not log to stdout")
	}

	// TODO: More validation? Check file is good?
	// TODO: Validate level is actually possible?

	log := &Log{}

	logger := logging.MustGetLogger(MODULE)
	var backends []logging.Backend

	colorFormatter := logging.MustStringFormatter(
		"%{color}[%{time}] [%{level}] %{message}%{color:reset}",
	)

	standardFormatter := logging.MustStringFormatter(
		"[%{time}] [%{level}] %{message}",
	)

	var formattedBackend = func(writer io.Writer, isColored bool) logging.Backend {
		backend := logging.NewLogBackend(writer, "", 0)
		formatter := standardFormatter
		if isColored {
			formatter = colorFormatter
		}
		return logging.NewBackendFormatter(backend, formatter)
	}

	if config.LogFile != "" {
		var logFile *os.File

		if _, err = os.Stat(config.LogFile); os.IsNotExist(err) {
			if logFile, err = os.Create(config.LogFile); err != nil {
				logFile.Close()
				return nil, err
			}
		} else {
			if logFile, err = os.OpenFile(config.LogFile, os.O_APPEND|os.O_WRONLY, 0666); err != nil {
				logFile.Close()
				return nil, err
			}
		}

		log.file = logFile
		backends = append(backends, formattedBackend(logFile, false))
	}

	if config.Stdout {
		backends = append(backends, formattedBackend(os.Stdout, config.Color))
	}

	multiBackend := logging.MultiLogger(backends...)
	logger.SetBackend(multiBackend)
	logging.SetLevel(levelFromString(config.Level), MODULE)
	log.Logger = logger

	return log, nil
}

func levelFromString(str string) logging.Level {
	var level logging.Level

	switch str {
	case "critical":
		level = logging.CRITICAL
	case "error":
		level = logging.ERROR
	case "warning":
		level = logging.WARNING
	case "notice":
		level = logging.NOTICE
	case "info":
		level = logging.INFO
	default:
		level = logging.DEBUG
	}

	return level
}
