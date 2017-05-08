package kv

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"

	// libkv backends
	boltdb "github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	// etcd "github.com/docker/libkv/store/etcd"
	// zk "github.com/docker/libkv/store/zookeeper"

	log "github.com/Sirupsen/logrus"
)

// NotFound - message for objects that could not be found
const NotFound = "Not found"

// Namespace - namespace in order not to clash with other systems
const Namespace = "apitest/"

// backend related errors
const (
	ErrFailedToLock string = "error, failed to acquire lock"
	ErrLockNotFound string = "error, lock not found"
)

const (

	// kv config//default KVstorage configuration
	DefaultStorageBackend = StorageBoltDB
	StorageBoltDB         = "boltdb"
	StorageConsul         = "consul"
	StorageETCD           = "etcd"
	DefaultConsulAddr     = "consul:8500"
)

func addNamespace(key string) string {
	var buffer bytes.Buffer
	buffer.WriteString(Namespace)
	if key != "/" {
		buffer.WriteString(key)
	}
	return buffer.String()
}

func removeNamespace(key string) string {
	return strings.TrimPrefix(key, Namespace)
}

// KVPair - kv pair to store results from kv backends
type KVPair struct {
	Key       string
	Value     []byte
	LastIndex uint64
}

// KVBackend - generic kv backend interface
type KVBackend interface {
	Put(key string, value []byte) error
	Get(key string) (*KVPair, error)
	Delete(key string) error
	Exists(key string) (bool, error)
	List(directory string) ([]*KVPair, error)
	DeleteTree(directory string) error
	Lock(key string) error
	Unlock(key string) error

	// renewable lock
	LockTTL(key string, ttl int, stopChan chan struct{}) (renew chan struct{}, stop <-chan struct{}, err error)
	ReleaseLock(key string) error

	Watch(key string, stopCh <-chan struct{}) (<-chan *KVPair, error)
	WatchTree(key string, stopCh <-chan struct{}) (<-chan []*KVPair, error)
}

// LibKVBackend - libkv container
type LibKVBackend struct {
	Options *store.Config
	Backend string
	Addrs   []string
	Store   store.Store
	Locks   map[string]store.Locker

	MutexLocks map[string]*sync.Mutex

	TTLLocks map[string]chan struct{}
}

func init() {
	consul.Register()
	boltdb.Register()
	// etcd.Register()
	log.Debug("stores registered")
}

// available backend stores
const (
	BoltDBBackend    = "boltdb"
	ZookeeperBackend = "zk"
	EtcdBackend      = "etcd"
	ConsulBackend    = "consul"
)

// NewLibKVBackend - returns new libkv based backend (https://github.com/docker/libkv)
func NewLibKVBackend(backend, bucket string, addrs []string) (*LibKVBackend, error) {
	// default backend - consul
	var backendStore store.Backend

	// TODO: options could include TLS details (certs) or bucket if backend is BoltDB
	options := &store.Config{
		ConnectionTimeout: 10 * time.Second,
	}

	if backend == EtcdBackend {
		backendStore = store.ETCD
		return nil, fmt.Errorf("this backend is currently not supported")
	} else if backend == ZookeeperBackend {
		backendStore = store.ZK
		return nil, fmt.Errorf("this backend is currently not supported")
	} else if backend == BoltDBBackend {
		backendStore = store.BOLTDB
		// setting bucket name to address
		options.Bucket = bucket
	} else {
		backendStore = store.CONSUL
	}

	log.WithFields(log.Fields{
		"backend": backendStore,
		"addrs":   addrs,
	}).Info("initiating store")

	libstore, err := libkv.NewStore(backendStore, addrs, options)
	if err != nil {
		log.WithFields(log.Fields{
			"backend": backendStore,
			"addrs":   addrs,
		}).Error("failed to create store")
		return nil, err
	}

	locks := make(map[string]store.Locker)

	ttlLocks := make(map[string]chan struct{})

	// preparing mutex locks, only used by BoltDB
	ml := make(map[string]*sync.Mutex)

	return &LibKVBackend{
		Backend:    backend,
		Addrs:      addrs,
		Options:    options,
		Store:      libstore,
		Locks:      locks,
		TTLLocks:   ttlLocks,
		MutexLocks: ml,
	}, nil
}

