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
