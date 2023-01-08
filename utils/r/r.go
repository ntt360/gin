package r

type Option func(g *gr)

// WithErrCallbackOpt setting panic err callback function
func WithErrCallbackOpt(f func(any)) Option {
	return func(g *gr) {
		g.errCallback = f
	}
}

type gr struct {
	errCallback func(any)
}

// Go run go routine
func Go(f func(), opts ...Option) {
	g := &gr{}
	for _, opt := range opts {
		opt(g)
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				if g.errCallback != nil {
					g.errCallback(err)
				}
			}
		}()

		f()
	}()
}
