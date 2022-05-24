package main

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
)

type Resources struct {
	ResourcesInterface
	Items         []ResourceInfo
	ResourceType  string
	ResourcesList ResourceList
}

type ResourcesInterface interface {
	GetItems() []ResourceInfo
	GetResourceType() string
	GetObjectItems(clientset *kubernetes.Clientset, namespace string) error
	getObjectList(clientset *kubernetes.Clientset, namespace string) (ResourceList, error)
	GetObjectList(clientset *kubernetes.Clientset, namespace string) (ResourceList, error)
	CombineLabels(metric *prometheus.GaugeVec, prometheus_labels ArrayFlags) error
}

type ResourceList []ResourceInterface

type ResourceInterface interface {
	GetAnnotations() map[string]string
	GetLabels() map[string]string
	GetName() string
}
type ResourceInfo struct {
	Annotations map[string]string
	Labels      map[string]string
	Name        string
	Kind        string
	Namespace   string
}

type ArrayFlags []string

func (i *ArrayFlags) String() string {
	return "Annotations/labels to export"
}

func (i *ArrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (m *Resources) GetObjectList(clientset *kubernetes.Clientset, namespace string) (ResourceList, error) {
	result, err := m.ResourcesInterface.getObjectList(clientset, namespace)
	if err != nil {
		return ResourceList{}, err
	}
	return result, nil
}

func (m *Resources) GetObjectItems(clientset *kubernetes.Clientset, namespace string) error {

	res, err := m.GetObjectList(clientset, namespace)
	if err != nil {
		return err
	}
	for _, res := range res {
		err := m.getInfo(res, m.ResourceType, namespace)
		if err != nil {
			return err
		}
	}
	return nil
}
func (m *Resources) CombineLabels(metric *prometheus.GaugeVec, prometheus_labels ArrayFlags) error {
	for _, value := range m.Items {
		var result_labels []string
		for _, label := range prometheus_labels {
			if res, found := value.Annotations[label]; found {
				result_labels = append(result_labels, res)
			} else if res, found := value.Labels[label]; found {
				result_labels = append(result_labels, res)
			} else {
				result_labels = append(result_labels, "-")
			}
		}
		// fmt.Println(append([]string{value.Kind, value.Name, value.Namespace}, result_labels...))
		metric.WithLabelValues(append([]string{value.Kind, value.Name, value.Namespace}, result_labels...)...).SetToCurrentTime()
	}
	return nil
}
func (m *Resources) getInfo(resource ResourceInterface, kind string, namespace string) error {

	result := ResourceInfo{
		Name:        resource.GetName(),
		Namespace:   namespace,
		Kind:        kind,
		Annotations: resource.GetAnnotations(),
		Labels:      resource.GetLabels(),
	}
	m.Items = append(m.Items, result)
	// fmt.Println(result)
	return nil
}

func (m *Resources) GetItems() []ResourceInfo {
	return m.Items
}
func (m *Resources) GetResourceType() string {
	return m.ResourceType
}

func formatString(str string) string {
	str = strings.ReplaceAll(str, "/", "_")
	str = strings.ReplaceAll(str, ".", "_")
	str = strings.ReplaceAll(str, "-", "_")
	return str
}

func GetClearLabels(array []string) []string {
	var result []string
	for _, value := range array {
		result = append(result, formatString(value))
	}
	return result
}
