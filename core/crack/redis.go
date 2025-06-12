package crack

import (
	"time"

	"github.com/go-redis/redis"
)

type RedisCracker struct {
	CrackBase
}

func (r *RedisCracker) Ping() (succ bool, err error) {
	var timeout = 3
	if r.Timeout > 0 {
		timeout = r.Timeout
	}
	opt := redis.Options{Addr: r.Target,
		Password: "", DB: 0, DialTimeout: time.Second * time.Duration(timeout)}
	client := redis.NewClient(&opt)
	_, err = client.Ping().Result()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *RedisCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if r.Timeout > 0 {
		timeout = r.Timeout
	}
	opt := redis.Options{Addr: r.Target,
		Password: r.Pass, DB: 0, DialTimeout: time.Second * time.Duration(timeout)}
	client := redis.NewClient(&opt)
	_, err = client.Ping().Result()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (*RedisCracker) Class() string {
	return CLASS_DB
}
