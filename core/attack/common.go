package attack

import "github.com/dusbot/maxx/core/types"

type (
	Input interface {
		GetTask() *types.Task
		SetTask(task *types.Task) error
	}

	Output interface {
		GetResult() *types.Result
		SetResult(result *types.Result) error
	}

	task struct {
		innerTask *types.Task
	}

	result struct {
		innerResult *types.Result
	}

	IAttack[in, out any] interface {
		Attack(in) (out, error)
	}
)

func NewTask() *task {
	return new(task)
}

func NewResult() *result {
	return new(result)
}

func (t *task) GetTask() *types.Task {
	return t.innerTask
}

func (t *task) SetTask(task *types.Task) error {
	t.innerTask = task
	return nil
}

func (r *result) GetResult() *types.Result {
	return r.innerResult
}

func (r *result) SetResult(result *types.Result) error {
	r.innerResult = result
	return nil
}
