package attack

type Plugin interface {
	Name() string
	Attack(Input) (Output, error)
}