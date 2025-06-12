package crack

type GodzillaCrack struct {
	CrackBase
}

func (g *GodzillaCrack) Crack() (succ bool, err error) {
	return
}

func (*GodzillaCrack) Class() string {
	return CLASS_WEBSHELL
}