package server

import "sync"

type TaskRunner interface {
	Runner(wg *sync.WaitGroup)
}
