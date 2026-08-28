package Utilities

import (
	"sync"
	"time"
)

// ApplicationAudioMeter reports the peak render level for one Windows process
// tree. Platform-specific constructors keep Windows Core Audio out of other
// builds.
type ApplicationAudioMeter struct {
	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func newApplicationAudioMeter() *ApplicationAudioMeter {
	return &ApplicationAudioMeter{
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

func (m *ApplicationAudioMeter) Stop() {
	if m == nil {
		return
	}
	m.stopOnce.Do(func() { close(m.stopCh) })
	select {
	case <-m.doneCh:
	case <-time.After(time.Second):
	}
}
