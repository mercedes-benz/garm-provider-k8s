// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	KubeConfigPath    string `mapstructure:"KubeConfigPath"`
	ContainerRegistry string `mapstructure:"ContainerRegistry"`
	RunnerNamespace   string `mapstructure:"RunnerNamespace"`
}

func NewConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, errors.New("no config file path provided")
	}
	dir, file := filepath.Split(configPath)
	filenameWithoutExt := file[0 : len(file)-len(filepath.Ext(file))]

	viper.SetConfigName(filenameWithoutExt)
	viper.AddConfigPath(dir)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Println("No configuration file found, using default values")
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if config.ContainerRegistry == "" {
		return nil, errors.New("property `ContainerRegistry` in provider config must be set")
	}

	if config.RunnerNamespace == "" {
		config.RunnerNamespace = "runner"
	}

	return &config, nil
}
