package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/nanoteck137/validate"
	"github.com/nanoteck137/versionctl"
	"github.com/spf13/viper"
)

type Config struct {
	PreCmd string `mapstructure:"pre_cmd"`
	Push   bool   `mapstructure:"push"`
}

func (c Config) Validate() error {
	return validate.ValidateStruct(&c,
		validate.Field(&c.PreCmd),
		validate.Field(&c.Push),
	)
}

func defaults(v *viper.Viper) {
	v.SetDefault("pre_cmd", "")
	v.SetDefault("push", true)
}

func WriteNewConfig() error {
	v := viper.New()

	defaults(v)

	filename := "." + versionctl.AppName + ".toml"

	err := v.SafeWriteConfigAs(filename)
	if err != nil {
		return err
	}

	fmt.Printf("wrote config '%s'\n", filename)

	return nil
}

func Load() (*Config, error) {
	v := viper.New()

	defaults(v)

	filename := "." + versionctl.AppName + ".toml"
	v.SetConfigFile(filename)

	err := v.ReadInConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var config Config
	err = v.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	slog.Info("loaded config", "config", config)

	// TODO(patrik): I hate this
	oldTag := validate.ErrorTag
	validate.ErrorTag = "mapstructure"
	defer func() {
		validate.ErrorTag = oldTag
	}()

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}
