package zmap_task_test

import (
	"testing"

	zmap_task "github.com/WangYihang/digital-ocean-docker-executor/examples/zmap/pkg/model/task"
)

// Test ParseZMapProgress
func TestParseZMapProgress(t *testing.T) {
	testcases := []struct {
		input    string
		expected float64
	}{
		{"2024-01-31 23:05:01,0,98066,0.000024,0.000000,1,1,1,43,0,0,0,0,0,0,0,0,0,0,0,0", 0.000024},
		{"2024-01-31 23:05:05,4,425,0.937285,0.381937,1,40059,10001,9955,153,35,38,556,175,138,0,0,0,0,0,0", 0.937285},
		{"2024-01-31 15:36:49,131,0,99.666921,1.577377,0,3615812,31234,29258,57035,0,435,192548,14,1468,0,0,0,0,0,0", 99.666921},
	}
	for _, testcase := range testcases {
		progress, err := zmap_task.NewZMapProgress(testcase.input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if progress.PercentComplete != testcase.expected {
			t.Errorf("expected progress.Current to be %f, got %f", testcase.expected, progress.PercentComplete)
		}
	}
}
