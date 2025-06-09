package crack

import (
	"errors"

	"github.com/dusbot/maxx/libs/C-Sto/goWMIExec/pkg/wmiexec"
)

type WmiCracker struct {
	CrackBase
}

func (t *WmiCracker) Ping() (succ bool, err error) {
	return false, errors.ErrUnsupported
}

func (t *WmiCracker) Crack() (succ bool, err error) {
	return WMIExec(t.Target, t.User, t.Pass, "", "", "", "", nil)
}

func WMIExec(target, username, password, hash, domain, clientHostname, binding string, cfgIn *wmiexec.WmiExecConfig) (flag bool, err error) {
	if cfgIn == nil {
		cfg, err1 := wmiexec.NewExecConfig(username, password, hash, domain, target, clientHostname, true, nil, nil)
		if err1 != nil {
			err = err1
			return
		}
		cfgIn = &cfg
	}
	execer := wmiexec.NewExecer(cfgIn)
	err = execer.Auth()
	if err != nil {
		return
	}
	flag = true
	return
}
