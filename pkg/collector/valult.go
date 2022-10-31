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

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsVault struct {
	metrics map[string]ConstMetricCollector
}

type Mapping struct {
	Name string `yaml:"name"`
	Help string `yaml:"help,omitempty"`

	KubeResourceMeta []string `yaml:"resource_meta"`
	KubeAnnotations  []string `yaml:"kube_labels,omitempty"`
	KubeLabels       []string `yaml:"kube_annotations,omitempty"`
	MaxRevisions     int      `yaml:"max_revisions,omitempty"`
}

type Sample struct {
	ResourceLabels      map[string]string
	ResourceAnnotations map[string]string
	ResourceMeta        []string
}

func NewVault() *MetricsVault {
	return &MetricsVault{metrics: make(map[string]ConstMetricCollector)}
}

func (v *MetricsVault) RegisterMappings(mappings []Mapping) error {
	for _, mapping := range mappings {

		collector := NewConstGaugeCollector(mapping)
		v.metrics[mapping.Name] = collector

		if err := prometheus.Register(collector); err != nil {
			return fmt.Errorf("mapping registration: %v", err)
		}
	}
	return nil
}

func (v *MetricsVault) Store(index string, sample Sample) {
	v.metrics[index].Store(sample)
}

func (v *MetricsVault) Clear(index string, sample Sample) {
	v.metrics[index].Clear(sample)
}
