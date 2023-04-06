package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host        string `yaml:"host"`
		Port        string `yaml:"port"`
		HashWorkers int    `yaml:"worker-count"`
		CacheCount  int    `yaml:"cache-count"`
		DSN         string
		Name        string `yaml:"name"`
		DB          struct {
			DBType    string `yaml:"db-type"`
			PoolCount int    `yaml:"pool-count"`
			Host      string `yaml:"host"`
			Port      string `yaml:"port"`
			User      string `yaml:"user"`
			Pass      string `yaml:"pass"`
			DBName    string `yaml:"dbname"`
			SSL       string `yaml:"ssl"`
		} `yaml:"db"`
	} `yaml:"server"`
	Metrics struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
		Path string `yaml:"path"`
	} `yaml:"metric"`
	Logging struct {
		Provider string `yaml:"provider"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Level    int    `yaml:"level"`
	} `yaml:"logging"`
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	config.Server.DSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Server.DB.Host,
		config.Server.DB.Port,
		config.Server.DB.User,
		config.Server.DB.Pass,
		config.Server.DB.DBName,
		config.Server.DB.SSL)

	return config, nil
}
