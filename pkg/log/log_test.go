// Copyright (c) 2018 Lotus Load
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
