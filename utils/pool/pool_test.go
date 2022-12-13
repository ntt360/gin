package pool

import (
	"fmt"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	p := NewPool(&Options{
		Size:        2,
		IdleTimeout: time.Second * 3,
	})

	p.Submit(func() {
		fmt.Println("task 1 start")
		time.Sleep(time.Second * 10)
		fmt.Println("task end")
	})

	p.Submit(func() {
		fmt.Println("task 2 start")
		time.Sleep(time.Second * 10)
		fmt.Println("task end")
	})

	time.Sleep(time.Minute * 5)
}
