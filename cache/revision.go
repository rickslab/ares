package cache

import (
	"fmt"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

type GetRevisionHandler func(id int64) (int64, error)

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

func (r *Revision) existRevision(id int64) (bool, error) {
	return redis.Bool(r.conn.Do("EXISTS", id))
}

func (r *Revision) mgetRevision(ids []int64) ([]int64, error) {
	idis := make([]interface{}, len(ids))
	for i, id := range ids {
		idis[i] = id
	}

	strs, err := redis.Strings(r.conn.Do("MGET", idis...))
	if err != nil {
		return nil, err
	}
	result := make([]int64, len(strs))
	for i, str := range strs {
		if str == "" {
			result[i] = -1 // nil items turn to -1
		} else {
			n, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return nil, err
			}
			result[i] = n
		}
	}
	return result, nil
}

func (r *Revision) incrRevision(id int64) (int64, error) {
	return redis.Int64(r.conn.Do("INCR", id))
}

func (r *Revision) setRevision(id int64, revision int64) error {
	_, err := r.conn.Do("SET", id, revision)
	return err
}

func (r *Revision) setRevisionNX(id int64, revision int64) error {
	_, err := r.conn.Do("SET", id, revision, "NX")
	return err
}

func (r *Revision) delRevision(id int64) error {
	_, err := r.conn.Do("DEL", id)
	return err
}

func (r *Revision) nextRevision(id int64) (int64, error) {
	exists, err := r.existRevision(id)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	return r.incrRevision(id)
}

func (r *Revision) NextRevision(id int64) (int64, error) {
	rev, err := r.nextRevision(id)
	if err != nil {
		return 0, err
	}
	if rev > 0 {
		return rev, nil
	}

	m := NewMutex(fmt.Sprintf("NextRevision:%s:%d", r.name, id))
	defer m.Close()

	err = m.Lock()
	if err != nil {
		return 0, err
	}
	defer m.Unlock()

	rev, err = r.nextRevision(id)
	if err != nil {
		return 0, err
	}
	if rev > 0 {
		return rev, nil
	}

	rev, err = r.getRevision(id)
	if err != nil {
		return 0, err
	}
	rev++

	_ = r.setRevision(id, rev)
	return rev, nil
}

func (r *Revision) FindRevision(ids []int64) (map[int64]int64, error) {
	revisions, err := r.mgetRevision(ids)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]int64, len(ids))
	for i, rev := range revisions {
		id := ids[i]
		if rev == -1 {
			rev, err = r.getRevision(id)
			if err != nil {
				return nil, err
			}

			_ = r.setRevisionNX(id, rev)
		}
		result[id] = rev
	}
	return result, nil
}
