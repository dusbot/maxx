package crack

import (
	"errors"

	"github.com/dusbot/maxx/libs/grdp"
)

type RdpCracker struct {
	CrackBase
}

func (f *RdpCracker) Ping() (succ bool, err error) {
	return false, errors.ErrUnsupported
}

func (f *RdpCracker) Crack() (succ bool, err error) {
	user, domain := SplitUserDomain(f.User)
	err = grdp.Login(f.Target, domain, user, f.Pass)
	if err != nil {
		return false, err
	}
	return true, nil
}
