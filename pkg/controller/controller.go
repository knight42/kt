package controller

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sync/atomic"

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
	"github.com/knight42/kt/pkg/query"
	"github.com/knight42/kt/pkg/tailer"
)

type Controller struct {
	f            genericclioptions.RESTClientGetter
	kubeClient   kubernetes.Interface
	namespace    string
	color        string
	nodeName     string
	prefixMode string
	logsOptions  *corev1.PodLogOptions

	enableColor        bool
	singlePodContainer atomic.Bool

	labelSelector string

	logCh chan *api.Log

	podNameRegex       *regexp.Regexp
	containerNameRegex *regexp.Regexp

	queryExpr  query.Expr
	podsTailer   map[types.UID]tailer.Tailer
	newTailerFn  func(ns, name string, ctNames map[string]struct{}, enableColor bool, client kubernetes.Interface, logsOptions *corev1.PodLogOptions, logCh chan<- *api.Log) tailer.Tailer
}

func New(f genericclioptions.RESTClientGetter, logsOpts *corev1.PodLogOptions, opts ...Option) *Controller {
	c := &Controller{
		f:           f,
		logCh:       make(chan *api.Log, 1),
		logsOptions: logsOpts,
		podsTailer:  make(map[types.UID]tailer.Tailer),
		newTailerFn: tailer.New,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *Controller) Run() error {
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

	var err error
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
		LabelSelectorParam(c.labelSelector).
		Flatten().Latest()
	if len(c.nodeName) > 0 {
		builder = builder.FieldSelectorParam("spec.nodeName=" + c.nodeName)
	} else {
		builder = builder.SelectAllParam(true)
	}

	result := builder.Do()
	if err := result.Err(); err != nil {
		return err
	}

	byName := c.podNameRegex != nil

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
		case watch.Modified:
			log.V(4).Infof(">>>>> [DEBUG] pod modified: %s", pod.Name)
			c.onPodModified(pod)
		case watch.Deleted:
			log.Errorf("- [%s] pod deleted", pod.Name)
			c.onPodDeleted(pod)
		}
	}
	return nil
}

func (c *Controller) onPodAdded(pod *corev1.Pod) {
	names := getContainerNames(pod, c.containerNameRegex)
	log.V(4).Infof(">>>>> [DEBUG] new pod: %s, names: %v", pod.Name, names)
	if len(names) == 0 {
		log.V(4).Infof(">>>>> [DEBUG] no container found for pod: %s regex: %s", pod.Name, c.containerNameRegex)
		return
	}
	t := c.newTailerFn(
		c.namespace, pod.Name,
		names,
		c.enableColor,
		c.kubeClient,
		c.logsOptions,
		c.logCh,
	)
	c.podsTailer[pod.UID] = t
	c.updatePrefixState()
	t.Tail()
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
	c.updatePrefixState()
}

func (c *Controller) updatePrefixState() {
	if c.prefixMode != "auto" {
		return
	}
	single := len(c.podsTailer) == 1
	if single {
		for _, t := range c.podsTailer {
			single = t.ContainerCount() == 1
		}
	}
	c.singlePodContainer.Store(single)
}

func (c *Controller) shouldShowPrefix() bool {
	switch c.prefixMode {
	case "always":
		return true
	case "off":
		return false
	default:
		return !c.singlePodContainer.Load()
	}
}

func (c *Controller) consumeLog() {
	w := bufio.NewWriter(os.Stdout)
	for i := range c.logCh {
		if c.queryExpr != nil && !c.queryExpr.Match(i.Content) {
			continue
		}
		if c.shouldShowPrefix() {
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
