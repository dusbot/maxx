package crack

import (
	"net"
	"time"

	"github.com/mitchellh/go-vnc"
)

type VncCracker struct {
	CrackBase
}

func (v *VncCracker) Ping() (succ bool, err error) {
	var timeout = 3
	if v.Timeout > 0 {
		timeout = v.Timeout
	}
	target := v.Target

	tcpconn, err := net.DialTimeout("tcp", target, time.Duration(timeout)*time.Second)
	if err != nil {
		return false, err
	}

	config := vnc.ClientConfig{
		Auth: []vnc.ClientAuth{
			&vnc.PasswordAuth{Password: ""},
		},
	}
	conn, err := vnc.Client(tcpconn, &config)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	return true, nil
}

func (v *VncCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if v.Timeout > 0 {
		timeout = v.Timeout
	}
	target := v.Target

	tcpconn, err := net.DialTimeout("tcp", target, time.Duration(timeout)*time.Second)
	if err != nil {
		return false, err
	}

	config := vnc.ClientConfig{
		Auth: []vnc.ClientAuth{
			&vnc.PasswordAuth{Password: v.Pass},
		},
	}
	conn, err := vnc.Client(tcpconn, &config)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	return true, nil
}

func (*VncCracker) Class() string {
	return CLASS_REMOTE_ACCESS
}
