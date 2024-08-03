package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type ExtraConfig struct {
	BaseConfig
	Foo struct {
		Bar string `yaml:"bar" env:"BAR"`
	} `yaml:"foo"`
}

func TestGetConfig(t *testing.T) {
	type args struct {
		env Environment
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test empty environment",
			args: args{
				env: "",
			},
			wantErr: false,
		},
		{
			name: "Test local environment",
			args: args{
				env: EnvironmentLocal,
			},
			wantErr: false,
		},
		{
			name: "Test prod environment",
			args: args{
				env: EnvironmentProd,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := GetConfig[ExtraConfig](tt.args.env)

			if err != nil && tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if cfg != nil {
				assert.Equal(t, "changeme", cfg.API.Token)
				assert.Equal(t, "baz", cfg.Foo.Bar)
			}
		})
	}
}
