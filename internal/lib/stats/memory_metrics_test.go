package statsserver

import (
	"testing"
	"time"

	types "k8s.io/cri-api/pkg/apis/runtime/v1"

	"github.com/cri-o/cri-o/internal/config/cgmgr"
	"github.com/cri-o/cri-o/internal/oci"
)

func newTestContainer(t *testing.T) *oci.Container {
	t.Helper()

	ctr, err := oci.NewContainer("id", "name", "bundlePath", "logPath",
		nil, nil, nil,
		"", nil, nil, "", &types.ContainerMetadata{}, "sandbox",
		false, false, false, "", "dir", time.Now(), "")
	if err != nil {
		t.Fatalf("failed to build test container: %v", err)
	}

	return ctr
}

// TestGenerateContainerMemoryMetricsMemoryStatFields verifies the memory.stat
// derived reclaim/writeback metrics are emitted with the expected type and value.
func TestGenerateContainerMemoryMetricsMemoryStatFields(t *testing.T) {
	mem := &cgmgr.MemoryStats{
		FileDirty:             1,
		FileWriteback:         2,
		Pgscan:                3,
		Pgsteal:               4,
		WorkingSetRefaultAnon: 5,
		WorkingSetRefaultFile: 6,
	}

	want := map[string]struct {
		value      uint64
		metricType types.MetricType
	}{
		"container_memory_file_dirty_bytes":              {1, types.MetricType_GAUGE},
		"container_memory_file_writeback_bytes":          {2, types.MetricType_GAUGE},
		"container_memory_pgscan_total":                  {3, types.MetricType_COUNTER},
		"container_memory_pgsteal_total":                 {4, types.MetricType_COUNTER},
		"container_memory_workingset_refault_anon_total": {5, types.MetricType_COUNTER},
		"container_memory_workingset_refault_file_total": {6, types.MetricType_COUNTER},
	}

	got := map[string]*types.Metric{}
	for _, m := range generateContainerMemoryMetrics(newTestContainer(t), mem) {
		got[m.GetName()] = m
	}

	for name, exp := range want {
		m, ok := got[name]
		if !ok {
			t.Errorf("metric %q not emitted", name)

			continue
		}

		if m.GetValue().GetValue() != exp.value {
			t.Errorf("metric %q value = %d, want %d", name, m.GetValue().GetValue(), exp.value)
		}

		if m.GetMetricType() != exp.metricType {
			t.Errorf("metric %q type = %v, want %v", name, m.GetMetricType(), exp.metricType)
		}
	}
}
