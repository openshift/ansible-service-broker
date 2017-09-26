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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package app

import (
	"errors"
	"io"
	"os"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/config"
)

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
	logFile := c.GetString("log.logfile")
	stdOut := c.GetBool("log.stdout")
	level := c.GetString("log.level")
	color := c.GetBool("log.color")

	if logFile == "" && !stdOut {
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

	if logFile != "" {
		var lFile *os.File

		if _, err = os.Stat(logFile); os.IsNotExist(err) {
			if lFile, err = os.Create(logFile); err != nil {
				lFile.Close()
				return nil, err
			}
		} else {
			if lFile, err = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0666); err != nil {
				lFile.Close()
				return nil, err
			}
		}

		log.file = lFile
		backends = append(backends, formattedBackend(lFile, false))
	}

	if stdOut {
		backends = append(backends, formattedBackend(os.Stdout, color))
	}

	multiBackend := logging.MultiLogger(backends...)
	logger.SetBackend(multiBackend)
	logging.SetLevel(levelFromString(level), MODULE)
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
