package client

var _ Factory = (*factory)(nil)

type Factory interface {
	WebClient
	Fiber() FiberFactory
}

type factory struct {
	WebClient
	opts Options
}

func New(opts ...Option) Factory {
	return &factory{
		WebClient: NewNetHttp(opts...),
		opts:      opts,
	}
}

func (c *factory) Fiber() FiberFactory {
	return NewFiber(c.opts...)
}
