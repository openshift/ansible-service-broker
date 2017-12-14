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

package logger

import (
	"errors"
	"io"
	"os"

	logging "github.com/op/go-logging"
)

// LogConfig - The configuration for the logging.
type LogConfig struct {
	LogFile string
	Stdout  bool
	Level   string
	Color   bool
}

var backends []logging.Backend
var logLevel logging.Level

//InitializeLog - initilize the logging utility
func InitializeLog(config LogConfig) error {
	if config.LogFile == "" && !config.Stdout {
		return errors.New("Cannot have a blank logfile and not log to stdout")
	}

	logLevel = levelFromString(config.Level)

	colorFormatter := logging.MustStringFormatter(
		"%{color}[%{time}] [%{level}] - %{message}%{color:reset}",
	)

	standardFormatter := logging.MustStringFormatter(
		"[%{time}] [%{level}] - %{message}",
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

		if _, err := os.Stat(config.LogFile); os.IsNotExist(err) {
			if logFile, err = os.Create(config.LogFile); err != nil {
				logFile.Close()
				return err
			}
		} else {
			if logFile, err = os.OpenFile(config.LogFile, os.O_APPEND|os.O_WRONLY, 0666); err != nil {
				logFile.Close()
				return err
			}
		}
		b := formattedBackend(logFile, false)
		bl := logging.AddModuleLevel(b)
		bl.SetLevel(logLevel, "ASB")
		backends = append(backends, bl)
	}

	if config.Stdout {
		b := formattedBackend(os.Stdout, config.Color)
		bl := logging.AddModuleLevel(b)
		bl.SetLevel(logLevel, "ASB")
		backends = append(backends, bl)
	}
	logging.SetBackend(backends...)
	return nil
}

// NewLog - Creates a new logging object for the module provided.
func NewLog() *logging.Logger {
	return logging.MustGetLogger("ASB")
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
