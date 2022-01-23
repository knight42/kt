package controller

import (
	"regexp"
)

type Option func(t *Controller)

func WithPodNameRegexp(pat *regexp.Regexp) Option {
	return func(t *Controller) {
		t.podNameRegex = pat
	}
}

func WithContainerNameRegexp(pat *regexp.Regexp) Option {
	return func(t *Controller) {
		t.containerNameRegex = pat
	}
}

func WithPodLabelsSelector(selector string) Option {
	return func(t *Controller) {
		t.labelSelector = selector
	}
}

func WithColor(when string) Option {
	return func(t *Controller) {
		t.color = when
	}
}

func EnableExitWithPods(enable bool) Option {
	return func(t *Controller) {
		t.exitWithPods = enable
	}
}

func WithPrefix(enable bool) Option {
	return func(t *Controller) {
		t.showPrefix = enable
	}
}

func WithNodeName(nodeName string) Option {
	return func(t *Controller) {
		t.nodeName = nodeName
	}
}
