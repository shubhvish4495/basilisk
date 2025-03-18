package config

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/shubhvish4495/basilisk/pkg/helper"
	"github.com/shubhvish4495/basilisk/pkg/jwt"
	"gopkg.in/yaml.v3"
)

type Config struct {
	TlsConfig TlsConfig `yaml:"tlsConfig"`
	Database  Database  `yaml:"database"`
	JWT       JWT       `yaml:"jwt"`
}

type JWT struct {
	Secret string `yaml:"secret"`
}

type TlsConfig struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		helper.GetLogger().Info("No Configuration found. Loading default configuration")
		if err := Load(); err != nil {
			panic(err)
		}
	}
	return config
}

// Load reads the configuration from a YAML file and unmarshals it into the config variable.
// It returns an error if reading the file or unmarshalling the data fails.
func Load() error {
	data, err := os.ReadFile("./config/config.yml")
	expData := os.ExpandEnv(string(data))
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(expData), &config)
	if err != nil {
		return err
	}

	// jwt secret is base64 encoded. We will decode it first and then set it in config
	decStr, err := base64.StdEncoding.DecodeString(config.JWT.Secret)
	if err != nil {
		return fmt.Errorf("error while decoding jwt secret %v", err)
	}

	// once decoded, then set it into config.JWT.Secret
	jwt.SetTokenSecret(string(decStr))
	return nil
}
