package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type AppCfg struct {
	Server struct {
		Port int
	}
	Database struct {
		User     string
		Password string
		Name     string
		Host     string
		Port     uint16
		SSLMode  string
	}
	Kafka KafkaConfig
}

type KafkaConfig struct {
	Brokers       string
	Topic         string
	ConsumerGroup string
}

const (
	configPath = "configuration"
)

func LoadConfig(env string) (*AppCfg, error) {
	v := viper.New()

	// load base config.yaml
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// load and merge environment-specific config (if exists)
	envFile := fmt.Sprintf("%s/config.%s.yaml", configPath, env)
	if _, err := os.Stat(envFile); err == nil {
		v.SetConfigFile(envFile)
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("error reading %s config: %w", env, err)
		}
	}

	v.AutomaticEnv()

	var cfg AppCfg
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
