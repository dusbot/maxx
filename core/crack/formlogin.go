package crack

type FormLoginCracker struct {
	CrackBase
}

// to be implemented
func (f *FormLoginCracker) Crack() (succ bool, err error) {

	return
}

func (*FormLoginCracker) Class() string {
	return CLASS_WEB
}
