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

func NewCompositeOutcome(outcomes ...Outcome) Outcome {
	outcome := NewOutcome()
	go func() {
		var err error
		for _, outcome := range outcomes {
			if result := outcome.Wait(); result.Err != nil {
				err = result.Err
			}
		}
		outcome.Record(Result{
			Err: err,
		})
	}()
	return outcome
}

func NewOutcome() Outcome {
	return Outcome{
		ch: make(chan Result, 1),
	}
}

func NewValueOutcome(value interface{}) Outcome {
	outcome := NewOutcome()
	outcome.Record(Result{
		Value: value,
	})
	return outcome
}

func NewErrorOutcome(err error) Outcome {
	outcome := NewOutcome()
	outcome.Record(Result{
		Err: err,
	})
	return outcome
}

type Outcome struct {
	ch chan Result
}

func (o Outcome) Record(result Result) {
	o.ch <- result
}

func (o Outcome) OnSuccess(callback func(value interface{})) Outcome {
	outcome := NewOutcome()
	go func() {
		result := o.Wait()
		if result.Err == nil {
			callback(result.Value)
		}
		outcome.Record(result)
	}()
	return outcome
}

func (o Outcome) OnError(callback func(err error)) Outcome {
	outcome := NewOutcome()
	go func() {
		result := o.Wait()
		if result.Err != nil {
			callback(result.Err)
		}
		outcome.Record(result)
	}()
	return outcome
}

func (o Outcome) IsAvailable() bool {
	if o.ch == nil {
		return false
	}
	select {
	case result := <-o.ch:
		o.ch <- result
		return true
	default:
		return false
	}
}

func (o Outcome) Wait() Result {
	result := <-o.ch
	o.ch <- result
	return result
}
