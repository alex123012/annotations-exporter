package collector

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"
)

func ConcatMultipleSlices(slices [][]string) []string {
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]string, totalLen)

	i := 0
	for _, s := range slices {
		i += copy(result[i:], s)
	}

	return result
}

func compareLabelsSliceWithMap(labelsSlice []string, labelsMap map[string]string) []string {
	resourceLabels := make([]string, len(labelsSlice))
	for i, labelName := range labelsSlice {
		if value, f := labelsMap[labelName]; f {
			resourceLabels[i] = value
		}
	}
	return resourceLabels
}

func shiftMetricsSlice(metrics []RevisionGaugeMetric, metricsLen int) []RevisionGaugeMetric {

	for i := metricsLen - 1; i > 0; i-- {
		revisionValue := i - 1
		revisionMetric := metrics[revisionValue]
		if revisionMetric.LabelValues == nil {
			continue
		}
		revisionMetric.RevisionValue = float64(i)
		revisionMetric.LabelValues[len(revisionMetric.LabelValues)-1] = fmt.Sprint(i)
		metrics[i] = revisionMetric
	}
	return metrics
}

func hashLabels(labels []string) uint64 {
	// TODO(nabokihms): declare hasher once
	// TODO(nabokihms): consider better hashing
	hasher := fnv.New64a()
	var hashbuf bytes.Buffer

	for _, labelValue := range labels {
		hashbuf.WriteString(labelValue)
		hashbuf.WriteByte(labelsSeparator)
	}

	_, _ = hasher.Write(hashbuf.Bytes())
	return hasher.Sum64()
}

func formatPromethuesLabelSlice(slice []string, prefix string) []string {
	result := make([]string, len(slice))
	for i, kubeMeta := range slice {
		result[i] = formatPromethuesLabelName(prefix + kubeMeta)
	}
	return result
}

func formatPromethuesLabelName(labelName string) string {
	labelName = strings.ToLower(labelName)
	labelName = strings.ReplaceAll(labelName, "/", "_")
	labelName = strings.ReplaceAll(labelName, ".", "_")
	labelName = strings.ReplaceAll(labelName, "-", "_")
	return labelName
}
