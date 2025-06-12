package crack

import (
	"strconv"
	"time"

	"github.com/knadh/go-pop3"
)

// fixme: Service probe before ping maybe, if the dest tcp server is not pop3 server, the Ping or Crack will block forever
// I will be back later
type Pop3Cracker struct {
	CrackBase
}


func (f *Pop3Cracker) Ping() (succ bool, err error) {
	var port int
	if port, err = strconv.Atoi(f.Port); err != nil {
		return
	}
	p := pop3.New(pop3.Opt{
		Host:          f.Ip,
		Port:          port,
		DialTimeout:   time.Duration(f.Timeout) * time.Second,
		TLSSkipVerify: true,
	})
	c, err := p.NewConn()
	if err != nil {
		return
	}
	defer c.Quit()

	if err = c.Noop(); err != nil {
		return
	}
	return true, nil
}

func (f *Pop3Cracker) Crack() (succ bool, err error) {
	var port int
	if port, err = strconv.Atoi(f.Port); err != nil {
		return
	}
	p := pop3.New(pop3.Opt{
		Host:          f.Ip,
		Port:          port,
		DialTimeout:   time.Duration(f.Timeout) * time.Second,
		TLSSkipVerify: true,
	})
	c, err := p.NewConn()
	if err != nil {
		return
	}
	defer c.Quit()
	if err = c.Auth(f.User, f.Pass); err != nil {
		return
	}
	if err = c.Noop(); err != nil {
		return
	}
	return true, nil
}

func (*Pop3Cracker) Class() string {
	return CLASS_EMAIL
}
