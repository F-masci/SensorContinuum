package utils

import "sync"

type PauseSignal struct {
	mu sync.Mutex
	ch chan bool
}

func NewPauseSignal() *PauseSignal {
	return &PauseSignal{ch: make(chan bool, 1)}
}

func (p *PauseSignal) Send(v bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	select {
	case p.ch <- v:
	default:
		<-p.ch
		p.ch <- v
	}
}

func (p *PauseSignal) Chan() <-chan bool {
	return p.ch
}
