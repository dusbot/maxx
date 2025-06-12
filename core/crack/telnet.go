package crack

import "time"

type TelnetCracker struct {
	CrackBase
}

func (t *TelnetCracker) Ping() (succ bool, err error) {
	var timeout = 3
	if t.Timeout > 0 {
		timeout = t.Timeout
	}
	c, err := NewClient(t.Target, "", "", time.Duration(timeout)*time.Second)
	if err != nil {
		return false, err
	}
	defer c.close()
	err = c.Login()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (t *TelnetCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if t.Timeout > 0 {
		timeout = t.Timeout
	}
	c, err := NewClient(t.Target, t.User, t.Pass, time.Duration(timeout)*time.Second)
	if err != nil {
		return false, err
	}
	defer c.close()
	err = c.Login()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (*TelnetCracker) Class() string {
	return CLASS_REMOTE_ACCESS
}
