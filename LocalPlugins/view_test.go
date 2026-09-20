package LocalPlugins

import (
	"fyne.io/fyne/v2/test"
	"testing"
)

func TestObserveDoesNotStartHelper(t *testing.T) {
	test.NewTempApp(t)
	Observe(func(Snapshot) {}, func(string) {})
	state.Lock()
	defer state.Unlock()
	if state.host != nil || state.wanted {
		t.Fatal("opening controls must not silently start a helper in fully local mode")
	}
}
