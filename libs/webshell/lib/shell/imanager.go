package shell

type IManager interface {
	Ping(p ...IParams) (bool, error)
	BasicInfo(p ...IParams) (IResult, error)
	CommandExec(p IParams) (IResult, error)
	FileManagement(p IParams) (IResult, error)
	DatabaseManagement(p IParams) (IResult, error)
}
