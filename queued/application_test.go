package queued

import (
	"testing"
	"time"

	"github.com/bmizerany/assert"
)

func TestApplication(t *testing.T) {
	itemstore := NewLevelStore("./test1item.db", true)
	defer itemstore.Drop()
	queuestore := NewJsonConfigStore("./test1config.json")
	defer queuestore.Drop()

	app, _ := NewApplication(queuestore, itemstore)

	getq1, _ := app.GetQueue("test")
	getq2, _ := app.GetQueue("test")
	assert.Equal(t, getq1, getq2)
	getq3, _ := app.GetQueue("test")
	getq4, _ := app.GetQueue("foobar")
	assert.NotEqual(t, getq3, getq4)

	record, err := app.Enqueue("test", []byte("foo"), "")

	assert.Equal(t, err, nil)
	assert.Equal(t, record.Id, 1)
	assert.Equal(t, record.Value, []byte("foo"))
	assert.Equal(t, record.Queue, "test")

	stats, _ := app.Stats("test")

	assert.Equal(t, stats["enqueued"], 1)
	assert.Equal(t, stats["dequeued"], 0)
	assert.Equal(t, stats["depth"], 1)
	assert.Equal(t, stats["timeouts"], 0)

	info, err := app.Info("test", 1)

	assert.Equal(t, err, nil)
	assert.Equal(t, info.record.Value, []byte("foo"))
	assert.Equal(t, info.dequeued, false)

	record, err = app.Dequeue("test", NilDuration, NilDuration)

	assert.Equal(t, err, nil)
	assert.T(t, record != nil)
	assert.Equal(t, record.Id, 1)
	assert.Equal(t, record.Value, []byte("foo"))
	assert.Equal(t, record.Queue, "test")

	ok, err := app.Complete("test", 1)
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, false)

	app.Enqueue("test", []byte("bar"), "")
	record, err = app.Dequeue("test", NilDuration, time.Millisecond)

	assert.Equal(t, err, nil)
	assert.T(t, record != nil)
	assert.Equal(t, record.Id, 2)
	assert.Equal(t, record.Value, []byte("bar"))
	assert.Equal(t, record.Queue, "test")

	ok, err = app.Complete("test", 2)
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)

	ok, err = app.Complete("test", 2)
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, false)
}

func TestNewApplication(t *testing.T) {
	itemstore := NewLevelStore("./test1item.db", true)
	defer itemstore.Drop()
	queuestore := NewJsonConfigStore("./test1config.json")
	defer queuestore.Drop()

	itemstore.Put(NewRecord([]byte("foo"), "test"))
	itemstore.Put(NewRecord([]byte("bar"), "test"))
	itemstore.Put(NewRecord([]byte("baz"), "another"))

	app, _ := NewApplication(queuestore, itemstore)

	one, _ := app.Dequeue("test", NilDuration, NilDuration)
	assert.Equal(t, one.Id, 1)
	assert.Equal(t, one.Value, []byte("foo"))

	two, _ := app.Dequeue("test", NilDuration, NilDuration)
	assert.Equal(t, two.Id, 2)
	assert.Equal(t, two.Value, []byte("bar"))

	three, _ := app.Dequeue("another", NilDuration, NilDuration)
	assert.Equal(t, three.Id, 3)
	assert.Equal(t, three.Value, []byte("baz"))
}
