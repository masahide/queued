package queued

import (
	"encoding/json"
	"fmt"
	"os"
)

type JsonConfigStore struct {
	path   string
	Queues map[string]QueueConfig
}

func NewJsonConfigStore(path string) *JsonConfigStore {
	fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		panic(fmt.Sprintf("queued.JsonConfigStore: unable to open file: %s, err:%v", path, err))
	}
	defer fh.Close()
	cs := JsonConfigStore{}
	dec := json.NewDecoder(fh)
	dec.Decode(&cs)
	cs.path = path
	if cs.Queues == nil {
		cs.Queues = map[string]QueueConfig{}
	}

	return &cs
}

func (qs *JsonConfigStore) GetQueueConfigs() map[string]QueueConfig {
	return qs.Queues
}
func (qs *JsonConfigStore) GetQueueConfig(name string) QueueConfig {
	return qs.Queues[name]
}

func (qs *JsonConfigStore) PutQueue(config *QueueConfig) error {
	qs.Queues[config.Name] = *config
	fh, err := os.OpenFile(qs.path, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(fmt.Sprintf("queued.JsonConfigStore: unable to open file: %s, err:%v", qs.path, err))
	}
	defer fh.Close()

	b, err := json.MarshalIndent(qs, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("queued.JsonConfigStore: %v", err))
	}
	fh.Write(b)
	return err
}

func (qs *JsonConfigStore) Drop() {
	err := os.RemoveAll(qs.path)
	if err != nil {
		panic(fmt.Sprintf("queued.QueueConfigStore: Error Remove file: %v", err))
	}
}
