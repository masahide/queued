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

func NewConfigStore(path string) ConfigStore {
	fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		panic(fmt.Sprintf("queued.ConfigStore: unable to open file: %v", err))
	}
	defer fh.Close()
	dec := json.NewDecoder(fh)
	var configs map[string]QueueConfig
	dec.Decode(&configs)

	return ConfigStore{
		path:    path,
		configs: configs,
	}

}

func (qs ConfigStore) GetQueueConfigs() map[string]QueueConfig {
	return qs.configs
}
func (qs ConfigStore) Get(name string) QueueConfig {
	return qs.configs[name]
}

func (qs ConfigStore) PutQueue(config *QueueConfig) error {
	qs.configs[config.Name] = *config
	fh, err := os.OpenFile(qs.path, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(fmt.Sprintf("queued.QueueConfigStore: unable to open file: %v", err))
	}
	defer fh.Close()

	b, err := json.Marshal(qs)
	if err != nil {
		panic(fmt.Sprintf("queued.QueueConfigStore: %v", err))
	}
	fh.Write(b)
	return err
}

func (qs ConfigStore) Drop() {
	err := os.RemoveAll(qs.path)
	if err != nil {
		panic(fmt.Sprintf("queued.QueueConfigStore: Error Remove file: %v", err))
	}
}
