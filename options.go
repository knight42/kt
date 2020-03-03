package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"

	"github.com/knight42/kt/pkg/controller"
)

type Options struct {
	color        string
	selector     string
	sinceSeconds time.Duration
	sinceTime    string
	previous     bool
	timestamps   bool
	exitWithPods bool
	tail         int64
	container    string

	restClientGetter genericclioptions.RESTClientGetter

	namespace string

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
		return fmt.Errorf("unknown value of flag `color`: %s", o.color)
	}

	switch len(args) {
	case 0:
		if len(o.selector) == 0 {
			return fmt.Errorf("empty selector")
		}
	case 1:
		if len(o.selector) > 0 {
			return fmt.Errorf("label selector cannot be used here")
		}
		o.podNamePattern, err = regexp.Compile(args[0])
		if err != nil {
			return err
		}
	case 2:
		if len(o.selector) > 0 {
			return fmt.Errorf("label selector cannot be used here")
		}
		result := newBuilder(getter).
			NamespaceParam(o.namespace).DefaultNamespace().
			ResourceNames(args[0], args[1]).SingleResourceType().
			RequireObject(true).
			Do()
		if err := result.Err(); err != nil {
			return err
		}

		return result.Visit(func(info *resource.Info, e error) error {
			if e != nil {
				return e
			}
			gr := formatGroupResources(info.Mapping.Resource)
			objName := info.Name
			switch gr {
			case "pods":
				o.podNamePattern = regexp.MustCompile("^" + objName + "$")
			case "batch/jobs":
				o.podNamePattern = regexp.MustCompile(fmt.Sprintf(`^%s-\w+$`, objName))
			case "batch/cronjobs":
				o.podNamePattern = regexp.MustCompile(fmt.Sprintf(`^%s-\d+-\w+$`, objName))

			default:
				selector, err := getPodsSelector(info.Object, getter)
				if err != nil {
					return err
				}
				o.selector = labels.FormatLabels(selector)
			}

			return nil
		})
	}
	return nil
}

func (o *Options) Run(cmd *cobra.Command) error {
	logsOptions, err := o.toLogsOptions()
	if err != nil {
		return err
	}
	c := controller.New(o.restClientGetter, logsOptions,
		controller.WithColor(o.color),
		controller.WithPodLabelsSelector(o.selector),
		controller.WithPodNameRegexp(o.podNamePattern),
		controller.WithContainerNameRegexp(o.containerNamePattern),
		controller.EnableExitWithPods(o.exitWithPods),
	)
	return c.Run()
}

func (o *Options) toLogsOptions() (*corev1.PodLogOptions, error) {
	opt := corev1.PodLogOptions{
		Follow:     !o.previous,
		Timestamps: o.timestamps,
		Previous:   o.previous,
	}
	if len(o.sinceTime) > 0 {
		t, err := parseRFC3339(o.sinceTime)
		if err != nil {
			return nil, err
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
	return &opt, nil
}
