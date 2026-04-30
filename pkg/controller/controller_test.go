package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/knight42/kt/pkg/api"
	"github.com/knight42/kt/pkg/tailer"
)

type fakeTailer struct {
	containerCount int
	onTail         func()
}

func (f *fakeTailer) Tail() {
	if f.onTail != nil {
		f.onTail()
	}
}
func (f *fakeTailer) RetryContainers(names []string) {}
func (f *fakeTailer) ContainerCount() int          { return f.containerCount }
func (f *fakeTailer) Close()                       {}

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

func TestOnPodAdded_PrefixStateSetBeforeTail(t *testing.T) {
	tailCalled := false
	prefixCorrectAtTailTime := false

	c := &Controller{
		prefixMode:  "auto",
		podsTailer:  make(map[types.UID]tailer.Tailer),
		logCh:       make(chan *api.Log, 1),
		logsOptions: &corev1.PodLogOptions{},
	}
	c.newTailerFn = func(ns, name string, ctNames map[string]struct{}, enableColor bool, client kubernetes.Interface, logsOptions *corev1.PodLogOptions, logCh chan<- *api.Log) tailer.Tailer {
		ft := &fakeTailer{containerCount: len(ctNames)}
		ft.onTail = func() {
			tailCalled = true
			prefixCorrectAtTailTime = !c.shouldShowPrefix()
		}
		return ft
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			UID:  "uid-1",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app"},
			},
		},
	}

	c.onPodAdded(pod)

	if !tailCalled {
		t.Fatal("Tail() was not called")
	}
	if !prefixCorrectAtTailTime {
		t.Error("prefix state must be set before Tail(); single pod with single container should hide prefix at Tail() time")
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
