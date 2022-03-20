package component

type Controller interface {
	Subscribe(callback ControllerCallback) ControllerSubscription
	Alter(func() error) error
	NotifyChanged()
}

type ControllerCallback func(controller Controller)

type ControllerSubscription interface {
	Unsubscribe()
}

func NewBaseController() Controller {
	return &BaseController{}
}

var _ Controller = (*BaseController)(nil)

type BaseController struct {
	subscriptions  []*baseSubscription
	ongoingChanges int
}

func (c *BaseController) Subscribe(callback ControllerCallback) ControllerSubscription {
	subscription := &baseSubscription{
		controller: c,
		callback:   callback,
	}
	c.subscriptions = append(c.subscriptions, subscription)
	return subscription
}

func (c *BaseController) Alter(fn func() error) error {
	defer func() {
		if c.ongoingChanges == 0 {
			c.NotifyChanged()
		}
	}()
	c.ongoingChanges++
	defer func() {
		c.ongoingChanges--
	}()
	return fn()
}

func (c *BaseController) NotifyChanged() {
	if c.ongoingChanges > 0 {
		return
	}
	for _, subscription := range c.subscriptions {
		subscription.callback(c)
	}
}

func (c *BaseController) removeSubscription(subscription *baseSubscription) {
	index := c.findSubscription(subscription)
	if index >= 0 {
		c.subscriptions = append(c.subscriptions[:index], c.subscriptions[index+1:]...)
	}
}

func (c *BaseController) findSubscription(subscription *baseSubscription) int {
	for i, candidate := range c.subscriptions {
		if candidate == subscription {
			return i
		}
	}
	return -1
}

type baseSubscription struct {
	controller *BaseController
	// TODO: Use linked list and reference reuse
	callback ControllerCallback
}

func (s *baseSubscription) Unsubscribe() {
	s.controller.removeSubscription(s)
}
