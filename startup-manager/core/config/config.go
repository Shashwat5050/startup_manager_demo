package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ServiceConfig interface {
	GetAppConfig() *AppConfig
	GetDbConfig() *DbConfig
}

type AppConfig struct {
	DbConfig            *DbConfig    `json:"db_config" yaml:"db_config"`
	DomainName          string       `json:"domain_name" yaml:"domain_name"`
	AuthorizerTableName string       `json:"authorizer_table_name" yaml:"authorizer_table_name"`
	RedisConfig         *RedisConfig `json:"redis_config" yaml:"redis_config"`
	ActivityManagerURL  string       `json:"activity_manager_url" yaml:"activity_manager_url"`
	Probe               *ProbeConfig `json:"probe" yaml:"probe"`
	CorsDomains         []string     `json:"cors_domains" yaml:"cors_domains"`
}

type RedisConfig struct {
	URL      string `json:"url" yaml:"url"`
	Password string `json:"password" yaml:"password"`
}

type ProbeConfig struct {
	Health string `json:"health" yaml:"health"`
	Ready  string `json:"ready" yaml:"ready"`
	Prefix string `json:"prefix" yaml:"prefix"`
}

type DbConfig struct {
	ConnectionString string `json:"connection_string" yaml:"connection_string"`
	MaxConnRetries   int    `json:"max_conn_retries" yaml:"max_conn_retries"`
	MaxOpenConns     int    `json:"max_open_conns" yaml:"max_open_conns"`
	MigrationsPath   string `json:"migrations_path" yaml:"migrations_path"`
}

// LoadConfig loads config from yml or json file
// you can load your own service config which has AppConfig in it
func LoadConfig[T ServiceConfig](fileName string) (T, error) {
	var (
		config T
		err    error
	)

	ext := filepath.Ext(fileName)

	switch ext {
	case ".yml", ".yaml":
		config, err = loadYmlConfig[T](fileName)
		// log.Println("config", config)
	case ".json":
		config, err = loadJsonConfig[T](fileName)
	default:
		err = fmt.Errorf("invalid config format: %s", fileName)
	}

	if err != nil {
		// log.Println("err in switch", err)
		return config, err
	}
	config.GetAppConfig().DbConfig = &DbConfig{ConnectionString: "host=localhost port=5432 user=postgres dbname=iceline password=postgres sslmode=disable"}

	config.GetAppConfig().setDefaults()
	// maybe parse env
	return config, nil
}

func (c *AppConfig) setDefaults() {
	if c.DbConfig != nil {
		c.DbConfig.setDefaults()
	}
	if c.Probe == nil {
		c.Probe = defaultProbeConfig()
	} else {
		c.Probe.setDefaults()
	}
	if c.ActivityManagerURL == "" {
		c.ActivityManagerURL = "activity:9003"
	}
	if c.CorsDomains == nil {
		c.CorsDomains = []string{"*"}
	}
}

func (c *DbConfig) setDefaults() {
	if c.MaxConnRetries == 0 {
		c.MaxConnRetries = 5
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = -1
	}
	if c.MigrationsPath == "" {
		c.MigrationsPath = "./migrations/"
	}
}

func (c *ProbeConfig) setDefaults() {
	if c.Health == "" {
		c.Health = "/health"
	} else if c.Health[0] != '/' {
		c.Health = "/" + c.Health
	}
	if c.Ready == "" {
		c.Ready = "/ready"
	} else if c.Ready[0] != '/' {
		c.Ready = "/" + c.Ready
	}
	if c.Prefix == "" {
		c.Prefix = "/probe"
	} else if c.Prefix[0] != '/' {
		c.Prefix = "/" + c.Prefix
	}
}

func defaultProbeConfig() *ProbeConfig {
	return &ProbeConfig{
		Health: "/health",
		Ready:  "/ready",
		Prefix: "/probe",
	}
}

func loadYmlConfig[T ServiceConfig](fileName string) (T, error) {
	var config T
	// currentDir,err:=os.Getwd()
	// if err!=nil{
	// 	log.Println("error is -------",err)

	// }
	// log.Println("currentDir",currentDir)
	// log.Println(fileName)

	yfile, err := os.ReadFile(filepath.Clean(fileName))
	// log.Println("yfile", string(yfile))
	if err != nil {
		// log.Println(yfile, "yfile,err", err)
		return config, err
	}

	err = yaml.Unmarshal(yfile, &config)
	if err != nil {

		return config, err
	}
	// log.Println(config, "*********")

	return config, nil
}

func loadJsonConfig[T ServiceConfig](fileName string) (T, error) {
	var config T

	file, err := os.Open(filepath.Clean(fileName))
	if err != nil {
		return config, err
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		return config, err
	}

	return config, nil
}
