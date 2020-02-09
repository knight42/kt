package controller

import (
	"regexp"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getContainerNames(pod *corev1.Pod, pat *regexp.Regexp) (names map[string]struct{}) {
	cts := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	names = make(map[string]struct{})
	for _, ct := range cts {
		if pat == nil || pat.MatchString(ct.Name) {
			names[ct.Name] = struct{}{}
		}
	}
	return names
}

func isCloseEnough(l metav1.Time, r time.Time) bool {
	return r.Sub(l.Time) < time.Second
}

func getRetryableContainerNames(pod *corev1.Pod) []string {
	sts := append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...)
	var ret []string
	now := time.Now()
	for _, st := range sts {
		s := st.State
		switch {
		case s.Running != nil:
			if isCloseEnough(s.Running.StartedAt, now) {
				ret = append(ret, st.Name)
			}
		}
	}
	return ret
}
