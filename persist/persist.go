package persist

import (
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"log"
	"strconv"
	"time"
)

var db store.Store

func Init(dbpath string) {
	boltdb.Register()

	_db, err := libkv.NewStore(
		store.BOLTDB,
		[]string{dbpath},
		&store.Config{
			Bucket:            "http_poll",
			ConnectionTimeout: time.Second,
		},
	)
	if err != nil {
		log.Fatal("Cannot create store consul ", err)
	}

	db = _db
}

func Get() store.Store {
	return db
}

//
// prefixed store.Store wrapper
//

func GetPrefixed(prefix string) *prefixedStore {
	return &prefixedStore{
		Get(),
		prefix + ".",
	}
}

type prefixedStore struct {
	db     store.Store
	prefix string
}

// Put a value at the specified key
func (s *prefixedStore) Put(key string, value []byte, options *store.WriteOptions) error {
	return s.db.Put(s.prefix+key, value, options)
}

// Get a value given its key
func (s *prefixedStore) Get(key string) (*store.KVPair, error) {
	return s.db.Get(s.prefix + key)
}

// Delete the value at the specified key
func (s *prefixedStore) Delete(key string) error {
	return s.db.Delete(s.prefix + key)
}

// Verify if a Key exists in the store
func (s *prefixedStore) Exists(key string) (bool, error) {
	return s.db.Exists(s.prefix + key)
}

// Watch for changes on a key
func (s *prefixedStore) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	return s.db.Watch(s.prefix+key, stopCh)
}

// WatchTree watches for changes on child nodes under
// a given directory
func (s *prefixedStore) WatchTree(directory string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	return s.db.WatchTree(s.prefix+directory, stopCh)
}

// NewLock creates a lock for a given key.
// The returned Locker is not held and must be acquired
// with `.Lock`. The Value is optional.
func (s *prefixedStore) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	return s.db.NewLock(s.prefix+key, options)
}

// List the content of a given prefix
func (s *prefixedStore) List(directory string) ([]*store.KVPair, error) {
	return s.db.List(s.prefix + directory)
}

// DeleteTree deletes a range of keys under a given directory
func (s *prefixedStore) DeleteTree(directory string) error {
	return s.db.DeleteTree(s.prefix + directory)
}

// Atomic CAS operation on a single value.
// Pass previous = nil to create a new key.
func (s *prefixedStore) AtomicPut(key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	return s.db.AtomicPut(s.prefix+key, value, previous, options)
}

// Atomic delete of a single value
func (s *prefixedStore) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	return s.db.AtomicDelete(s.prefix+key, previous)
}

// Close the store connection
func (s *prefixedStore) Close() {
	s.db.Close()
}

//
// Typed store
//

type TypedStore struct {
	store.Store
}

func (s *TypedStore) passValue(key string, f func(*store.KVPair) error) (bool, error) {
	kv, err := s.Get(key)
	switch err {
	case store.ErrKeyNotFound:
		return false, nil
	case nil:
	default:
		return false, err
	}
	return true, f(kv)
}

func (s *TypedStore) Uint64(key string, ptr *uint64) (exists bool, err error) {
	return s.passValue(key, func(kv *store.KVPair) error {
		value, err := strconv.ParseUint(string(kv.Value), 10, 64)
		if err == nil {
			*ptr = value
		}
		return err

	})
}

func (s *TypedStore) PutUint64(key string, value uint64) error {
	return s.Put(key, []byte(strconv.FormatUint(value, 10)), nil)
}

func (s *TypedStore) Int64(key string, ptr *int64) (exists bool, err error) {
	return s.passValue(key, func(kv *store.KVPair) error {
		value, err := strconv.ParseInt(string(kv.Value), 10, 64)
		if err == nil {
			*ptr = value
		}
		return err

	})
}

func (s *TypedStore) PutInt64(key string, value int64) error {
	return s.Put(key, []byte(strconv.FormatInt(value, 10)), nil)
}

func (s *TypedStore) Int32(key string, ptr *int32) (exists bool, err error) {
	return s.passValue(key, func(kv *store.KVPair) error {
		value, err := strconv.ParseInt(string(kv.Value), 10, 32)
		if err == nil {
			*ptr = int32(value)
		}
		return err
	})
}

func (s *TypedStore) PutInt32(key string, value int32) error {
	return s.Put(key, []byte(strconv.FormatInt(int64(value), 10)), nil)
}