// Put - puts object into kv store
func (l *LibKVBackend) Put(key string, value []byte) error {
	return l.Store.Put(fmt.Sprintf("%s%s", Namespace, key), value, nil)
}

// Get - gets object from kv store
func (l *LibKVBackend) Get(key string) (*KVPair, error) {
	kvPair, err := l.Store.Get(addNamespace(key))
	if err != nil {
		// replacing error with locally defined
		if err == store.ErrKeyNotFound {
			return nil, fmt.Errorf(NotFound)
		}

		return nil, err
	}
	return &KVPair{
		Key:       removeNamespace(kvPair.Key),
		Value:     kvPair.Value,
		LastIndex: kvPair.LastIndex,
	}, nil
}

// Delete - deletes specified key
func (l *LibKVBackend) Delete(key string) error {
	return l.Store.Delete(addNamespace(key))
}

// Exists - checks whether object exists
func (l *LibKVBackend) Exists(key string) (bool, error) {
	return l.Store.Exists(addNamespace(key))
}

// List - returns all keys from specified directory
func (l *LibKVBackend) List(directory string) (kvs []*KVPair, err error) {
	pairs, err := l.Store.List(addNamespace(directory))
	if err != nil {
		// replacing error with empty slice
		if err == store.ErrKeyNotFound {
			return kvs, nil
		}
		return
	}

	for _, v := range pairs {
		kv := &KVPair{
			Key:       removeNamespace(v.Key),
			Value:     v.Value,
			LastIndex: v.LastIndex,
		}
		kvs = append(kvs, kv)
	}
	return
}

// DeleteTree - deletes specified directory
func (l *LibKVBackend) DeleteTree(directory string) error {
	return l.Store.DeleteTree(addNamespace(directory))
}

// Lock - creates new locker and locks
func (l *LibKVBackend) Lock(key string) error {
	key = addNamespace(key)
	if l.Backend == BoltDBBackend {
		_, ok := l.MutexLocks[key]
		if !ok {
			l.MutexLocks[key] = &sync.Mutex{}
		}
		l.MutexLocks[key].Lock()

		return nil
	}

	options := &store.LockOptions{}
	locker, err := l.Store.NewLock(key, options)
	if err != nil {
		return fmt.Errorf("something went wrong when trying to initialize the Lock (%s)", key)
	}
	_, err = locker.Lock(nil)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"key":   key,
		}).Error("failed to lock locker")
		return err
	}
	l.Locks[key] = locker
	return err
}

// Unlock - unlocks named lock
func (l *LibKVBackend) Unlock(key string) error {
	key = addNamespace(key)
	if l.Backend == BoltDBBackend {
		log.WithFields(log.Fields{
			"backend": l.Backend,
			"key":     key,
			"command": "unlock",
		}).Info("got request to unlock key, usin mutex locks")
		_, ok := l.MutexLocks[key]
		if !ok {
			l.MutexLocks[key] = &sync.Mutex{}
			// mutex didn't exist, nothing to do
			return nil
		}
		l.MutexLocks[key].Unlock()
		return nil
	}

	locker, ok := l.Locks[key]
	if !ok {
		return fmt.Errorf(ErrLockNotFound)
	}
	delete(l.Locks, key)
	return locker.Unlock()
}

