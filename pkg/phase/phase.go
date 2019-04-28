package phase

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/apex/log"
)

type Phase struct {
	cancelWithSignals []os.Signal
	duration          time.Duration
	complete          bool
	ctx               context.Context
}

func NewPhase(ctx context.Context) *Phase {
	return &Phase{
		ctx:               ctx,
		cancelWithSignals: []os.Signal{},
		duration:          0,
		complete:          false,
	}
}

func (p *Phase) CancelWithSignal(sig os.Signal) *Phase {
	p.cancelWithSignals = append(p.cancelWithSignals, sig)
	return p
}

func (p *Phase) WithTimeout(duration time.Duration) *Phase {
	p.duration = duration
	return p
}

func (p *Phase) Run(fn func(ctx context.Context) error) error {
	if len(p.cancelWithSignals) > 0 {
		signalChan := make(chan os.Signal, 1)
		for _, sig := range p.cancelWithSignals {
			signal.Notify(signalChan, sig)
		}

		newCtx := p.ctx
		var cancel context.CancelFunc
		if p.duration > 0 {
			newCtx, cancel = context.WithTimeout(p.ctx, p.duration)
		} else {
			newCtx, cancel = context.WithCancel(p.ctx)
		}

		defer func() {
			signal.Stop(signalChan)
			cancel()
		}()

		go func() {
			select {
			case <-signalChan:
				log.Infof("Cancelling phase")
				cancel()
			case <-newCtx.Done():
			}
		}()

		return fn(newCtx)
	}

	if p.duration > 0 {
		newCtx, cancelFunc := context.WithTimeout(p.ctx, p.duration)
		defer cancelFunc()

		return fn(newCtx)
	}

	return fn(p.ctx)
}
