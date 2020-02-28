package controller

import (
	"regexp"

	corev1 "k8s.io/api/core/v1"
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

func getRetryableContainerNames(pod *corev1.Pod) []string {
	sts := append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...)
	var ret []string
	for _, st := range sts {
		switch {
		case st.State.Running != nil:
			ret = append(ret, st.Name)
		}
	}
	return ret
}
