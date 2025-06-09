package crack

import (
	"errors"
	"strings"

	"github.com/gosnmp/gosnmp"
)

type SnmpCracker struct {
	CrackBase
}

func (s *SnmpCracker) Ping() (succ bool, err error) {
	cli := gosnmp.Default
	cli.Target = s.Target
	if strings.Contains(s.Target, ":") {
		targetSplit := strings.Split(s.Target, ":")
		if targetSplit[1] != "161" {
			return false, errors.ErrUnsupported
		}
		cli.Target = targetSplit[0]
	}
	cli.Version = gosnmp.Version2c
	err = cli.Connect()
	if err == nil {
		defer cli.Conn.Close()
		return true, nil
	}
	return false, err
}

func (s *SnmpCracker) Crack() (succ bool, err error) {
	cli := gosnmp.Default
	cli.Target = s.Target
	if strings.Contains(s.Target, ":") {
		targetSplit := strings.Split(s.Target, ":")
		if targetSplit[1] != "161" {
			return false, errors.ErrUnsupported
		}
		cli.Target = targetSplit[0]
	}
	cli.Community = s.Pass
	cli.Version = gosnmp.Version2c
	err = cli.Connect()
	if err == nil {
		defer cli.Conn.Close()
		return true, nil
	}
	return false, err
}
