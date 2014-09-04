package queued

type MemoryConfigStore struct {
	Queues map[string]QueueConfig
}

func NewMemoryConfigStore() *MemoryConfigStore {
	cs := MemoryConfigStore{}
	cs.Queues = map[string]QueueConfig{}
	return &cs
}

func (qs *MemoryConfigStore) GetQueueConfigs() map[string]QueueConfig {
	return qs.Queues
}
func (qs *MemoryConfigStore) GetQueueConfig(name string) QueueConfig {
	return qs.Queues[name]
}

func (qs *MemoryConfigStore) PutQueue(config *QueueConfig) error {
	qs.Queues[config.Name] = *config
	return nil
}

func (qs *MemoryConfigStore) Drop() {
	qs.Queues = map[string]QueueConfig{}
}
