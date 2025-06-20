package attack

type ping struct {
}

func (p *ping) Name() string {
	return "[ATK-Ping]"
}

func (p *ping) Attack(in Input) (out Output, err error) {

	return
}
