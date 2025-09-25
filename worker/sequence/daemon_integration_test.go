//go:build integration

package sequence

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/0xSplits/otelgo/recorder"
	"github.com/0xSplits/workit/handler"
	"github.com/0xSplits/workit/registry"
	"github.com/google/go-cmp/cmp"
	"github.com/xh3b4sd/logger"
)

// Test_Worker_Sequence_Integration verifies that graph executions are
// synchronized between Worker.Daemon and Worker.Ensure, so that calling
// Worker.Ensure resets the wait duration for the next tick delivered to
// Worker.Daemon.
//
// go test -tags=integration ./worker/sequence -v -race -run Test_Worker_Sequence_Integration
func Test_Worker_Sequence_Integration(t *testing.T) {
	var mut sync.Mutex
	var tim []time.Time

	var wor *Worker
	{
		wor = New(Config{
			Coo: 10 * time.Second,
			Han: [][]handler.Ensure{
				{&funcHandler{func() {
					mut.Lock()
					now := time.Now() // TODO use synctest in Go 1.25, https://go.dev/blog/testing-time
					tim = append(tim, now)
					fmt.Printf("%s\n", now)
					mut.Unlock()
				}}},
			},
			Log: logger.Fake(),
			Reg: registry.New(registry.Config{
				Env: "testing",
				Fil: isErr,
				Log: logger.Fake(),
				Met: recorder.NewMeter(recorder.MeterConfig{
					Env: "testing",
					Sco: "workit",
					Ver: "v0.1.0",
				}),
			}),
		})
	}

	// Run the daemon and record the first two ticks.

	{
		go wor.Daemon()
	}

	{
		time.Sleep(15 * time.Second)
	}

	// After 15 seconds of calling Worker.Daemon, call Worker.Ensure. This should
	// reset the internal ticker 5 seconds after the second tick, so that we then
	// record the third tick.

	{
		wor.Ensure()
	}

	// The fourth tick should be recorded 10 seconds after calling Worker.Ensure.

	{
		time.Sleep(15 * time.Second)
	}

	var act []int64
	{
		mut.Lock()
		act = tesDel(tim)
		mut.Unlock()
	}

	var exp []int64
	{
		exp = []int64{
			10, // 10 seconds between first and second tick
			5,  // 5 seconds later we call Worker.Ensure, here we reset
			10, // 10 seconds later Worker.Daemon triggers the fourth tick
		}
	}

	if dif := cmp.Diff(exp, act); dif != "" {
		t.Fatalf("-expected +actual:\n%s", dif)
	}
}

//
//
//

type funcHandler struct {
	fnc func()
}

func (h *funcHandler) Active() bool {
	return true
}

func (h *funcHandler) Ensure() error {
	{
		h.fnc()
	}

	return nil
}

//
//
//

func tesDel(tim []time.Time) []int64 {
	del := make([]int64, 0, len(tim)-1)

	for i := 1; i < len(tim); i++ {
		del = append(del, rndTo5(math.Ceil(tim[i].Sub(tim[i-1]).Seconds())))
	}

	return del
}

func rndTo5(f float64) int64 {
	return ((int64(f) + 2) / 5) * 5
}
