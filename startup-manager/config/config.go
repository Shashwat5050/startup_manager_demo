package config

import core "startup-manager/core/config"

type Config struct {
	AppConfig core.AppConfig `yaml:inline`
	NomadURL  string         `json:"nomad_url" yaml:"nomad_url"`
	HttpPort  string         `json:"http_port" yaml:"http_port"`
}

func (c *Config) GetAppConfig() *core.AppConfig {
	return &c.AppConfig
}

func (c *Config) GetDbConfig() *core.DbConfig {
	return c.AppConfig.DbConfig
}
