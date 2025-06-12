package crack

type ActivemqCracker struct {
	CrackBase
}

func (f *ActivemqCracker) Crack() (succ bool, err error) {
	return
}

func (*ActivemqCracker) Class() string {
	return CLASS_MQ_MIDDLEWARE
}
