package log

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLoggerOK(t *testing.T) {
	validLevels := []string{
		"debug",
		"info",
		"warn",
		"error",
		"dpanic",
		"panic",
		"fatal",
	}
	validEncodings := []string{
		"json",
		"console",
	}
	for _, level := range validLevels {
		for _, encoding := range validEncodings {
			cfg := Configs{
				Level:    level,
				Encoding: encoding,
				ServiceContext: &ServiceContext{
					Service: "test-service",
					Version: "1.0.0",
				},
			}
			logger, err := NewLogger(cfg)
			des := fmt.Sprintf("level: %s, encoding: %s", level, encoding)
			assert.Nil(t, err, des)
			assert.NotNil(t, logger, des)
		}
	}
}

func TestNewLoggerFailed(t *testing.T) {
	configs := []Configs{
		Configs{
			Level:    "foo",
			Encoding: "json",
		},
		Configs{
			Level:    "info",
			Encoding: "foo",
		},
	}
	for _, cfg := range configs {
		logger, err := NewLogger(cfg)
		des := fmt.Sprintf("level: %s, encoding: %s", cfg.Level, cfg.Encoding)
		assert.NotNil(t, err, des)
		assert.Nil(t, logger, des)
	}
}
