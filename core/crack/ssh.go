package crack

import (
	"errors"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type SshCracker struct {
	CrackBase
}

func (s *SshCracker) Ping() (succ bool, err error) {
	return false, errors.ErrUnsupported
}

func (f *SshCracker) Crack() (succ bool, err error) {
	config := &ssh.ClientConfig{
		User: f.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(f.Pass),
		},
		Timeout: time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	client, err := ssh.Dial("tcp", f.Target, config)
	if err == nil {
		defer client.Close()
		session, err := client.NewSession()
		errRet := session.Run("echo max")
		if err == nil && errRet == nil {
			defer session.Close()
			return true, nil
		} else {
			return false, errRet
		}
	} else {
		return
	}
}
