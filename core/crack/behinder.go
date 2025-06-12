package crack

type BehinderCrack struct {
	CrackBase
}

func (b *BehinderCrack) Crack() (succ bool, err error) {

	return
}

func (*BehinderCrack) Class() string {
	return CLASS_WEBSHELL
}