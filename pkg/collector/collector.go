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

package collector

import (
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	labelsSeparator   = byte(255)
	ApplicationPrefix = "annotations_exporter_"
)

type ConstMetricCollector interface {
	Describe(chan<- *prometheus.Desc)
	Collect(chan<- prometheus.Metric)
	Store(Sample)
	Clear(Sample)
}

type ResourceGaugeMetric struct {
	RevisionMetrics []RevisionGaugeMetric
}
type RevisionGaugeMetric struct {
	RevisionValue float64
	LabelValues   []string
}

type GaugeCollector struct {
	mu sync.RWMutex

	collection map[uint64]*ResourceGaugeMetric
	desc       *prometheus.Desc
	mapping    Mapping
}

func NewConstGaugeCollector(mapping Mapping) *GaugeCollector {

	resultPrometheusLabels := ConcatMultipleSlices(
		[][]string{
			formatPromethuesLabelSlice(mapping.KubeResourceMeta, ApplicationPrefix),

			formatPromethuesLabelSlice(mapping.ReferenceLabels, ApplicationPrefix+"label_"),
			formatPromethuesLabelSlice(mapping.ReferenceAnnotations, ApplicationPrefix+"annotation_"),

			formatPromethuesLabelSlice(mapping.KubeLabels, ApplicationPrefix+"label_"),
			formatPromethuesLabelSlice(mapping.KubeAnnotations, ApplicationPrefix+"annotation_"),
			{ApplicationPrefix + "revision"},
		})

	desc := prometheus.NewDesc(mapping.Name, mapping.Help, resultPrometheusLabels, nil)
	return &GaugeCollector{mapping: mapping, collection: make(map[uint64]*ResourceGaugeMetric), desc: desc}
}

func (c *GaugeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c *GaugeCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, s := range c.collection {
		for _, metric := range s.RevisionMetrics {
			if metric.LabelValues == nil {
				continue
			}
			metric, err := prometheus.NewConstMetric(c.desc, prometheus.GaugeValue, metric.RevisionValue, metric.LabelValues...)
			if err != nil {
				log.Printf("prepare gauge: %v\n", err)
				continue
			}
			ch <- metric
		}
	}
}

func (c *GaugeCollector) Store(sample Sample) {
	kubeReferenceForHash := ConcatMultipleSlices(
		[][]string{
			compareLabelsSliceWithMap(c.mapping.ReferenceLabels, sample.ResourceLabels),
			compareLabelsSliceWithMap(c.mapping.ReferenceAnnotations, sample.ResourceAnnotations),
		})

	if !c.mapping.OnlyLabelsAndAnnotations {
		kubeReferenceForHash = ConcatMultipleSlices([][]string{
			sample.ResourceMeta,
			kubeReferenceForHash,
		})
	}

	labelsHash := hashLabels(kubeReferenceForHash)

	lastRevision := 0
	newMetric := RevisionGaugeMetric{
		RevisionValue: float64(lastRevision),
		LabelValues: ConcatMultipleSlices(
			[][]string{
				kubeReferenceForHash,
				compareLabelsSliceWithMap(c.mapping.KubeLabels, sample.ResourceLabels),
				compareLabelsSliceWithMap(c.mapping.KubeAnnotations, sample.ResourceAnnotations),
				{fmt.Sprint(lastRevision)},
			}),
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	storedResourceMetrics, ok := c.collection[labelsHash]
	if !ok {
		storedResourceMetrics = &ResourceGaugeMetric{
			RevisionMetrics: make([]RevisionGaugeMetric, c.mapping.MaxRevisions),
		}
		storedResourceMetrics.RevisionMetrics[lastRevision] = newMetric
	} else {
		if reflect.DeepEqual(newMetric.LabelValues, storedResourceMetrics.RevisionMetrics[lastRevision].LabelValues) {
			return
		}
		storedResourceMetrics.RevisionMetrics = shiftMetricsSlice(storedResourceMetrics.RevisionMetrics, c.mapping.MaxRevisions)
		storedResourceMetrics.RevisionMetrics[0] = newMetric
	}
	c.collection[labelsHash] = storedResourceMetrics
}

func (c *GaugeCollector) Clear(sample Sample) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.collection, hashLabels(sample.ResourceMeta))
}
