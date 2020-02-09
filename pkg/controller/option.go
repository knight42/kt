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

func WithPodLabelsSelector(selector map[string]string) Option {
	return func(t *Controller) {
		t.labelSelector = selector
	}
}
