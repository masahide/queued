package queued

import (
	"sync"
	"time"
)

type Info struct {
	record   *Record
	dequeued bool
}

type Application struct {
	itemStore   Store
	ConfigStore ConfigStore
	queues      map[string]*Queue
	items       map[int]*Item
	qmutex      sync.Mutex
	imutex      sync.RWMutex
}

func NewApplication(ConfigStore ConfigStore, itemStore Store) (*Application, error) {
	app := &Application{
		itemStore:   itemStore,
		ConfigStore: ConfigStore,
		queues:      make(map[string]*Queue),
		items:       make(map[int]*Item),
	}

	QueueConfigs := ConfigStore.GetQueueConfigs()

	for _, queueConfig := range QueueConfigs {
		app.makeQueue(&queueConfig)
	}

	it := itemStore.Iterator()
	record, ok := it.NextRecord()

	for ok {
		queue, err := app.GetQueue(record.Queue)
		if err != nil {
			return nil, err
		}
		item := queue.Enqueue(record.Id)
		app.items[item.value] = item

		record, ok = it.NextRecord()
	}

	return app, nil
}

func (a *Application) makeQueue(config *QueueConfig) *Queue {
	queue, ok := a.queues[config.Name]
	if !ok {
		queue = NewQueue(config)
		queue.app = a
		a.queues[config.Name] = queue
	}
	return queue
}

func (a *Application) CreateQueue(config *QueueConfig) (*Queue, error) {
	a.qmutex.Lock()
	defer a.qmutex.Unlock()
	err := a.ConfigStore.PutQueue(config)
	if err != nil {
		return nil, err
	}

	return a.makeQueue(config), err
}

func (a *Application) GetQueue(name string) (*Queue, error) {
	a.qmutex.Lock()
	defer a.qmutex.Unlock()

	queue, ok := a.queues[name]
	if !ok {
		config := NewQueueConfig()
		config.Name = name
		err := a.ConfigStore.PutQueue(config)
		if err != nil {
			return nil, err
		}
		return a.makeQueue(config), err
	}
	return queue, nil
}

func (a *Application) GetItem(id int) (*Item, bool) {
	a.imutex.RLock()
	defer a.imutex.RUnlock()

	item, ok := a.items[id]
	return item, ok
}

func (a *Application) PutItem(item *Item) {
	a.imutex.Lock()
	defer a.imutex.Unlock()

	a.items[item.value] = item
}

func (a *Application) RemoveItem(id int) {
	a.imutex.Lock()
	defer a.imutex.Unlock()

	delete(a.items, id)
}

func (a *Application) DeadLetterQueue(name string, item *Item) error {
	item.dequeued = false
	record, err := a.itemStore.Get(item.value)
	if err != nil {
		return err
	}
	err = a.EnqueueRecord(name, record)
	if err != nil {
		return err
	}
	err = a.itemStore.Remove(item.value)
	if err != nil {
		return err
	}
	a.RemoveItem(item.value)
	return nil
}

func (a *Application) Enqueue(name string, value []byte, mime string) (*Record, error) {
	record := NewRecord(value, name)
	record.Mime = mime

	err := a.EnqueueRecord(name, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (a *Application) EnqueueRecord(name string, record *Record) error {
	queue, err := a.GetQueue(name)
	if err != nil {
		return err
	}
	err = a.itemStore.Put(record)
	if err != nil {
		return err
	}

	item := queue.Enqueue(record.Id)
	a.PutItem(item)

	return nil
}

func (a *Application) Dequeue(name string, wait time.Duration, timeout time.Duration) (*Record, error) {
	queue, err := a.GetQueue(name)
	if err != nil {
		return nil, err
	}
	item := queue.Dequeue(wait, timeout)
	if item == nil {
		return nil, nil
	}

	record, err := a.itemStore.Get(item.value)
	if err != nil {
		return nil, err
	}

	if !item.dequeued {
		a.Complete(name, item.value)
	}

	return record, nil
}

func (a *Application) Remove(item *Item) (bool, error) {
	if !item.dequeued {
		return false, nil
	}

	err := a.itemStore.Remove(item.value)
	if err != nil {
		return false, err
	}

	item.Complete()
	a.RemoveItem(item.value)

	return true, nil
}
func (a *Application) Complete(name string, id int) (bool, error) {
	item, ok := a.GetItem(id)

	if !ok {
		return false, nil
	}

	return a.Remove(item)
}

func (a *Application) Info(name string, id int) (*Info, error) {
	record, err := a.itemStore.Get(id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}

	if record.Queue != name {
		return nil, nil
	}

	item, ok := a.GetItem(id)
	info := &Info{record, ok && item.dequeued}

	return info, nil
}

func (a *Application) Stats(name string) (map[string]int, error) {
	queue, err := a.GetQueue(name)
	if err != nil {
		return nil, err
	}
	return queue.Stats(), err
}

func (a *Application) ListQueues() map[string]*Queue {

	a.qmutex.Lock()
	defer a.qmutex.Unlock()

	queues := a.queues
	return queues
}
