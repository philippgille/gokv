package Client

import (
	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/redis"
)

type Storage struct {
	Redis interface {
		GetClient(option redis.Options) (gokv.Store, error)
	}
}

func NewStorage() Storage {
	return Storage{
		Redis: &RedisStore{},
	}
}
