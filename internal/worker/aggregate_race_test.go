package worker

import (
	"sync"
	"testing"

	telemetrydomain "github.com/acme/plantguard/internal/telemetry/domain"
)

func TestAggregateSummariesConcurrentReadOnly(t *testing.T) {
	summaries := []telemetrydomain.Summary{
		{DeviceID: "dev-1", Metric: "pressure", SampleCount: 1, Min: 1, Max: 3, Avg: 2, Last: 2},
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			<-start
			value, ok := aggregateSummaries(summaries, "last")
			if !ok || value != 2 {
				t.Errorf("unexpected aggregate result: value=%v ok=%v", value, ok)
			}
		}()
	}
	close(start)
	wg.Wait()
}

func TestClampAggregateValue(t *testing.T) {
	if got := clampAggregateValue(2); got != 2 {
		t.Fatalf("expected value 2, got %v", got)
	}
}

func TestSanitizeAggregateValue(t *testing.T) {
	if got := sanitizeAggregateValue(3); got != 3 {
		t.Fatalf("expected value 3, got %v", got)
	}
}

func TestRoundAggregateValue(t *testing.T) {
	if got := roundAggregateValue(4); got != 4 {
		t.Fatalf("expected value 4, got %v", got)
	}
}
