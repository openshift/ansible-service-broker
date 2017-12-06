//
// Copyright (c) 2017 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package app

import (
	"errors"
	"io"
	"os"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/config"
)

// LogConfig - The configuration for the logging.
type LogConfig struct {
	LogFile string
	Stdout  bool
	Level   string
	Color   bool
}

// Log - Logging struct that will contain https://godoc.org/os#File and the logging object.
type Log struct {
	*logging.Logger
	file *os.File
}

// MODULE - Module for the logger.
const MODULE = "asb"

// NewLog - Creates a new logging object
// TODO: Consider no output?
func NewLog(c *config.Config) (*Log, error) {
	var err error
	if c.Empty() {
		return nil, errors.New("Cannot have a blank logfile and not log to stdout")
	}
	config := LogConfig{
		LogFile: c.GetString("log.logfile"),
		Stdout:  c.GetBool("log.stdout"),
		Level:   c.GetString("log.level"),
		Color:   c.GetBool("log.color"),
	}

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
