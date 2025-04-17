package config_test

import (
	"address-validator/config"
	"reflect"
	"testing"
)

func TestConfig_NewLoggerConfig(t *testing.T) {
	const (
		LEVEL       = "LEVEL"
		ENCODING    = "ENCODING"
		OUTPUT_PATH = "OUTPUT_PATH"
		ERROR_PATH  = "ERROR_PATH"
	)

	type args struct {
		environment config.Environment
	}
	tests := []struct {
		name string
		env  [][2]string
		args args
		want config.LoggerConfig
	}{
		{
			name: "Test Returns Default",
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "stdout",
				ErrorPath:     "stderr",
				IsDevelopment: false,
			},
		},
		{
			name: "Test invalid Log Level Returns Default",
			env:  [][2]string{{LEVEL, "stdout"}},
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "stdout",
				ErrorPath:     "stderr",
				IsDevelopment: false,
			},
		},
		{
			name: "Test lowercase Log Level Returns Level",
			env:  [][2]string{{LEVEL, "debug"}},
			want: config.LoggerConfig{
				Level:         "debug",
				Encoding:      "json",
				OutputPath:    "stdout",
				ErrorPath:     "stderr",
				IsDevelopment: false,
			},
		},
		{
			name: "Test UPPERCASE Log Level Returns Level",
			env:  [][2]string{{LEVEL, "DEBUG"}},
			want: config.LoggerConfig{
				Level:         "DEBUG",
				Encoding:      "json",
				OutputPath:    "stdout",
				ErrorPath:     "stderr",
				IsDevelopment: false,
			},
		},
		{
			name: "Test Invalid Log Encoding Returns Default",
			env:  [][2]string{{ENCODING, "JSON"}},
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "stdout",
				ErrorPath:     "stderr",
				IsDevelopment: false,
			},
		},
		{
			name: "Test Log Encoding Returns Encoding",
			env:  [][2]string{{ENCODING, "json"}},
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "stdout",
				ErrorPath:     "stderr",
				IsDevelopment: false,
			},
		},
		{
			name: "Test Log Paths Returns Unix Path",
			env:  [][2]string{{OUTPUT_PATH, "/var/log/app.log"}, {ERROR_PATH, "/var/errors/app.log"}},
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "/var/log/app.log",
				ErrorPath:     "/var/errors/app.log",
				IsDevelopment: false,
			},
		},
		{
			name: "Test Log Paths Returns Window Path",
			env:  [][2]string{{OUTPUT_PATH, "C:\\Logs\\app.json"}, {ERROR_PATH, "C:\\Errors\\app.json"}},
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "C:\\Logs\\app.json",
				ErrorPath:     "C:\\Errors\\app.json",
				IsDevelopment: false,
			},
		},
		{
			name: "Test Log Paths Returns Default Path",
			env:  [][2]string{{OUTPUT_PATH, "stdout"}, {ERROR_PATH, "stderr"}},
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "stdout",
				ErrorPath:     "stderr",
				IsDevelopment: false,
			},
		},
		{
			name: "Test Log Paths Returns Cloud Service Path",
			env:  [][2]string{{OUTPUT_PATH, "cloudwatch://prod/logs"}, {ERROR_PATH, "cloudwatch://prod/errors"}},
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "cloudwatch://prod/logs",
				ErrorPath:     "cloudwatch://prod/errors",
				IsDevelopment: false,
			},
		},
		{
			name: "Test Log Paths Returns Generic Protocol Path",
			env:  [][2]string{{OUTPUT_PATH, "custom://host:1234/path"}, {ERROR_PATH, "custom://host:1234/errors"}},
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "custom://host:1234/path",
				ErrorPath:     "custom://host:1234/errors",
				IsDevelopment: false,
			},
		},
		{
			name: "Test Relative Log Paths Returns Default",
			env:  [][2]string{{OUTPUT_PATH, "/tmp/../../../etc/passwd"}, {ERROR_PATH, "/tmp/../../../etc/passwd"}},
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "stdout",
				ErrorPath:     "stderr",
				IsDevelopment: false,
			},
		},
		{
			name: "Test Invalid Log Paths Returns Window Path",
			env:  [][2]string{{OUTPUT_PATH, "no_protocol"}, {ERROR_PATH, "no_protocol"}},
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "stdout",
				ErrorPath:     "stderr",
				IsDevelopment: false,
			},
		},
		{
			name: "Test Development Returns True",
			args: args{environment: config.ENV_DEVELOPMENT},
			want: config.LoggerConfig{
				Level:         "info",
				Encoding:      "json",
				OutputPath:    "stdout",
				ErrorPath:     "stderr",
				IsDevelopment: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Environment variables automatically cleans up after each test
			for _, pair := range tt.env {
				t.Setenv(pair[0], pair[1])
			}

			c := config.Config{}
			if got := c.NewLoggerConfig(tt.args.environment); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.NewLoggerConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
