//
// Copyright (c) 2018 Red Hat, Inc.
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
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

// LogConfig - The configuration for the logging.
type LogConfig struct {
	LogFile string
	Stdout  bool
	Level   string
	Color   bool
}

var logLevel logrus.Level

// InitializeLog - initilize the logrus utility
func InitializeLog(config LogConfig) error {
	if config.LogFile == "" && !config.Stdout {
		return errors.New("Cannot have a blank logfile and not log to stdout")
	}

	logLevel = levelFromString(config.Level)

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
	}

	if config.Stdout {
		log.SetOutput(os.Stdout)
	}
	return nil
}

func levelFromString(str string) logrus.Level {
	var level logrus.Level

	switch str {
	case "panic":
		level = logrus.PanicLevel
	case "critical", "fatal":
		level = logrus.FatalLevel
	case "error":
		level = logrus.ErrorLevel
	case "warning", "warn":
		level = logrus.WarnLevel
	case "notice", "info":
		level = logrus.InfoLevel
	default:
		level = logrus.DebugLevel
	}

	return level
}
