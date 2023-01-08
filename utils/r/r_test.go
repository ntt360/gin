package r_test

import (
	"github.com/ntt360/gin/utils/r"
	"testing"
)

func TestGo(t *testing.T) {
	// run goroutine
	r.Go(func() {

	})

	// run goroutine and with panic err callback
	r.Go(func() {

	}, r.WithErrCallbackOpt(func(err any) {

	}))
}

