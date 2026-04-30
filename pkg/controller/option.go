package controller

import (
	"regexp"

	"github.com/knight42/kt/pkg/query"
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

func WithPrefixMode(mode string) Option {
	return func(t *Controller) {
		t.prefixMode = mode
	}
}

func WithNodeName(nodeName string) Option {
	return func(t *Controller) {
		t.nodeName = nodeName
	}
}

func WithQuery(expr query.Expr) Option {
	return func(t *Controller) {
		t.queryExpr = expr
	}
}
