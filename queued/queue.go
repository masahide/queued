package queued

import (
	"sync"
	"time"
)

const NilDuration = time.Duration(-1)

type Queue struct {
	items   []*Item
	waiting chan *Item
	stats   *Stats
	mutex   sync.Mutex
	config  *QueueConfig
	app     *Application
}

func NewQueue(config *QueueConfig) *Queue {
	counters := map[string]int{
		"enqueued": 0,
		"dequeued": 0,
		"depth":    0,
		"timeouts": 0,
	}

	return &Queue{
		items:   []*Item{},
		waiting: make(chan *Item),
		stats:   NewStats(counters),
		config:  config,
	}
}

func (q *Queue) Enqueue(value int) *Item {
	item := NewItem(value)
	q.EnqueueItem(item)
	return item
}

func (q *Queue) EnqueueItem(item *Item) {
	q.stats.Inc("enqueued")

	select {
	case q.waiting <- item:
	default:
		q.append(item)
	}
}

func (q *Queue) Dequeue(wait time.Duration, timeout time.Duration) *Item {
	q.stats.Inc("dequeued")

	if timeout == NilDuration && q.config.Timeout != QueueNilTimeout {
		timeout = time.Duration(q.config.Timeout) * time.Second
	}
	if item := q.shift(); item != nil {
		q.timeout(item, timeout)
		return item
	} else if wait != NilDuration {
		select {
		case <-time.After(wait):
			return nil
		case item := <-q.waiting:
			q.timeout(item, timeout)
			return item
		}
	} else {
		return nil
	}
}

func (q *Queue) Stats() map[string]int {
	return q.stats.Get()
}

func (q *Queue) shift() *Item {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.items) > 0 {
		item := q.items[0]
		q.items = q.items[1:]
		q.stats.Dec("depth")
		return item
	} else {
		return nil
	}
}

func (q *Queue) append(item *Item) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.items = append(q.items, item)
	q.stats.Inc("depth")
}

func (q *Queue) timeout(item *Item, timeout time.Duration) {
	if timeout != NilDuration {
		item.dequeued = true
		if q.config.ExponentialBackoff {
			timeout = timeout * (1 << uint(item.dequeueCount))
		}
		item.dequeueCount++

		go func() {
			select {
			case <-time.After(timeout):
				if q.config.Redirve && item.dequeueCount >= q.config.MaximumReceives {
					q.app.DeadLetterQueue(q.config.DeadLetterQueue, item)
				} else {
					item.dequeued = false
					q.EnqueueItem(item)
				}
				q.stats.Inc("timeouts")
			case <-item.complete:
				item.dequeued = false
			}
		}()
	}
}
