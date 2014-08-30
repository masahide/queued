package queued

import "time"

type QueueConfig struct {
	Redirve         bool
	DeadLetterQueue string
	MaximumReceives int
	Timeout         time.Duration
}

var defaultQueueConfig = QueueConfig{
	Redirve:         false,
	DeadLetterQueue: "DeadLetter",
	MaximumReceives: 10,
	Timeout:         NilDuration,
}

func NewQueueConfig() *QueueConfig {
	config := defaultQueueConfig
	return &config
}
