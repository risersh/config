package config

import (
	"fmt"
	"os"
	"reflect"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/risersh/util/validation"
)

// BaseConfig is inherited by caller configs.
type BaseConfig struct {
	Environment struct {
		Name          string `yaml:"name" env:"NAME"`
		Containerized bool   `yaml:"containerized" env:"CONTAINERIZED" required:"false"`
	} `yaml:"environment" env-prefix:"ENVIRONMENT_"`
	Public struct {
		Hostname string `yaml:"hostname" env:"HOSTNAME"`
		TLS      struct {
			Disabled bool   `yaml:"disabled" env:"TLS_DISABLED" required:"false"`
			Cert     string `yaml:"cert" env:"TLS_CERT" required:"false"`
			Key      string `yaml:"key" env:"TLS_KEY" required:"false"`
		} `yaml:"tls" env-prefix:"TLS_" required:"false"`
	} `yaml:"public" env-prefix:"PUBLIC_" required:"true"`
	API struct {
		BaseURL string `yaml:"baseUrl" env:"BASE_URL" required:"false"`
		Token   string `yaml:"token" env:"TOKEN" required:"false"`
	} `yaml:"api" env-prefix:"API_"`
	Database struct {
		URI string `yaml:"uri" env:"URI" required:"false"`
	} `yaml:"database" env-prefix:"DATABASE_" required:"false"`
	RabbitMQ struct {
		URI string `yaml:"uri" env:"URI" required:"false"`
	} `yaml:"rabbitmq" env-prefix:"RABBITMQ_" required:"false"`
	Elasticsearch struct {
		URL      string `yaml:"url" env:"URL"`
		Username string `yaml:"username" env:"USERNAME"`
		Password string `yaml:"password" env:"PASSWORD"`
	} `yaml:"elasticsearch" env-prefix:"ELASTICSEARCH_"`
	Mail struct {
		Outbound struct {
			Key string `yaml:"key" env:"KEY" required:"false"`
		} `yaml:"outbound" env-prefix:"OUTBOUND_" required:"false"`
	} `yaml:"mail" env-prefix:"MAIL_" required:"false"`
	Sessions struct {
		PublicKey  string `yaml:"publicKey" env:"PUBLIC_KEY" required:"false"`
		PrivateKey string `yaml:"privateKey" env:"PRIVATE_KEY" required:"false"`
	} `yaml:"sessions" env-prefix:"SESSIONS_" required:"false"`
	Monitoring struct {
		Enabled bool `yaml:"enabled" env:"ENABLED"`
		Tracing struct {
			Enabled   bool   `yaml:"enabled" env:"ENABLED" required:"false"`
			Collector string `yaml:"collector" env:"COLLECTOR" required:"false"`
		} `yaml:"tracing" env-prefix:"TRACING_" required:"false"`
	} `yaml:"monitoring" env-prefix:"MONITORING_"`
}

// GetConfig returns a config of type T.
// It will merge the base config with the environment config.
// If the environment config does not exist, it will use the base config.
//
// Arguments:
//   - env: The environment to use.
//
// Returns:
//   - A pointer to the config of type T.
//   - An error if the config could not be found.
func GetConfig[T any](env Environment) (*T, error) {
	var base *BaseConfig
	c := new(T)

	if _, err := os.Stat(".env.local.base.yaml"); err == nil {
		base = &BaseConfig{}
		cleanenv.ReadConfig(".env.local.base.yaml", &base)
	} else if _, err := os.Stat("../.env.local.base.yaml"); err == nil {
		base = &BaseConfig{}
		cleanenv.ReadConfig("../.env.local.base.yaml", &base)
	} else if _, err := os.Stat("../../.env.local.base.yaml"); err == nil {
		base = &BaseConfig{}
		cleanenv.ReadConfig("../../.env.local.base.yaml", &base)
	} else if _, err := os.Stat("../../../.env.local.base.yaml"); err == nil {
		base = &BaseConfig{}
		cleanenv.ReadConfig("../../../.env.local.base.yaml", &base)
	} else if _, err := os.Stat("../../../../.env.local.base.yaml"); err == nil {
		base = &BaseConfig{}
		cleanenv.ReadConfig("../../../../.env.local.base.yaml", &base)
	} else {
		cleanenv.ReadEnv(&base)
	}

	if base == nil {
		return nil, fmt.Errorf("base config not found in search paths")
	}

	if env == "" {
		env = EnvironmentLocal
	}

	if _, err := os.Stat(fmt.Sprintf(".env.%s.yaml", env)); err == nil {
		cleanenv.ReadConfig(fmt.Sprintf("./.env.%s.yaml", env), &c)
	} else if _, err := os.Stat(fmt.Sprintf("../.env.%s.yaml", env)); err == nil {
		cleanenv.ReadConfig(fmt.Sprintf("../.env.%s.yaml", env), &c)
	} else if _, err := os.Stat(fmt.Sprintf("../../.env.%s.yaml", env)); err == nil {
		cleanenv.ReadConfig(fmt.Sprintf("../../.env.%s.yaml", env), &c)
	} else {
		cleanenv.ReadEnv(&c)
	}

	// Merge base values into c if c does not have a value.
	if baseC, ok := any(c).(*BaseConfig); ok {
		// If c is of type *BaseConfig, we can directly assign.
		*baseC = *base
	} else {
		// If not, we need to copy fields manually.
		cVal := reflect.ValueOf(c).Elem()
		baseVal := reflect.ValueOf(base).Elem()

		for i := 0; i < baseVal.NumField(); i++ {
			baseField := baseVal.Field(i)
			baseFieldName := baseVal.Type().Field(i).Name

			if cField := cVal.FieldByName(baseFieldName); cField.IsValid() && cField.CanSet() {
				if cField.Type() == baseField.Type() {
					cField.Set(baseField)
				}
			}
		}
	}

	emptyFields, err := validation.ValidateStructFields(base, "")
	if err != nil {
		return nil, err
	}
	if len(emptyFields) > 0 {
		return nil, fmt.Errorf("empty fields: %v", emptyFields)
	}

	emptyFields, err = validation.ValidateStructFields(c, "")
	if err != nil {
		return nil, err
	}
	if len(emptyFields) > 0 {
		return nil, fmt.Errorf("empty fields: %v", emptyFields)
	}

	return c, nil
}
