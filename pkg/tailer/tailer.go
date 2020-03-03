package tailer

import (
	"bufio"
	"context"
	"io"

	"github.com/fatih/color"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/knight42/kt/pkg/api"
	"github.com/knight42/kt/pkg/log"
)

type Tailer interface {
	Tail()
	TailSync()
	RetryContainers(names []string)
	Close()
}

func New(
	ns, name string,
	ctNames map[string]struct{},
	enableColor bool,
	client kubernetes.Interface,
	logsOptions *corev1.PodLogOptions,
	logCh chan<- *api.Log,
) Tailer {
	rootCtx, cancel := context.WithCancel(context.Background())
	var podColor, ctColor *color.Color
	if enableColor {
		podColor, ctColor = pickColor()
	}
	return &tailer{
		client:      client,
		namespace:   ns,
		podName:     name,
		ctNames:     ctNames,
		logsOptions: logsOptions,
		logCh:       logCh,

		rootCtx:  rootCtx,
		cancel:   cancel,
		tasks:    make(map[string]*Task),
		podColor: podColor,
		ctColor:  ctColor,
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
	tasks   map[string]*Task

	podColor, ctColor *color.Color
}

func (t *tailer) Tail() {
	for ct := range t.ctNames {
		k := t.newTask(ct)
		t.tasks[ct] = k
		go k.Job()
	}
}

func (t *tailer) TailSync() {
	for ct := range t.ctNames {
		t.newTask(ct).Job()
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
			Pod:            t.podName,
			Container:      container,
			Content:        bytes,
			PodColor:       t.podColor,
			ContainerColor: t.ctColor,
		}
	}
}

func (t *tailer) RetryContainers(names []string) {
	var retried []string
	for _, name := range names {
		tsk, ok := t.tasks[name]
		if !ok {
			continue
		}
		log.V(5).Infof(">>>>> [DEBUG] [%s/%s] retrying, completed: %v", t.podName, name, tsk.Completed)
		if !tsk.Completed {
			continue
		}
		retried = append(retried, name)
		tsk.Cancel()
		newTask := t.newTask(name)
		t.tasks[name] = newTask
		go newTask.Job()
	}
	if len(retried) > 0 {
		log.V(5).Infof(">>>>> [DEBUG] modified pod: %s, retryable containers: %v", t.podName, retried)
	}
}

func (t *tailer) Close() {
	t.cancel()
}

func (t *tailer) newTask(ct string) *Task {
	ctx, cancel := context.WithCancel(t.rootCtx)
	task := &Task{
		Cancel: cancel,
	}
	task.Job = func() {
		err := t.fetchLog(ctx, ct)
		task.Completed = true
		log.V(5).Infof(">>>>> [DEBUG] [%s/%s] completed", t.podName, ct)
		if err != nil {
			log.V(3).Infof(">>>>> [ERROR] [%s/%s] tail: %v", t.podName, ct, err)
		}
	}
	return task
}
