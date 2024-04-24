package cache

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

type GetRevisionHandler func(ids ...int64) (int64, error)

type Revision struct {
	name        string
	conn        redis.Conn
	getRevision GetRevisionHandler
}

func NewRevision(name string, getRevision GetRevisionHandler) *Revision {
	return &Revision{
		conn:        Redis(name).Get(),
		getRevision: getRevision,
	}
}

func (r *Revision) Close() {
	r.conn.Close()
}

func (r *Revision) getKey(ids ...int64) string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(strs, ":")
}

func (r *Revision) existRevision(key string) (bool, error) {
	return redis.Bool(r.conn.Do("EXISTS", key))
}

func (r *Revision) incrRevision(key string) (int64, error) {
	return redis.Int64(r.conn.Do("INCR", key))
}

func (r *Revision) setRevision(key string, revision int64) error {
	_, err := r.conn.Do("SET", key, revision)
	return err
}

func (r *Revision) setRevisionNX(key string, revision int64) error {
	_, err := r.conn.Do("SET", key, revision, "NX")
	return err
}

func (r *Revision) nextRevision(key string) (int64, error) {
	exists, err := r.existRevision(key)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	return r.incrRevision(key)
}

func (r *Revision) DelRevision(ids ...int64) error {
	key := r.getKey(ids...)
	_, err := r.conn.Do("DEL", key)
	return err
}

func (r *Revision) NextRevision(ids ...int64) (int64, error) {
	key := r.getKey(ids...)
	rev, err := r.nextRevision(key)
	if err != nil {
		return 0, err
	}
	if rev > 0 {
		return rev, nil
	}

	m := NewMutex(fmt.Sprintf("NextRevision:%s:%s", r.name, key))
	defer m.Close()

	err = m.Lock()
	if err != nil {
		return 0, err
	}
	defer m.Unlock()

	rev, err = r.nextRevision(key)
	if err != nil {
		return 0, err
	}
	if rev > 0 {
		return rev, nil
	}

	rev, err = r.getRevision(ids...)
	if err != nil {
		return 0, err
	}
	rev++

	_ = r.setRevision(key, rev)
	return rev, nil
}
