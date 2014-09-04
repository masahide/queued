package queued

import "fmt"

type Config struct {
	Port       uint
	Auth       string
	Store      string
	DbPath     string
	ConfigPath string
	Sync       bool
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) CreateStore() Store {
	if c.Store == "leveldb" {
		return NewLevelStore(c.DbPath, c.Sync)
	} else if c.Store == "memory" {
		return NewMemoryStore()
	} else {
		panic(fmt.Sprintf("queued.Config: Invalid store: %s", c.Store))
	}
}
func (c *Config) CreateConfigStore() ConfigStore {
	if c.ConfigPath != "" {
		return NewJsonConfigStore(c.ConfigPath)
	} else {
		return NewMemoryConfigStore()
	}
}
