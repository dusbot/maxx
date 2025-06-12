package crack

type RsyncCracker struct {
	CrackBase
}

func (s *RsyncCracker) Ping() (succ bool, err error) {
	var timeout = 3
	if s.Timeout > 0 {
		timeout = s.Timeout
	}
	ver, modules, err := RsyncDetect(s.Target, timeout)
	if err != nil {
		return false, err
	}
	err = RsyncUnauth(s.Target, ver, modules, timeout)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *RsyncCracker) Crack() (succ bool, err error) {
	var timeout = 3
	if s.Timeout > 0 {
		timeout = s.Timeout
	}
	ver, modules, err := RsyncDetect(s.Target, timeout)
	if err != nil {
		return
	}
	err = RsyncLogin(s.Target, s.User, s.Pass, ver, modules, timeout)
	if err != nil {
		return
	}
	return true, nil
}

func (*RsyncCracker) Class() string {
	return CLASS_FILE_TRANSFER
}
