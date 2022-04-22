package cache

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rickslab/ares/errcode"
	"google.golang.org/grpc/status"
)

type Secure struct {
	key      string
	count    int
	duration time.Duration
}

func NewSecure(key string, maxCount int, duration time.Duration) (*Secure, error) {
	conn := Redis("secure").Get()
	defer conn.Close()

	count, err := redis.Int(conn.Do("GET", key))
	if err != nil && err != redis.ErrNil {
		return nil, err
	}
	if count >= maxCount {
		ttl, err := redis.Int(conn.Do("TTL", key))
		if err != nil && err != redis.ErrNil {
			return nil, err
		}

		return nil, status.Errorf(errcode.ErrRequestBackoff, "retry after : %d", ttl)
	}

	return &Secure{
		key:      key,
		count:    count,
		duration: duration,
	}, nil
}

func (s *Secure) Failed() {
	conn := Redis("secure").Get()
	defer conn.Close()

	conn.Do("INCR", s.key)
	if s.count == 0 {
		conn.Do("EXPIRE", s.key, s.duration.Nanoseconds()/int64(time.Second), "NX")
	}
}
