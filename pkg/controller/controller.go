package controller

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/fatih/color"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/knight42/kt/pkg/api"
	"github.com/knight42/kt/pkg/log"
	"github.com/knight42/kt/pkg/tailer"
)

type Controller struct {
	f            genericclioptions.RESTClientGetter
	kubeClient   kubernetes.Interface
	namespace    string
	color        string
	exitWithPods bool
	showPrefix   bool
	logsOptions  *corev1.PodLogOptions

	enableColor bool

	labelSelector string

	logCh chan *api.Log

	podNameRegex       *regexp.Regexp
	containerNameRegex *regexp.Regexp

	podsTailer map[types.UID]tailer.Tailer
}

func New(f genericclioptions.RESTClientGetter, logsOpts *corev1.PodLogOptions, opts ...Option) *Controller {
	c := &Controller{
		f:           f,
		logCh:       make(chan *api.Log, 1),
		logsOptions: logsOpts,
		podsTailer:  make(map[types.UID]tailer.Tailer),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *Controller) Run() error {
	var (
		err    error
		result *resource.Result
	)

	go c.consumeLog()
	defer close(c.logCh)

	switch c.color {
	case "always":
		c.enableColor = true
	case "auto":
		c.enableColor = !color.NoColor
	case "never":
		c.enableColor = false
	default:
		return fmt.Errorf("unknown value of flag `color`: %s", c.color)
	}

	byName := c.podNameRegex != nil

	c.namespace, _, err = c.f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	restConfig, err := c.f.ToRESTConfig()
	if err != nil {
		return err
	}

	c.kubeClient, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	builder := resource.NewBuilder(c.f).
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(c.namespace).DefaultNamespace().
		ResourceTypes("pods").SingleResourceType().
		RequireObject(true).
		Flatten().Latest()
	if byName {
		result = builder.SelectAllParam(true).Do()
	} else {
		// select pods using labels
		result = builder.LabelSelectorParam(c.labelSelector).Do()
	}

	if err := result.Err(); err != nil {
		return err
	}

	if c.logsOptions.Previous {
		err := result.Visit(func(info *resource.Info, err error) error {
			pod, ok := info.Object.(*corev1.Pod)
			if !ok {
				return nil
			}
			if byName && !c.podNameRegex.MatchString(pod.Name) {
				return nil
			}
			log.Errorf("+ [%s] pod added", pod.Name)
			c.onPodAdded(pod)
			c.podsTailer[pod.UID].TailSync()
			return nil
		})
		return err
	}

	watcher, err := result.Watch("")
	if err != nil {
		return err
	}

	defer watcher.Stop()
	for ev := range watcher.ResultChan() {
		pod, ok := ev.Object.(*corev1.Pod)
		if !ok {
			continue
		}
		if byName && !c.podNameRegex.MatchString(pod.Name) {
			continue
		}
		switch ev.Type {
		case watch.Added:
			log.Errorf("+ [%s] pod added", pod.Name)
			c.onPodAdded(pod)
			c.podsTailer[pod.UID].Tail()
		case watch.Modified:
			log.V(4).Infof(">>>>> [DEBUG] pod modified: %s", pod.Name)
			c.onPodModified(pod)
		case watch.Deleted:
			log.Errorf("- [%s] pod deleted", pod.Name)
			c.onPodDeleted(pod)
			if c.exitWithPods && len(c.podsTailer) == 0 {
				return nil
			}
		}
	}
	return nil
}

func (c *Controller) onPodAdded(pod *corev1.Pod) {
	names := getContainerNames(pod, c.containerNameRegex)
	log.V(4).Infof(">>>>> [DEBUG] new pod: %s, names: %v", pod.Name, names)
	if len(names) == 0 {
		return
	}
	c.podsTailer[pod.UID] = tailer.New(
		c.namespace, pod.Name,
		names,
		c.enableColor,
		c.kubeClient,
		c.logsOptions,
		c.logCh,
	)
}

func (c *Controller) onPodModified(pod *corev1.Pod) {
	t, ok := c.podsTailer[pod.UID]
	if !ok {
		return
	}
	t.RetryContainers(getRetryableContainerNames(pod))
}

func (c *Controller) onPodDeleted(pod *corev1.Pod) {
	t, ok := c.podsTailer[pod.UID]
	if !ok {
		return
	}
	t.Close()
	delete(c.podsTailer, pod.UID)
}

func (c *Controller) consumeLog() {
	w := bufio.NewWriter(os.Stdout)
	for i := range c.logCh {
		if c.showPrefix {
			if i.PodColor != nil {
				_, _ = i.PodColor.Fprint(w, i.Pod)
				_, _ = i.ContainerColor.Fprintf(w, "[%s] ", i.Container)
			} else {
				_, _ = w.WriteString(i.Pod + "[" + i.Container + "] ")
			}
		}
		_, _ = w.Write(i.Content)
		_ = w.Flush()
	}
}
