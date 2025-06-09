package crack

import (
	"net"
	"strings"
	"time"

	"github.com/hirochachacha/go-smb2"
)

type SmbCracker struct {
	CrackBase
}

func (s *SmbCracker) Ping() (succ bool, err error) {
	var timeout = 3
	if s.Timeout > 0 {
		timeout = s.Timeout
	}
	user, domain := SplitUserDomain(s.User)

	dialer := &smb2.Dialer{}
	dialer.Initiator = &smb2.NTLMInitiator{
		User:     user,
		Domain:   domain,
		Password: "",
	}

	c, err := net.DialTimeout("tcp", s.Target, time.Second*time.Duration(timeout))
	if err != nil {
		return false, err
	}

	conn, err := dialer.Dial(c)
	if err != nil {
		return false, err
	}
	_, err = conn.ListSharenames()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SmbCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if s.Timeout > 0 {
		timeout = s.Timeout
	}
	user, domain := SplitUserDomain(s.User)

	dialer := &smb2.Dialer{}
	dialer.Initiator = &smb2.NTLMInitiator{
		User:     user,
		Domain:   domain,
		Password: s.Pass,
	}
	c, err := net.DialTimeout("tcp", s.Target, time.Second*time.Duration(timeout))
	if err != nil {
		return false, err
	}

	conn, err := dialer.Dial(c)
	if err != nil {
		return false, err
	}
	_, err = conn.ListSharenames()
	if err != nil {
		return false, err
	}
	return true, nil
}

func SplitUserDomain(user string) (string, string) {
	var domain string
	if strings.Contains(user, "/") {
		user = strings.Split(user, "/")[1]
		domain = strings.Split(user, "/")[0]
	}
	return user, domain
}
