package crack

import (
	"errors"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
)

type LdapCracker struct {
	CrackBase
}

func (l *LdapCracker) Ping() (succ bool, err error) {
	var timeout = 3
	if l.Timeout > 0 {
		timeout = l.Timeout
	}
	ldap.DefaultTimeout = time.Second * time.Duration(timeout)
	conn, err := ldap.Dial("tcp", l.Target)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	return false, errors.ErrUnsupported
}

func (l *LdapCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if l.Timeout > 0 {
		timeout = l.Timeout
	}
	ldap.DefaultTimeout = time.Second * time.Duration(timeout)
	conn, err := ldap.Dial("tcp", l.Target)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	err = conn.Bind(l.User, l.Pass)
	if err == nil {
		return true, nil
	}
	return
}

func (*LdapCracker) Class() string {
	return CLASS_FILE_TRANSFER
}
