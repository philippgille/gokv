package Client

import (
	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/redis"
)

type RedisStore Storage

func (r *RedisStore) GetClient(option redis.Options) (gokv.Store, error) {
	client, err := redis.NewClient(option)
	if err != nil {
		return nil, err
	}
	return client, err
}
