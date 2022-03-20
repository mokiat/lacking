package async

type Task func() Result

func VoidTask(callback func() error) Task {
	return func() Result {
		return Result{
			Err: callback(),
		}
	}
}

type Result struct {
	Value interface{}
	Err   error
}
