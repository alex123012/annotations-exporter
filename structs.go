package main

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NewPod() *Pods {
	pod := &Pods{}
	pod.ResourcesInterface = pod
	pod.ResourceType = POD
	return pod
}

type Pods struct {
	Resources
}

func (m *Pods) getObjectList(clientset *kubernetes.Clientset, namespace string) (ResourceList, error) {
	result, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	resList := make(ResourceList, len(result.Items))
	if err != nil {
		return resList, err
	}
	for i := range result.Items {
		resList[i] = &result.Items[i]
	}
	return resList, nil
}

func NewDeployment() *Deployments {
	deploy := &Deployments{}
	deploy.ResourcesInterface = deploy
	deploy.ResourceType = DEPLOYMENT
	return deploy
}

type Deployments struct {
	Resources
}

func (m *Deployments) getObjectList(clientset *kubernetes.Clientset, namespace string) (ResourceList, error) {
	result, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	resList := make(ResourceList, len(result.Items))
	if err != nil {
		return resList, err
	}
	for i := range result.Items {
		resList[i] = &result.Items[i]
	}
	return resList, nil
}

func NewStatefulSets() *StatefulSets {
	sts := &StatefulSets{}
	sts.ResourcesInterface = sts
	sts.ResourceType = STATEFULSET
	return sts
}

type StatefulSets struct {
	Resources
}

func (m *StatefulSets) getObjectList(clientset *kubernetes.Clientset, namespace string) (ResourceList, error) {
	result, err := clientset.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	resList := make(ResourceList, len(result.Items))
	if err != nil {
		return resList, err
	}
	for i := range result.Items {
		resList[i] = &result.Items[i]
	}
	return resList, nil
}

func NewIngress() *Ingresses {
	ing := &Ingresses{}
	ing.ResourcesInterface = ing
	ing.ResourceType = INGRESS
	return ing
}

type Ingresses struct {
	Resources
}

func (m *Ingresses) getObjectList(clientset *kubernetes.Clientset, namespace string) (ResourceList, error) {
	result, err := clientset.NetworkingV1().Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	resList := make(ResourceList, len(result.Items))
	if err != nil {
		return resList, err
	}
	for i := range result.Items {
		resList[i] = &result.Items[i]
	}
	return resList, nil
}
