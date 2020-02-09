package tailer

import (
	"bufio"
	"context"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	"github.com/knight42/kt/pkg/api"
	"github.com/knight42/kt/pkg/log"
)

type Task struct {
	Job    func()
	Cancel context.CancelFunc
}

type Tailer interface {
	Tail()
	RetryContainers(names []string)
	Close()
}

func New(
	ns, name string,
	ctNames map[string]struct{},
	client kubernetes.Interface,
	logsOptions *corev1.PodLogOptions,
	logCh chan<- *api.Log,
) Tailer {
	rootCtx, cancel := context.WithCancel(context.Background())
	return &tailer{
		client:      client,
		namespace:   ns,
		podName:     name,
		ctNames:     ctNames,
		logsOptions: logsOptions,
		logCh:       logCh,

		rootCtx: rootCtx,
		cancel:  cancel,
		tasks:   make(map[string]Task),
	}
}

type tailer struct {
	client      kubernetes.Interface
	namespace   string
	podName     string
	ctNames     map[string]struct{}
	logsOptions *corev1.PodLogOptions
	logCh       chan<- *api.Log

	rootCtx context.Context
	cancel  context.CancelFunc
	tasks   map[string]Task
}

func (t *tailer) Tail() {
	for ct := range t.ctNames {
		k := t.newTask(ct)
		t.tasks[ct] = k
		go k.Job()
	}
}

func (t *tailer) fetchLog(ctx context.Context, container string) error {
	opt := t.logsOptions.DeepCopy()
	opt.Container = container
	stream, err := t.client.CoreV1().Pods(t.namespace).GetLogs(t.podName, opt).Stream()
	if err != nil {
		return err
	}
	stopCh := ctx.Done()
	defer stream.Close()
	r := bufio.NewReader(stream)
	for {
		select {
		case <-stopCh:
			return nil
		default:
		}
		bytes, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		t.logCh <- &api.Log{
			Pod:       t.podName,
			Container: container,
			Content:   bytes,
		}
	}
}

func (t *tailer) RetryContainers(names []string) {
	for _, name := range names {
		tsk, ok := t.tasks[name]
		if !ok {
			continue
		}
		tsk.Cancel()
		newTask := t.newTask(name)
		t.tasks[name] = newTask
		go newTask.Job()
	}
}

func (t *tailer) Close() {
	t.cancel()
}

func (t *tailer) newTask(ct string) Task {
	ctx, cancel := context.WithCancel(t.rootCtx)
	job := func() {
		err := retry.OnError(wait.Backoff{
			Duration: time.Second * 5,
			Factor:   2.0,
			Jitter:   0.1,
			Steps:    5,
		}, errors.IsBadRequest, func() error {
			return t.fetchLog(ctx, ct)
		})
		if err != nil {
			log.V(3).Infof(">>>>> [ERROR] [%s/%s] tail: %v", t.podName, ct, err)
		}
	}
	return Task{
		Job:    job,
		Cancel: cancel,
	}
}