// LockTTL - lock with ttl (time to live)
func (l *LibKVBackend) LockTTL(key string, ttl int, stopChan chan struct{}) (renew chan struct{}, stop <-chan struct{}, err error) {
	key = addNamespace(key)
	// boltdb doesn't have locks, it has transactions
	if l.Backend == BoltDBBackend {
		// using simple lock
		err = l.Lock(key)
		return nil, nil, err
	}

	renewCh := make(chan struct{})
	// TODO: this could be controller ID so other controllers would know
	// which contoller is keeping the lock
	value := []byte("ttl_lock")
	duration := time.Duration(ttl) * time.Second
	lock, err := l.Store.NewLock(key, &store.LockOptions{Value: value, TTL: duration, RenewLock: renewCh})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"key":   key,
			"ttl":   ttl,
		}).Error("failed to acquire lock")
		return nil, nil, fmt.Errorf(ErrFailedToLock)
	}
	stop, err = lock.Lock(stopChan)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"key":   key,
			"ttl":   ttl,
		}).Error("lock acquired but failed to lock it")
		return nil, nil, fmt.Errorf(ErrFailedToLock)
	}

	// adding lock to locker for later unlock
	l.Locks[key] = lock
	// adding lcok chan to ttl locks
	l.TTLLocks[key] = renewCh
	return renewCh, stop, nil

}

// ReleaseLock - stops renewing lock
func (l *LibKVBackend) ReleaseLock(key string) error {
	key = addNamespace(key)
	// boltdb doesn't have locks, it has transactions
	if l.Backend == BoltDBBackend {
		return l.Unlock(key)
	}
	// getting lock chan
	ch, ok := l.TTLLocks[key]
	if !ok {
		return fmt.Errorf(ErrLockNotFound)
	}
	log.Infof("lock found, stopping it, key: %s", key)
	delete(l.TTLLocks, key)
	close(ch)
	return nil
}

// Watch - watches for changes on a key
func (l *LibKVBackend) Watch(key string, stopCh <-chan struct{}) (<-chan *KVPair, error) {
	key = addNamespace(key)
	// boltdb doesn't have watches
	if l.Backend == BoltDBBackend {
		return nil, nil
	}

	log.Infof("Adding Watch for %s", key)
	watchCh, err := l.Store.Watch(key, stopCh)
	if err != nil {
		// replacing error with locally defined
		if err == store.ErrKeyNotFound {
			return nil, fmt.Errorf(NotFound)
		}
		return nil, err
	}
	outCh := make(chan *KVPair)
	go func() {
		select {
		case kvPair := <-watchCh:
			outCh <- &KVPair{
				Key:       removeNamespace(kvPair.Key),
				Value:     kvPair.Value,
				LastIndex: kvPair.LastIndex,
			}
		case <-stopCh:
			return
		}
	}()
	return outCh, nil
}

// WatchTree - watches for changes on directory
func (l *LibKVBackend) WatchTree(key string, stopCh <-chan struct{}) (<-chan []*KVPair, error) {
	key = addNamespace(key)
	// boltdb doesn't have watches
	if l.Backend == BoltDBBackend {
		return nil, nil
	}

	// Checking on the key before watching
	exists, err := l.Store.Exists(key)
	if err != nil {
		return nil, err
	}

	if !exists {
		// Don't forget IsDir:true if the code is used cross-backend
		err := l.Store.Put(key, []byte("placeholder"), &store.WriteOptions{IsDir: true})
		if err != nil {
			return nil, err
		}
	}

	watchCh, err := l.Store.WatchTree(key, stopCh)
	if err != nil {
		// replacing error with locally defined
		if err == store.ErrKeyNotFound {
			return nil, fmt.Errorf(NotFound)
		}
		return nil, err
	}

	outCh := make(chan []*KVPair)
	go func() {
		for {
			select {
			case pairs := <-watchCh:
				var payload []*KVPair
				for _, pair := range pairs {
					payload = append(payload, &KVPair{
						Key:       removeNamespace(pair.Key),
						Value:     pair.Value,
						LastIndex: pair.LastIndex,
					})
				}
				outCh <- payload

			case <-stopCh:
				return
			}
		}
	}()
	return outCh, nil

}
