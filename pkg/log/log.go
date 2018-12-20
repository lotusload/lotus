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
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DefaultLevel    = "info"
	DefaultEncoding = "json"
)

var (
	DefaultConfigs = Configs{
		Level:    DefaultLevel,
		Encoding: DefaultEncoding,
	}
)

type Configs struct {
	Level          string
	Encoding       string
	ServiceContext *ServiceContext
}

func NewLogger(c Configs) (*zap.Logger, error) {
	level := new(zapcore.Level)
	if err := level.Set(c.Level); err != nil {
		return nil, err
	}
	var options []zap.Option
	if c.ServiceContext != nil {
		options = []zap.Option{
			zap.Fields(zap.Object("serviceContext", c.ServiceContext)),
		}
	}
	logger, err := newConfig(*level, c.Encoding).Build(options...)
	if err != nil {
		return nil, err
	}
	return logger.Named(c.ServiceContext.Service), nil
}

func newConfig(level zapcore.Level, encoding string) zap.Config {
	return zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         encoding,
		EncoderConfig:    newEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func newEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "eventTime",
		LevelKey:       "severity",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    encodeLevel,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func encodeLevel(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString("DEBUG")
	case zapcore.InfoLevel:
		enc.AppendString("INFO")
	case zapcore.WarnLevel:
		enc.AppendString("WARNING")
	case zapcore.ErrorLevel:
		enc.AppendString("ERROR")
	case zapcore.DPanicLevel:
		enc.AppendString("CRITICAL")
	case zapcore.PanicLevel:
		enc.AppendString("ALERT")
	case zapcore.FatalLevel:
		enc.AppendString("EMERGENCY")
	}
}

type ServiceContext struct {
	Service string
	Version string
}

func (sc ServiceContext) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if sc.Service == "" {
		return errors.New("service name is mandatory")
	}
	enc.AddString("service", sc.Service)
	enc.AddString("version", sc.Version)
	return nil
}
