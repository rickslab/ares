package errcode

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func ErrorMap(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(ErrRecordNotFound, "record not found")
	}

	if myErr, ok := err.(*mysql.MySQLError); ok && myErr.Number == 1062 {
		return status.Error(ErrKeyDuplicated, "key duplicated")
	}

	switch err {
	case redis.ErrNil:
		err = status.Error(ErrValueNotFound, "value not found")
	}
	return err
}
