package queued

type QueueConfig struct {
	Name               string
	Redirve            bool
	DeadLetterQueue    string
	MaximumReceives    int
	ExponentialBackoff bool
	Timeout            int
}

const QueueNilTimeout = -1

var defaultQueueConfig = QueueConfig{
	Redirve:            false,
	DeadLetterQueue:    "DeadLetter",
	MaximumReceives:    10,
	ExponentialBackoff: false,
	Timeout:            QueueNilTimeout,
}

func NewQueueConfig() *QueueConfig {
	config := defaultQueueConfig
	return &config
}
