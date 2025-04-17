package config_test

import (
	"address-validator/config"
	"reflect"
	"testing"
)

func TestEnvironment_ToString(t *testing.T) {
	tests := []struct {
		name string
		e    config.Environment
		want string
	}{
		{name: "Test Production constant returns UPPER", e: config.ENV_PRODUCTION, want: "PRODUCTION"},
		{name: "Test Development constant returns UPPER", e: config.ENV_DEVELOPMENT, want: "DEVELOPMENT"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.ToString(); got != tt.want {
				t.Errorf("Environment.ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_NewInfraConfig(t *testing.T) {
	const (
		PORT          = "PORT"
		ENVIRONMENT   = "ENVIRONMENT"
		REQUIRE_HTTPS = "REQUIRE_HTTPS"
	)

	tests := []struct {
		name string
		env  [][2]string
		want config.InfraConfig
	}{
		{
			name: "Test Empty Environment Variables Returns Default Config",
			want: config.InfraConfig{
				Environment:  config.ENV_PRODUCTION,
				Port:         8080,
				IsHttpSecure: true,
			},
		},
		{
			name: "Test Reserved Port at 0 Returns 8080",
			env:  [][2]string{{PORT, "0"}},
			want: config.InfraConfig{
				Environment:  config.ENV_PRODUCTION,
				Port:         8080,
				IsHttpSecure: true,
			},
		},
		{
			name: "Test Blocked Port at 65535 Returns 8080",
			env:  [][2]string{{PORT, "65535"}},
			want: config.InfraConfig{
				Environment:  config.ENV_PRODUCTION,
				Port:         8080,
				IsHttpSecure: true,
			},
		},
		{
			name: "Test Priviledged Port (1-1023) Returns 8080",
			env:  [][2]string{{PORT, "1023"}},
			want: config.InfraConfig{
				Environment:  config.ENV_PRODUCTION,
				Port:         8080,
				IsHttpSecure: true,
			},
		},
		{
			name: "Test Invalid Uint16 Returns Default",
			env:  [][2]string{{PORT, "add_port"}},
			want: config.InfraConfig{
				Environment:  config.ENV_PRODUCTION,
				Port:         8080,
				IsHttpSecure: true,
			},
		},
		{
			name: "Test Allowed Port Returns Port",
			env:  [][2]string{{PORT, "3000"}},
			want: config.InfraConfig{
				Environment:  config.ENV_PRODUCTION,
				Port:         3000,
				IsHttpSecure: true,
			},
		},
		{
			name: "Test Not HttpSecure Returns False",
			env:  [][2]string{{REQUIRE_HTTPS, "false"}},
			want: config.InfraConfig{
				Environment:  config.ENV_PRODUCTION,
				Port:         8080,
				IsHttpSecure: false,
			},
		},
		{
			name: "Test Invalid HttpSecure Returns True",
			env:  [][2]string{{REQUIRE_HTTPS, "FALSE"}},
			want: config.InfraConfig{
				Environment:  config.ENV_PRODUCTION,
				Port:         8080,
				IsHttpSecure: true,
			},
		},
		{
			name: "Test Invalid Environment Returns PRODUCTION",
			env:  [][2]string{{ENVIRONMENT, "UAT"}},
			want: config.InfraConfig{
				Environment:  config.ENV_PRODUCTION,
				Port:         8080,
				IsHttpSecure: true,
			},
		},
		{
			name: "Test DEVELOPMENT Returns ENV_DEVELOPMENT",
			env:  [][2]string{{ENVIRONMENT, "DEVELOPMENT"}},
			want: config.InfraConfig{
				Environment:  config.ENV_DEVELOPMENT,
				Port:         8080,
				IsHttpSecure: true,
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
			if got := c.NewInfraConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.NewInfraConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
