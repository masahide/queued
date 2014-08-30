package queued

type Item struct {
	value        int
	dequeued     bool
	dequeueCount int
	complete     chan bool
}

func NewItem(value int) *Item {
	return &Item{
		value:        value,
		dequeued:     false,
		dequeueCount: 0,
		complete:     make(chan bool),
	}
}

func (item *Item) Complete() {
	if !item.dequeued {
		return
	}

	item.complete <- true
}
