package crack

import (
	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedCracker struct {
	CrackBase
}

func (r *MemcachedCracker) Ping() (succ bool, err error) {
	client := memcache.New(r.Target)
	err = client.Ping()
	if err != nil {
		if _, ok := err.(*memcache.ConnectTimeoutError); ok {
			return false, ERR_CONNECTION
		}
		return false, err
	}
	defer client.Close()
	return true, nil
}

func (r *MemcachedCracker) Crack() (succ bool, err error) {
	client := memcache.New(r.Target)
	if err = client.Ping(); err != nil {
		return false, err
	}
	defer client.Close()
	return true, nil
}

func (*MemcachedCracker) Class() string {
	return CLASS_DB
}
