package queued

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestLevelStore(t *testing.T) {
	store := NewLevelStore("./test1.db", true)
	defer store.Drop()

	assert.Equal(t, store.id, 0)

	record := NewRecord([]byte("foo"), "testqueue")

	err := store.Put(record)
	assert.Equal(t, err, nil)
	assert.Equal(t, record.Id, 1)

	record, err = store.Get(1)
	assert.Equal(t, err, nil)
	assert.Equal(t, record.Id, 1)
	assert.Equal(t, record.Value, []byte("foo"))
	assert.Equal(t, record.Queue, "testqueue")

	err = store.Remove(1)
	assert.Equal(t, err, nil)

	record, err = store.Get(1)
	assert.Equal(t, err.Error(), "leveldb: not found")
	assert.T(t, record == nil)
}

func TestLevelStoreLoad(t *testing.T) {
	temp := NewLevelStore("./test2.db", true)
	temp.Put(NewRecord([]byte("foo"), "testqueue"))
	temp.Put(NewRecord([]byte("bar"), "testqueue"))
	temp.Close()

	store := NewLevelStore("./test2.db", true)
	defer store.Drop()

	assert.Equal(t, store.id, 2)
}

func TestLevelStoreIterator(t *testing.T) {
	temp := NewLevelStore("./test3.db", true)
	temp.Put(NewRecord([]byte("foo"), "testqueue"))
	temp.Put(NewRecord([]byte("bar"), "testqueue"))
	temp.Close()

	store := NewLevelStore("./test3.db", true)
	defer store.Drop()

	it := store.Iterator()

	one, ok := it.NextRecord()
	assert.Equal(t, ok, true)
	assert.Equal(t, one.Id, 1)
	assert.Equal(t, one.Value, []byte("foo"))
	assert.Equal(t, one.Queue, "testqueue")

	two, ok := it.NextRecord()
	assert.Equal(t, ok, true)
	assert.Equal(t, two.Id, 2)
	assert.Equal(t, two.Value, []byte("bar"))
	assert.Equal(t, two.Queue, "testqueue")

	_, ok = it.NextRecord()
	assert.Equal(t, ok, false)
}
