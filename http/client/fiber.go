package client

var _ FiberFactory = (*fiber)(nil)

type FiberFactory interface {
	V2() WebClient
	V3() WebClient
}

type fiber struct {
	opts Options
}

func NewFiber(opts ...Option) FiberFactory {
	return &fiber{opts: opts}
}

func (c *fiber) V2() WebClient {
	return NewFiberV2(c.opts...)
}

func (c *fiber) V3() WebClient {
	return NewFiberV3(c.opts...)
}
