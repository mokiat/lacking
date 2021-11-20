package async

func SuccessfulEventual() Eventual {
	return ImmediateEventual(nil)
}

func FailedEventual(err error) Eventual {
	return ImmediateEventual(err)
}

func ImmediateEventual(err error) Eventual {
	ch := make(chan error, 1)
	ch <- err
	return Eventual{
		ch: ch,
	}
}

func NewEventual() (Eventual, func(error)) {
	ch := make(chan error, 1)
	return Eventual{
			ch: ch,
		},
		func(err error) {
			ch <- err
			ch = nil
		}
}

type Eventual struct {
	ch chan error
}

func (e Eventual) Wait() error {
	err := <-e.ch
	e.ch <- err
	return err
}

func (e Eventual) Done() bool {
	select {
	case err := <-e.ch:
		e.ch <- err
		return true
	default:
		return false
	}
}

func NewCompositeEventual(eventuals ...Eventual) Eventual {
	eventual, eventualDone := NewEventual()
	go func() {
		var eventualErr error
		for _, e := range eventuals {
			if err := e.Wait(); err != nil {
				eventualErr = err
			}
		}
		eventualDone(eventualErr)
	}()
	return eventual
}
