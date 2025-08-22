package registry

import (
	"github.com/0xSplits/otelgo/recorder"
	"github.com/0xSplits/otelgo/registry"
	"github.com/0xSplits/workit/handler"
	"github.com/0xSplits/workit/handler/metrics"
	"github.com/0xSplits/workit/handler/proxy"
)

// New returns a metrics handler by wrapping the given implementation of
// handler.Ensure within a proxy handler. The returned metrics handler is
// configured with its own metrics registry according to the underlying
// configuration.
func (r *Registry) New(han handler.Ensure) handler.Interface {
	var pro handler.Interface
	{
		pro = proxy.New(proxy.Config{
			Han: han,
		})
	}

	var nam string
	{
		nam = handler.Name(pro.Unwrap())
	}

	cou := map[string]recorder.Interface{}

	{
		cou[metrics.MetricTotal] = recorder.NewCounter(recorder.CounterConfig{
			Des: "the total amount of worker handler executions",
			Lab: map[string][]string{
				"handler": {nam},
				"success": {"true", "false"},
			},
			Met: r.met,
			Nam: metrics.MetricTotal,
		})
	}

	gau := map[string]recorder.Interface{}

	his := map[string]recorder.Interface{}

	{
		his[metrics.MetricDuration] = recorder.NewHistogram(recorder.HistogramConfig{
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
			Met: r.met,
			Nam: metrics.MetricDuration,
		})
	}

	var reg registry.Interface
	{
		reg = registry.New(registry.Config{
			Env: r.env,
			Log: r.log,

			Cou: cou,
			Gau: gau,
			His: his,
		})
	}

	return metrics.New(metrics.Config{
		Fil: r.fil,
		Han: pro,
		Log: r.log,
		Nam: nam,
		Reg: reg,
	})
}
