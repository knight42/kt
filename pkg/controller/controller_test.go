package controller

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/knight42/kt/pkg/tailer"
)

type fakeTailer struct {
	containerCount int
}

func (f *fakeTailer) Tail()                       {}
func (f *fakeTailer) TailSync()                   {}
func (f *fakeTailer) RetryContainers(names []string) {}
func (f *fakeTailer) ContainerCount() int         { return f.containerCount }
func (f *fakeTailer) Close()                      {}

var _ tailer.Tailer = (*fakeTailer)(nil)

func TestShouldShowPrefix_Always(t *testing.T) {
	c := &Controller{prefixMode: "always"}
	if !c.shouldShowPrefix() {
		t.Error("expected prefix to show in always mode")
	}
}

func TestShouldShowPrefix_Off(t *testing.T) {
	c := &Controller{prefixMode: "off"}
	if c.shouldShowPrefix() {
		t.Error("expected prefix to be hidden in off mode")
	}
}

func TestShouldShowPrefix_Auto(t *testing.T) {
	tests := map[string]struct {
		pods       map[types.UID]tailer.Tailer
		wantPrefix bool
	}{
		"no pods": {
			pods:       map[types.UID]tailer.Tailer{},
			wantPrefix: true,
		},
		"single pod single container": {
			pods: map[types.UID]tailer.Tailer{
				"uid-1": &fakeTailer{containerCount: 1},
			},
			wantPrefix: false,
		},
		"single pod multiple containers": {
			pods: map[types.UID]tailer.Tailer{
				"uid-1": &fakeTailer{containerCount: 2},
			},
			wantPrefix: true,
		},
		"multiple pods": {
			pods: map[types.UID]tailer.Tailer{
				"uid-1": &fakeTailer{containerCount: 1},
				"uid-2": &fakeTailer{containerCount: 1},
			},
			wantPrefix: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Controller{
				prefixMode: "auto",
				podsTailer: tt.pods,
			}
			c.updatePrefixState()
			got := c.shouldShowPrefix()
			if got != tt.wantPrefix {
				t.Errorf("shouldShowPrefix() = %v, want %v", got, tt.wantPrefix)
			}
		})
	}
}

func TestUpdatePrefixState_SkipsNonAuto(t *testing.T) {
	modes := map[string]struct{}{
		"always": {},
		"off":    {},
	}
	for mode := range modes {
		t.Run(mode, func(t *testing.T) {
			c := &Controller{
				prefixMode: mode,
				podsTailer: map[types.UID]tailer.Tailer{
					"uid-1": &fakeTailer{containerCount: 1},
				},
			}
			c.singlePodContainer.Store(false)
			c.updatePrefixState()
			if c.singlePodContainer.Load() {
				t.Error("updatePrefixState should not modify state for non-auto mode")
			}
		})
	}
}
