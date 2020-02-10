package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/knight42/kt/pkg/controller"
)

const (
	modeByNameRegex uint8 = iota
	modeByLabels
)

type Options struct {
	// TODO
	color        string
	selector     string
	sinceSeconds time.Duration
	sinceTime    string
	timestamps   bool
	tail         int64
	container    string

	restClientGetter genericclioptions.RESTClientGetter

	mode          uint8
	namespace     string
	labelSelector map[string]string

	podNamePattern       *regexp.Regexp
	containerNamePattern *regexp.Regexp
}

func (o *Options) Complete(getter genericclioptions.RESTClientGetter, args []string) error {
	o.restClientGetter = getter

	var err error
	o.namespace, _, err = getter.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	if len(o.container) > 0 {
		o.containerNamePattern, err = regexp.Compile(o.container)
		if err != nil {
			return err
		}
	}

	switch o.color {
	case "auto", "always", "never":
	default:
		return fmt.Errorf("unkown value of flag `color`: %s", o.color)
	}

	switch len(args) {
	case 0:
		if len(o.selector) == 0 {
			return fmt.Errorf("empty selector")
		}
		o.labelSelector, err = labels.ConvertSelectorToLabelsMap(o.selector)
		if err != nil {
			return err
		}
		o.mode = modeByLabels
	case 1:
		if len(o.selector) > 0 {
			return fmt.Errorf("label selector cannot be used here")
		}
		o.podNamePattern, err = regexp.Compile(args[0])
		if err != nil {
			return err
		}
		o.mode = modeByNameRegex
	case 2:
		if len(o.selector) > 0 {
			return fmt.Errorf("label selector cannot be used here")
		}
		result := resource.NewBuilder(getter).
			WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
			NamespaceParam(o.namespace).DefaultNamespace().
			Latest().
			ResourceNames(args[0], args[1]).
			SingleResourceType().
			Do()
		if err := result.Err(); err != nil {
			return err
		}
		obj, err := result.Object()
		if err != nil {
			return err
		}
		if isPod(obj) {
			name, _ := meta.NewAccessor().Name(obj)
			o.podNamePattern, _ = regexp.Compile(name)
			o.mode = modeByNameRegex
		} else {
			selector, err := getPodsSelector(obj)
			if err != nil {
				return err
			}
			o.labelSelector = selector
			o.mode = modeByLabels
		}
	}
	if o.mode == modeByLabels && o.tail == 0 {
		o.tail = 10
	}
	return nil
}

func (o *Options) Run(cmd *cobra.Command) error {
	logsOptions, err := o.toLogsOptions()
	if err != nil {
		return err
	}
	opts := []controller.Option{controller.WithColor(o.color)}
	switch o.mode {
	case modeByLabels:
		opts = append(opts, controller.WithPodLabelsSelector(o.labelSelector))
	case modeByNameRegex:
		opts = append(opts, controller.WithPodNameRegexp(o.podNamePattern))
	}
	if len(o.container) > 0 {
		opts = append(opts, controller.WithContainerNameRegexp(o.containerNamePattern))
	}
	c := controller.New(o.restClientGetter, &logsOptions, opts...)
	return c.Run()
}

func (o *Options) toLogsOptions() (corev1.PodLogOptions, error) {
	opt := corev1.PodLogOptions{
		Follow:     true,
		Timestamps: o.timestamps,
	}
	if len(o.sinceTime) > 0 {
		t, err := parseRFC3339(o.sinceTime)
		if err != nil {
			return corev1.PodLogOptions{}, err
		}
		opt.SinceTime = &t
	}
	if o.sinceSeconds > 0 {
		sec := int64(o.sinceSeconds.Round(time.Second).Seconds())
		opt.SinceSeconds = &sec
	}
	if o.tail > 0 {
		opt.TailLines = &o.tail
	}
	return opt, nil
}
