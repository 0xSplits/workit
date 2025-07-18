package workit

import (
	"slices"
	"testing"
	"time"

	"github.com/0xSplits/otelgo/recorder"
	"github.com/0xSplits/workit/handler"
	"github.com/xh3b4sd/logger"
)

// Test_Engine_Daemon verifies the isolation of worker handler failure domains
// so that the failure of one worker handler does not affect the execution of
// another. For this test to be of any use, it must be executed using the -race
// flag, which we do in CI according to .github/workflows/go-build.yaml.
func Test_Engine_Daemon(t *testing.T) {
	// For this test we use signal channels that control the fake worker handlers
	// that we schedule during testing. Every test handler receives its own input
	// channel in order to simulate work. The behaviour of each test handler
	// henceforth is contigent on the respective channel buffer. E.g. handler A is
	// configured with a minimally buffered signal channel, causing handler A to
	// block indefinitely. All handlers will use the same output channel, which we
	// use to verify the actually executed engine schedule.

	var in1 chan string
	var in2 chan string
	var out chan string
	{
		in1 = make(chan string, 1)
		in2 = make(chan string, 3)
		out = make(chan string, 6)
	}

	// We create a worker instance in order to test the engine's scheduling. Note
	// that handler A defines a Cooler() of 60 minutes, which goes way beyond the
	// exxecution time of this task. In other words, after the first execution,
	// handler A blocks forever, even if it were to receive more execution tickets
	// from its signal channel.

	var eng *Engine
	{
		eng = New(Config{
			Env: "testing",
			Han: []handler.Interface{
				&testHandler{inp: in1, out: out, coo: time.Hour},
				&testHandler{inp: in2, out: out, coo: 0},
			},
			Log: logger.Fake(),
			Met: recorder.NewMeter(recorder.MeterConfig{
				Env: "testing",
				Sco: "workit",
				Ver: "v0.1.0",
			}),
		})
	}

	// Start the worker engine by calling Daemon() and wait for that goroutine to
	// register. Once the unbuffered ready channel is closed, <-eng.rdy unblocks
	// and the test execution starts.

	{
		go eng.Daemon()
	}

	{
		<-eng.rdy
	}

	// We fill up the signal channels to equal amounts in order to verify that
	// handler A is operating on a different schedule than handler B.

	go func() {
		in1 <- "one"
		in1 <- "one"
		in1 <- "one"
	}()

	go func() {
		in2 <- "two"
		in2 <- "two"
		in2 <- "two"
	}()

	// Define the expected result and collect the actual output signals generated
	// by the fake handlers. Once the two slices exp and act are of the same
	// length, we stop the test by closing the output channel. That will then
	// break the for loop below, allowing us to verify the test result.

	var exp []string
	{
		exp = []string{
			"one",

			"two",
			"two",
			"two",
		}
	}

	var act []string
	for x := range out {
		{
			act = append(act, x)
		}

		if len(act) == len(exp) {
			close(out)
		}
	}

	// Finally we either timeout or verify the result received over the output
	// channel. If all goes well, <-out should not block since the output channel
	// should have been closed once we collected the expected amount of output
	// signals. If the output channel were never to close, then the test would
	// fail in a timeout. This pattern is important for concurency tests, because
	// not guarding against silence failure may hide relevant scheduling issues.

	// Note that we are not verifying the exact order of received output signals,
	// because the nature of concurrent execution is quite probabilistic. In other
	// words, it is impossible to predict the order of code path executions within
	// goroutines. What we are verifying here is that handler A does not block the
	// execution of handler B, while handler A itself gets stuck.

	select {
	case <-out:
		{
			slices.Sort(act)
			slices.Sort(exp)
		}

		if !slices.Equal(act, exp) {
			t.Fatalf("expected %#v got %#v", exp, act)
		}
	case <-time.After(time.Second):
		t.Fatal("test timeout")
	}
}

type testHandler struct {
	coo time.Duration
	inp chan string
	out chan string
}

// Cooler simply returns the underlying cooldown duration, defining how long
// this handler ought to sleep before being scheduled again.
func (h *testHandler) Cooler() time.Duration {
	return h.coo
}

// Ensure simply takes a signal out of the underlying input channel and puts it
// back into the underlying output channel.
func (h *testHandler) Ensure() error {
	var s string
	{
		s = <-h.inp
	}

	{
		h.out <- s
	}

	return nil
}
