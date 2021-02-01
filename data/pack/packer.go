package pack

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

type Action interface {
	Run(ctx *Context) error
}

func NewPacker() *Packer {
	return &Packer{
		storage: &fileStorage{},
	}
}

type Packer struct {
	actions           []Action
	hasIgnoredActions bool

	logMutex sync.Mutex

	storageMutex sync.Mutex
	storage      Storage
}

func (p *Packer) Schedule(action Action) {
	p.actions = append(p.actions, action)
}

func (p *Packer) XSchedule(action Action) {
	p.hasIgnoredActions = true
}

func (p *Packer) Run(workerCount int) error {
	group, groupCtx := errgroup.WithContext(context.Background())

	actionChan := make(chan Action, workerCount)
	group.Go(func() error {
		return p.produceActions(groupCtx, actionChan)
	})
	for i := 0; i < workerCount; i++ {
		group.Go(func() error {
			return p.processActions(groupCtx, actionChan)
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}
	return nil
}

func (p *Packer) produceActions(ctx context.Context, out chan<- Action) error {
	defer close(out)
	for _, action := range p.actions {
		select {
		case out <- action:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (p *Packer) processActions(ctx context.Context, in <-chan Action) error {
	for {
		select {
		case action, ok := <-in:
			if !ok {
				return nil
			}
			if err := p.processAction(action); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Packer) processAction(action Action) error {
	ctx := &Context{
		logMutex:     &p.logMutex,
		storageMutex: &p.storageMutex,
		storage:      p.storage,
	}
	if err := action.Run(ctx); err != nil {
		return fmt.Errorf("failed to process action: %w", err)
	}
	if p.hasIgnoredActions {
		return fmt.Errorf("failing on purpose due to ignored actions")
	}
	return nil
}
