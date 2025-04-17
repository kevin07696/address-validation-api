package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Level         string `json:"level" yaml:"level"`           // debug, info, warn, error, dpanic, panic, fatal
	Encoding      string `json:"encoding" yaml:"encoding"`     // json or console
	OutputPath    string `json:"outputPath" yaml:"outputPath"` // stdout, stderr, or file path
	ErrorPath     string `json:"errorPath" yaml:"errorPath"`   // separate path for error logs
	IsDevelopment bool   `json:"development" yaml:"development"`
}

func NewLogger(config LoggerConfig) (*zap.Logger, error) {
	// Set log level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, err
	}

	// Configure encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	if config.IsDevelopment {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		encoderConfig.EncodeTime = zapcore.EpochTimeEncoder
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create encoder based on config
	var encoder zapcore.Encoder
	switch config.Encoding {
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default: // Default to JSON
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create output syncers
	var outputSyncer zapcore.WriteSyncer
	switch config.OutputPath {
	case "", "stdout":
		outputSyncer = zapcore.AddSync(os.Stdout)
	case "stderr":
		outputSyncer = zapcore.AddSync(os.Stderr)
	default:
		file, err := os.OpenFile(config.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		outputSyncer = zapcore.AddSync(file)
	}

	// Create error syncer (defaults to outputSyncer if not specified)
	errorSyncer := outputSyncer
	if config.ErrorPath != "" && config.ErrorPath != config.OutputPath {
		switch config.ErrorPath {
		case "stdout":
			errorSyncer = zapcore.AddSync(os.Stdout)
		case "stderr":
			errorSyncer = zapcore.AddSync(os.Stderr)
		default:
			file, err := os.OpenFile(config.ErrorPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
			errorSyncer = zapcore.AddSync(file)
		}
	}

	// Create cores directly
	core := zapcore.NewTee(
		zapcore.NewCore(
			encoder,
			outputSyncer,
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= level && lvl < zapcore.ErrorLevel
			}),
		),
		zapcore.NewCore(
			encoder,
			errorSyncer,
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.ErrorLevel
			}),
		),
	)

	// Build logger with options
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	if config.IsDevelopment {
		options = append(options, zap.Development())
	}

	return zap.New(core, options...), nil
}

func SugarLogger(cfg LoggerConfig) (*zap.SugaredLogger, error) {
	logger, err := NewLogger(cfg)
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}

func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:         "info",
		Encoding:      "json",
		OutputPath:    "stdout",
		ErrorPath:     "stderr",
		IsDevelopment: false,
	}
}
