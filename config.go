package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MeiliHost   string        `yaml:"meili_host"`
	MeiliKey    string        `yaml:"meili_key"`
	Postgres    string        `yaml:"postgres_dsn"`
	Indexes     []IndexConfig `yaml:"indexes"`
	BatchSize   int           `yaml:"batch_size"`
	EnableAsync bool          `yaml:"enable_async"`
	WaitTime    int           `yaml:"wait_time"`
}

func (c *Config) Parse(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, c)
}

func (c *Config) Save(path string) error {
	b, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0644)
}

type IndexConfig struct {
	Source      string   `yaml:"source"`
	Destination string   `yaml:"destination"`
	Primary     string   `yaml:"primary"`
	Searchable  []string `yaml:"searchable"`
	Filterable  []string `yaml:"filterable"`
	Sortable    []string `yaml:"sortable"`
	Cursor      Cursor   `yaml:"cursor"`
}

type Cursor struct {
	Column   string    `yaml:"column"`
	LastSync time.Time `yaml:"last_sync"`
}
