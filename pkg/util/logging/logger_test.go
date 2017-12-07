package logger

import (
	"fmt"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"

	logging "github.com/op/go-logging"
)

func TestLevelFromString(t *testing.T) {
	testCases := []struct {
		LogLevel string
		Level    logging.Level
	}{
		{
			LogLevel: "critical",
			Level:    logging.CRITICAL,
		}, {
			LogLevel: "error",
			Level:    logging.ERROR,
		}, {
			LogLevel: "warning",
			Level:    logging.WARNING,
		}, {
			LogLevel: "notice",
			Level:    logging.NOTICE,
		}, {
			LogLevel: "info",
			Level:    logging.INFO,
		}, {
			LogLevel: "debug",
			Level:    logging.DEBUG,
		}, {
			LogLevel: "nothing",
			Level:    logging.DEBUG,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("test - %v", tc.LogLevel), func(t *testing.T) {
			ft.AssertEqual(t, tc.Level, levelFromString(tc.LogLevel))
		})
	}
}

func TestNewLog(t *testing.T) {
	ft.AssertNotNil(t, NewLog())
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
