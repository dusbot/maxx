package crack

import (
	"time"

	"github.com/jlaffaye/ftp"
)

type FtpCracker struct {
	CrackBase
}

func (f *FtpCracker) Ping() (succ bool, err error) {
	var timeout = 3
	if f.Timeout > 0 {
		timeout = f.Timeout
	}
	conn, err := ftp.DialTimeout(f.Target, time.Duration(timeout)*time.Second)
	if err != nil {
		return false, err
	}
	defer conn.Quit()
	err = conn.Login("anonymous", "")
	if err != nil {
		return false, err
	}
	return true, nil
}

func (f *FtpCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if f.Timeout > 0 {
		timeout = f.Timeout
	}
	conn, err := ftp.DialTimeout(f.Target, time.Duration(timeout)*time.Second)
	if err != nil {
		return false, err
	}
	defer conn.Quit()
	err = conn.Login(f.User, f.Pass)
	if err != nil {
		return false, err
	}
	return true, nil
}
