package util

import (
	"log"
	"time"
)

func Retry(f func() error, times int, interval, dt time.Duration) (err error) {
	for i := 0; i < times; i++ {
		if i > 0 {
			time.Sleep(interval)
			interval += dt
			log.Printf("Retry %d time(s)\n", i)
		}
		err = f()
		if err == nil {
			return
		}
	}
	return
}
