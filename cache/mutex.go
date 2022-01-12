package cache

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/rickslab/ares/errcode"
	"google.golang.org/grpc/status"
)

const (
	mutexDefaultLockTimeout = 6 * time.Second
	mutexDefaultLeaseTime   = 20 * time.Second
	backoffTime             = 1 * time.Millisecond
)

type Mutex struct {
	conn  redis.Conn
	key   string
	value string
	opt   mutexOption
}

type mutexOption struct {
	lockTimeout time.Duration
	leaseTime   time.Duration
}

type MutexOption func(*mutexOption)

func NewMutex(key string, opts ...MutexOption) *Mutex {
	m := &Mutex{
		conn:  Redis("mutex").Get(),
		key:   key,
		value: uuid.New().String(),
		opt: mutexOption{
			lockTimeout: mutexDefaultLockTimeout,
			leaseTime:   mutexDefaultLeaseTime,
		},
	}

	for _, opt := range opts {
		opt(&m.opt)
	}
	return m
}

func WithLockTimeout(lockTimeout time.Duration) MutexOption {
	return func(opt *mutexOption) {
		opt.lockTimeout = lockTimeout
	}
}

func WithLeaseTime(leaseTime time.Duration) MutexOption {
	return func(opt *mutexOption) {
		opt.leaseTime = leaseTime
	}
}

func (m *Mutex) Close() {
	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
	}
}

func (m *Mutex) TryLock() error {
	_, err := redis.String(m.conn.Do("SET", m.key, m.value, "EX", m.opt.leaseTime.Seconds()+2, "NX"))
	if err == redis.ErrNil {
		return status.Errorf(errcode.ErrMutexLock, "lock '%s' failed", m.key)
	}
	return err
}

func (m *Mutex) Lock() error {
	total := time.Duration(0)
	s := backoffTime
	for {
		err := m.TryLock()
		if err == nil {
			return nil
		}
		if code, _ := errcode.From(err); code != errcode.ErrMutexLock {
			return err
		}
		if total > m.opt.lockTimeout {
			return err
		}

		time.Sleep(s)
		total += s
		s *= 2
	}
}

func (m *Mutex) Unlock() error {
	value, err := redis.String(m.conn.Do("GET", m.key))
	if err != nil {
		return err
	}
	if value != m.value {
		return status.Errorf(errcode.ErrMutexUnlock, "unlock '%s' failed", m.key)
	}

	ttl, err := redis.Int(m.conn.Do("TTL", m.key))
	if err != nil {
		return err
	}

	if ttl > 1 {
		_, err = m.conn.Do("DEL", m.key)
		return err
	}

	return nil
}

func MutexWrap(key string, f func() (interface{}, error)) (reply interface{}, err error) {
	m := NewMutex(key)
	defer m.Close()

	err = m.Lock()
	if err != nil {
		return
	}

	defer func() {
		mErr := m.Unlock()
		if err == nil && mErr != nil {
			err = mErr
		}
	}()

	return f()
}
