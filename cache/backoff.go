package cache

import (
	"fmt"
	"math"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rickslab/ares/errcode"
	"google.golang.org/grpc/status"
)

var (
	levelExpireIn = 24 * time.Hour
)

type Backoff struct {
	field  string
	failed bool
}

func NewBackoff(field string) (*Backoff, error) {
	err := checkBackoff(field)
	if err != nil {
		return nil, err
	}

	return &Backoff{
		field:  field,
		failed: false,
	}, nil
}

func checkBackoff(field string) error {
	ttl, err := redis.Int64(Redis("backoff").Do("TTL", fmt.Sprintf("%s:locked", field)))
	if err != nil {
		return err
	}

	if ttl > 0 {
		return status.Error(errcode.ErrRequestBackoff, fmt.Sprintf("retry after %d", ttl))
	}
	return nil
}

func (b *Backoff) Failed() {
	b.failed = true
}

func (b *Backoff) Close() error {
	conn := Redis("backoff").Get()
	defer conn.Close()

	levelKey := fmt.Sprintf("%s:level", b.field)
	lockedKey := fmt.Sprintf("%s:locked", b.field)

	if b.failed {
		level, err := redis.Int(conn.Do("INCR", levelKey))
		if err != nil && err != redis.ErrNil {
			return err
		}

		if level == 1 {
			conn.Do("EXPIRE", levelKey, levelExpireIn.Seconds())
		}

		lockedTTL := int64(10 * math.Pow(2, float64(level)))

		_, err = conn.Do("SET", lockedKey, 1, "EX", lockedTTL)
		return err
	}

	conn.Do("DEL", levelKey)
	conn.Do("DEL", lockedKey)
	return nil
}
