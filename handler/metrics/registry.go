package metrics

import (
	"github.com/0xSplits/otelgo/recorder"
	"github.com/0xSplits/otelgo/registry"
	"github.com/xh3b4sd/logger"
	"go.opentelemetry.io/otel/metric"
)

func newRegistry(env string, log logger.Interface, met metric.Meter, nam string) registry.Interface {
	cou := map[string]recorder.Interface{}

	{
		cou[MetricTotal] = recorder.NewCounter(recorder.CounterConfig{
			Des: "the total amount of worker handler executions",
			Lab: map[string][]string{
				"handler": {nam},
				"success": {"true", "false"},
			},
			Met: met,
			Nam: MetricTotal,
		})
	}

	gau := map[string]recorder.Interface{}

	his := map[string]recorder.Interface{}

	{
		his[MetricDuration] = recorder.NewHistogram(recorder.HistogramConfig{
			Des: "the time it takes for worker handler executions to complete",
			Lab: map[string][]string{
				"handler": {nam},
				"success": {"true", "false"},
			},
			Buc: []float64{
				0.10, //  100 ms
				0.15, //  150 ms
				0.20, //  200 ms
				0.25, //  250 ms
				0.50, //  500 ms

				1.00, // 1000 ms
				1.50, // 1500 ms
				2.00, // 2000 ms
				2.50, // 2500 ms
				5.00, // 5000 ms
			},
			Met: met,
			Nam: MetricDuration,
		})
	}

	return registry.New(registry.Config{
		Env: env,
		Log: log,

		Cou: cou,
		Gau: gau,
		His: his,
	})
}
