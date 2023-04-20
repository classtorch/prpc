package wrapper

import (
	"context"
	"errors"
	"github.com/classtorch/prpc/balancer"
	logger2 "github.com/classtorch/prpc/logger"
	"sync"
)

// PickerWrapper is a wrapper of balancer.Picker. It blocks on certain Pick
// actions and unblock when there's a picker update.
type PickerWrapper struct {
	mu         sync.Mutex
	blockingCh chan struct{}
	picker     balancer.Picker
	log        logger2.Log
}

// NewPickerWrapper return a pickerWrapper
func NewPickerWrapper(log logger2.Log) *PickerWrapper {
	return &PickerWrapper{
		blockingCh: make(chan struct{}),
		log:        log,
	}
}

// updatePicker update current picker
func (pw *PickerWrapper) updatePicker(p balancer.Picker) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	pw.picker = p
	close(pw.blockingCh)
	pw.blockingCh = make(chan struct{})
}

// Pick pick a available addresses
func (pw *PickerWrapper) Pick(ctx context.Context, failfast bool) (string, error) {
	var ch chan struct{}

	var lastPickErr error
	for {
		pw.mu.Lock()
		if pw.picker == nil {
			ch = pw.blockingCh
		}
		if ch == pw.blockingCh {
			pw.mu.Unlock()
			select {
			case <-ctx.Done():
				var errStr string
				if lastPickErr != nil {
					errStr = "latest balancer error: " + lastPickErr.Error()
				} else {
					errStr = ctx.Err().Error()
				}
				return "", errors.New(errStr)
			case <-ch:
			}
			continue
		}

		ch = pw.blockingCh
		p := pw.picker
		pw.mu.Unlock()

		pickResult, err := p.Pick()

		if err != nil {
			if !failfast {
				lastPickErr = err
				pw.log.Errorf("pick err:%+v,rePicking...", err)
				continue
			}
			return "", err
		}
		if pickResult.Done != nil {
			pickResult.Done(balancer.DoneInfo{})
		}
		return pickResult.Address.Addr, nil
	}
}

// close pickerWrapper
func (pw *PickerWrapper) close() {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	close(pw.blockingCh)
}
