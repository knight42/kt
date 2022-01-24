package main

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
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

func getMatchLabels(selector *metav1.LabelSelector) (map[string]string, error) {
	if selector == nil {
		return nil, fmt.Errorf("nil labelSelector")
	}
	return selector.MatchLabels, nil
}

type objectReference struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
}

func getRefObj(ref objectReference, f genericclioptions.RESTClientGetter) (runtime.Object, error) {
	gv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return nil, err
	}
	gvk := gv.WithKind(ref.Kind)
	mapper, err := f.ToRESTMapper()
	if err != nil {
		return nil, err
	}
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	result := newBuilder(f).
		ResourceNames(mapping.Resource.Resource, ref.Name).SingleResourceType().
		NamespaceParam(ref.Namespace).
		RequireObject(true).
		Latest().Do()
	if err := result.Err(); err != nil {
		return nil, err
	}
	return result.Object()
}

func getPodsSelector(obj runtime.Object, f genericclioptions.RESTClientGetter) (map[string]string, error) {
	switch t := obj.(type) {
	// Service
	case *corev1.Service:
		return t.Spec.Selector, nil

	// Deployment
	case *extensionsv1beta1.Deployment:
		return getMatchLabels(t.Spec.Selector)
	case *appsv1.Deployment:
		return getMatchLabels(t.Spec.Selector)
	case *appsv1beta1.Deployment:
		return getMatchLabels(t.Spec.Selector)
	case *appsv1beta2.Deployment:
		return getMatchLabels(t.Spec.Selector)

	// DaemonSet
	case *extensionsv1beta1.DaemonSet:
		return getMatchLabels(t.Spec.Selector)
	case *appsv1beta2.DaemonSet:
		return getMatchLabels(t.Spec.Selector)
	case *appsv1.DaemonSet:
		return getMatchLabels(t.Spec.Selector)

	// StatefulSet
	case *appsv1.StatefulSet:
		return getMatchLabels(t.Spec.Selector)
	case *appsv1beta1.StatefulSet:
		return getMatchLabels(t.Spec.Selector)
	case *appsv1beta2.StatefulSet:
		return getMatchLabels(t.Spec.Selector)

	// Job
	case *batchv1.Job:
		return getMatchLabels(t.Spec.Selector)

	// CronJob
	case *batchv1beta1.CronJob:
		return t.Spec.JobTemplate.Spec.Template.GetLabels(), nil

	// HPA
	case *autoscalingv1.HorizontalPodAutoscaler:
		ref := t.Spec.ScaleTargetRef
		refObj, err := getRefObj(objectReference{
			APIVersion: ref.APIVersion,
			Kind:       ref.Kind,
			Name:       ref.Name,
			Namespace:  t.Namespace,
		}, f)
		if err != nil {
			return nil, err
		}
		return getPodsSelector(refObj, f)
	case *autoscalingv2beta1.HorizontalPodAutoscaler:
		ref := t.Spec.ScaleTargetRef
		refObj, err := getRefObj(objectReference{
			APIVersion: ref.APIVersion,
			Kind:       ref.Kind,
			Name:       ref.Name,
			Namespace:  t.Namespace,
		}, f)
		if err != nil {
			return nil, err
		}
		return getPodsSelector(refObj, f)
	case *autoscalingv2beta2.HorizontalPodAutoscaler:
		ref := t.Spec.ScaleTargetRef
		refObj, err := getRefObj(objectReference{
			APIVersion: ref.APIVersion,
			Kind:       ref.Kind,
			Name:       ref.Name,
			Namespace:  t.Namespace,
		}, f)
		if err != nil {
			return nil, err
		}
		return getPodsSelector(refObj, f)

	// Deprecated ReplicationController
	case *corev1.ReplicationController:
		return t.Spec.Selector, nil

	// ReplicaSet
	case *appsv1.ReplicaSet:
		return getMatchLabels(t.Spec.Selector)
	case *extensionsv1beta1.ReplicaSet:
		return getMatchLabels(t.Spec.Selector)
	}
	return nil, fmt.Errorf("unknown object: %#v", obj)
}

func newBuilder(f genericclioptions.RESTClientGetter) *resource.Builder {
	return resource.NewBuilder(f).WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...)
}
