package logger

import (
	"fmt"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"

	"github.com/sirupsen/logrus"
)

func TestLevelFromString(t *testing.T) {
	testCases := []struct {
		LogLevel string
		Level    logrus.Level
	}{
		{
			LogLevel: "panic",
			Level:    logrus.PanicLevel,
		}, {
			LogLevel: "critical",
			Level:    logrus.FatalLevel,
		}, {
			LogLevel: "fatal",
			Level:    logrus.FatalLevel,
		}, {
			LogLevel: "error",
			Level:    logrus.ErrorLevel,
		}, {
			LogLevel: "warning",
			Level:    logrus.WarnLevel,
		}, {
			LogLevel: "warn",
			Level:    logrus.WarnLevel,
		}, {
			LogLevel: "notice",
			Level:    logrus.InfoLevel,
		}, {
			LogLevel: "info",
			Level:    logrus.InfoLevel,
		}, {
			LogLevel: "debug",
			Level:    logrus.DebugLevel,
		}, {
			LogLevel: "nothing",
			Level:    logrus.DebugLevel,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("test - %v", tc.LogLevel), func(t *testing.T) {
			ft.AssertEqual(t, tc.Level, levelFromString(tc.LogLevel))
		})
	}
}

func TestInitializeLog(t *testing.T) {
	testCases := []struct {
		Config      LogConfig
		ShouldError bool
		Name        string
	}{
		{
			Config: LogConfig{
				LogFile: "/tmp/ansible-service-broker-asb.log",
				Stdout:  true,
				Level:   "debug",
				Color:   true,
			},
			ShouldError: false,
			Name:        "logDoesNotExist",
		},
		{
			Config: LogConfig{
				LogFile: "testdata/testing.log",
				Stdout:  false,
				Level:   "debug",
				Color:   false,
			},
			ShouldError: false,
			Name:        "logDoesExist",
		},
		{
			Config: LogConfig{
				LogFile: "testdata/testing.log",
				Stdout:  false,
				Level:   "critical",
				Color:   false,
			},
			ShouldError: false,
			Name:        "logIsCritical",
		},
		{
			Config: LogConfig{
				LogFile: "",
				Stdout:  false,
			},
			ShouldError: true,
			Name:        "logShouldFail",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("test - %v", tc.Name), func(t *testing.T) {
			err := InitializeLog(tc.Config)
			if err != nil && !tc.ShouldError {
				t.Fail()
			}
			if err == nil && tc.ShouldError {
				t.Fail()
			}
		})
	}
}
