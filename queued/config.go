package queued

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Port   uint
	Auth   string
	Store  string
	DbPath string
	Sync   bool
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) CreateStore(name string) Store {
	if c.Store == "leveldb" {
		if _, err := os.Stat(c.DbPath); err != nil && os.IsNotExist(err) {
			if err := os.Mkdir(c.DbPath, 0755); err != nil {
				panic(fmt.Sprintf("queued.CreateStore: Error os.Mkdir: %v", err))
			}
		}
		return NewLevelStore(filepath.Join(c.DbPath, name), c.Sync)
	} else if c.Store == "memory" {
		return NewMemoryStore()
	} else {
		panic(fmt.Sprintf("queued.Config: Invalid store: %s", c.Store))
	}
}
