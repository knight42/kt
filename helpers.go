package main

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func isPod(obj runtime.Object) bool {
	_, ok := obj.(*corev1.Pod)
	return ok
}

func parseRFC3339(s string) (metav1.Time, error) {
	if t, timeErr := time.Parse(time.RFC3339Nano, s); timeErr == nil {
		return metav1.Time{Time: t}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return metav1.Time{}, err
	}
	return metav1.Time{Time: t}, nil
}

func getPodsSelector(obj runtime.Object) (map[string]string, error) {
	switch t := obj.(type) {
	// Service
	case *corev1.Service:
		return t.Spec.Selector, nil

	// Deployment
	case *extensionsv1beta1.Deployment:
		return t.Spec.Selector.MatchLabels, nil
	case *appsv1.Deployment:
		return t.Spec.Selector.MatchLabels, nil
	case *appsv1beta1.Deployment:
		return t.Spec.Selector.MatchLabels, nil
	case *appsv1beta2.Deployment:
		return t.Spec.Selector.MatchLabels, nil

	// DaemonSet
	case *extensionsv1beta1.DaemonSet:
		return t.Spec.Selector.MatchLabels, nil
	case *appsv1beta2.DaemonSet:
		return t.Spec.Selector.MatchLabels, nil
	case *appsv1.DaemonSet:
		return t.Spec.Selector.MatchLabels, nil

	// StatefulSet
	case *appsv1.StatefulSet:
		return t.Spec.Selector.MatchLabels, nil
	case *appsv1beta1.StatefulSet:
		return t.Spec.Selector.MatchLabels, nil
	case *appsv1beta2.StatefulSet:
		return t.Spec.Selector.MatchLabels, nil

	// Job
	case *batchv1.Job:
		return t.Spec.Selector.MatchLabels, nil

	// CronJob
	case *batchv1beta1.CronJob:
		return t.Spec.JobTemplate.Spec.Template.GetLabels(), nil
	case *batchv2alpha1.CronJob:
		return t.Spec.JobTemplate.Spec.Template.GetLabels(), nil

	// FIXME
	case *autoscalingv1.HorizontalPodAutoscaler:

	// Deprecated ReplicationController
	case *corev1.ReplicationController:
		return t.Spec.Selector, nil

	// ReplicaSet
	case *appsv1.ReplicaSet:
		return t.Spec.Selector.MatchLabels, nil
	case *extensionsv1beta1.ReplicaSet:
		return t.Spec.Selector.MatchLabels, nil
	}
	return nil, fmt.Errorf("unknown object: %#v", obj)
}
