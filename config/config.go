package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Clusters []Cluster
	Interval int
	SignalFx SignalFxCfg
}

type SignalFxCfg struct {
	AuthToken string `yaml:"auth_token"`
}

type Cluster struct {
	Uris     []string
	Username string
	Password string
}

func ReadConfig(path string) (*Config, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err = yaml.Unmarshal(buf, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
