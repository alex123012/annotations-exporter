package resourcecontroller

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type MetricInterface interface {
	getClearLabels()
	RegisterMetric()
	GetLabels() []string
	GetMetric() *prometheus.GaugeVec
	ExportLabelsAndAnnotations(object *unstructured.Unstructured) []string
	ExportMetricForResource(generatedLabels []string, set float64)
	DeletetMetricForResource(generatedLabels []string)
}

// []string{},
type GaugeMetric struct {
	Metric           *prometheus.GaugeVec
	PrometheusLabels []string
	mapLabels        map[string]int
	defaultLabels    []string
}

func (m *GaugeMetric) getClearLabels() {
	for i, value := range m.PrometheusLabels {
		m.PrometheusLabels[i] = formatString(value)
	}
	klog.Infoln("Prometheus labels: ", m.PrometheusLabels)
}

func (m *GaugeMetric) RegisterMetric() {
	m.getClearLabels()
	m.Metric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "custom_app_annotations_and_labels",
		Help: "annotation and labels from kube resources",
	}, m.PrometheusLabels,
	)
	prometheus.Register(m.Metric)
}

func (m *GaugeMetric) ExportLabelsAndAnnotations(object *unstructured.Unstructured) []string {
	values := make([]string, len(m.PrometheusLabels))

	values[0] = object.GetKind()
	values[1] = object.GetName()
	values[2] = object.GetNamespace()
	values[3] = getStatusField(object)

	for label, value := range object.GetLabels() {
		if i, f := m.mapLabels[label]; f {
			values[i] = value
		}
	}
	for annotation, value := range object.GetAnnotations() {
		if i, f := m.mapLabels[annotation]; f {
			values[i] = value
		}
	}

	if len(values) != len(m.PrometheusLabels) {
		klog.Fatal("Error in labels")
	}
	return values
}

func (m *GaugeMetric) ExportMetricForResource(generatedLabels []string, set float64) {
	m.Metric.WithLabelValues(generatedLabels...).Set(set)
}

func (m *GaugeMetric) DeletetMetricForResource(generatedLabels []string) {
	m.Metric.DeleteLabelValues(generatedLabels...)
}

func (m *GaugeMetric) GetLabels() []string {
	return m.PrometheusLabels
}

func (m *GaugeMetric) GetMetric() *prometheus.GaugeVec {
	return m.Metric
}

func NewGaugeMetric(annotations, labels []string) *GaugeMetric {
	metric := &GaugeMetric{}
	metric.defaultLabels = []string{"kind", "name", "namespace", "status_phase"}
	customLabels := append(annotations, labels...)
	metric.PrometheusLabels = append(metric.defaultLabels, customLabels...)
	metric.mapLabels = make(map[string]int)
	for i, value := range metric.PrometheusLabels {
		metric.mapLabels[value] = i
	}
	metric.RegisterMetric()
	return metric
}

func formatString(str string) string {
	str = strings.ReplaceAll(str, "/", "_")
	str = strings.ReplaceAll(str, ".", "_")
	str = strings.ReplaceAll(str, "-", "_")
	return str
}

func getStatusField(object *unstructured.Unstructured) string {

	switch object.GetKind() == "Pod" {
	case true:
		status, found, err := unstructured.NestedString(object.Object, "status", "phase")
		if !found || err != nil {
			return ""
		}
		return status
	case false:
		return ""
	}
	klog.Warning("Something went wrong while getting status field")
	return ""
}
