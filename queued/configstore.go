package queued

type ConfigStore interface {
	GetQueueConfigs() map[string]QueueConfig
	GetQueueConfig(name string) QueueConfig
	PutQueue(config *QueueConfig) error
	Drop()
}
