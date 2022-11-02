// Copyright 2022.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"github.com/alex123012/annotations-exporter/pkg/collector"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ResourceToSample converts Kubernetes unstructured.Unstructured to the prometheus metric sample.
func ResourceToSample(resource *unstructured.Unstructured) collector.Sample {
	labels, annotations := resource.GetLabels(), resource.GetAnnotations()
	if labels == nil {
		labels = make(map[string]string)
	}
	if annotations == nil {
		annotations = make(map[string]string)
	}

	resourceMeta := []string{
		resource.GetAPIVersion(),
		resource.GetKind(),
		resource.GetNamespace(),
		resource.GetName(),
	}
	return collector.Sample{
		ResourceLabels:      labels,
		ResourceAnnotations: annotations,
		ResourceMeta:        resourceMeta,
	}
}

// ResourceMapping creates the mapping for the prometheus metrics vault. The order of the labels here should match the one
// from the sample converter function.
func ResourceMapping(kubeLabelNames, kubeAnnotationsNames []string, maxRevisions int, onlyLabelsAndAnnotations bool, referenceLabels, referenceAnnotations []string) collector.Mapping {
	resourceMeta := make([]string, 0)
	if !onlyLabelsAndAnnotations {
		resourceMeta = []string{"api_version", "kind", "namespace", "name"}
	}
	return collector.Mapping{
		Name: ExporterMetricName,
		Help: "Expose Kubernetes annotations and lables from kubernetes objects",

		ReferenceLabels:      referenceLabels,
		ReferenceAnnotations: referenceAnnotations,

		KubeResourceMeta: resourceMeta,
		KubeLabels:       kubeLabelNames,
		KubeAnnotations:  kubeAnnotationsNames,

		MaxRevisions:             maxRevisions,
		OnlyLabelsAndAnnotations: onlyLabelsAndAnnotations,
	}
}
