package queued

import (
	"encoding/json"
	"fmt"
	"os"
)

type ConfigStore struct {
	path    string
	configs map[string]QueueConfig
}

func NewConfigStore(path string, sync bool) *ConfigStore {
	fh, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("queued.ConfigStore: unable to open file: %v", err))
	}
	defer fh.Close()
	dec := json.NewDecoder(fh)
	var configs map[string]QueueConfig
	dec.Decode(&configs)
	qs := ConfigStore{
		path:    path,
		configs: configs,
	}

	return &qs
}

func (qs *ConfigStore) GetQueueConfigs() map[string]QueueConfig {
	return qs.configs
}
func (qs *ConfigStore) Get(name string) QueueConfig {
	return qs.configs[name]
}

func (qs *ConfigStore) PutQueue(config *QueueConfig) error {
	qs.configs[config.Name] = *config
	fh, err := os.Open(qs.path)
	if err != nil {
		panic(fmt.Sprintf("queued.QuereConfigStore: unable to open file: %v", err))
	}
	defer fh.Close()

	b, err := json.Marshal(*qs)
	if err != nil {
		panic(fmt.Sprintf("queued.QuereConfigStore: %v", err))
	}
	fh.Write(b)
	return err
}
