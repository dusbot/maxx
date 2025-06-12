package utcp

import (
	"net"
	"time"

	"github.com/dusbot/maxx/libs/urandom"
	"golang.org/x/net/proxy"
)

type dialer struct{}

func NewDialer() *dialer {
	return &dialer{}
}

func (d *dialer) Dial(addr string, timeout int, proxies ...string) (conn net.Conn, err error) {
	if len(proxies) == 0 || proxies[0] == "" {
		return net.DialTimeout("tcp", addr, time.Duration(timeout)*time.Second)
	}
	proxyAddr := proxies[urandom.RandInt(0, len(proxies)-1)] //select one randomly for now
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, &net.Dialer{
		Timeout: time.Duration(timeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return dialer.Dial("tcp", addr)
}
